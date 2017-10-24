package kubernetes

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/kelseyhightower/confd/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// Resync period for the kube controller loop.
	resyncPeriod = 5 * time.Minute

	typ  = "external_type"
	http = "http"
	grpc = "grpc"
)

type listener struct {
	typ        runtime.Object
	restClient rest.Interface
}

type Client struct {
	kubeClient *kubernetes.Clientset

	types map[string]*listener

	// First Key is the typ
	// Second key is the namespace
	// Third key is the name
	// Value is the jsonified object (Either Service, Endpoints or Statefulset)
	data map[string]map[string]map[string]string

	lock      sync.Mutex
	newChange chan struct{}
	// Already registered listeners
	listeners map[string]struct{}

	lastIndex uint64
}

func New(backends []string) (*Client, error) {
	var (
		config *rest.Config
		err    error
	)
	if os.Getenv("KUBERNETES_OUTOFCLUSTER") != "" {
		// creates the out-of-cluster config: Used for dev only
		var home string
		if home = os.Getenv("HOME"); home == "" {
			return nil, errors.New("Unable to retrieve HOME env var")
		}
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
		if err != nil {
			return nil, err
		}
	} else {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// Use protobuf for communication with apiserver.
	config.ContentType = "application/vnd.kubernetes.protobuf"

	client := &Client{
		data:      map[string]map[string]map[string]string{},
		listeners: map[string]struct{}{},
		newChange: make(chan struct{}, 1),
	}

	// creates the clientset
	client.kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.types = map[string]*listener{
		"services": &listener{
			typ:        &v1.Service{},
			restClient: client.kubeClient.Core().RESTClient(),
		},
		"endpoints": &listener{
			typ:        &v1.Endpoints{},
			restClient: client.kubeClient.Core().RESTClient(),
		},
		"statefulsets": &listener{
			typ:        &v1beta1.StatefulSet{},
			restClient: client.kubeClient.AppsV1beta1().RESTClient(),
		},
	}

	return client, nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	log.Debug("GetValues called: %#v", keys)
	c.lock.Lock()
	defer c.lock.Unlock()
	result := map[string]string{}
	for _, key := range keys {
		// TODO: Refacto this ugly inner loop
		log.Debug("KEY = %s", key)

		if key == "/" {
			// Returns eveything
			for typ, typMap := range c.data {
				for ns, nsMap := range typMap {
					for el, elData := range nsMap {
						result[path.Join(typ, ns, el)] = elData
					}
				}
			}
		}

		parts := strings.Split(strings.Trim(key, "/"), "/")

		if len(parts) == 1 {
			// Return all the entries for a type
			typ := parts[0]
			typMap, ok := c.data[typ]
			if !ok {
				continue
			}
			for ns, nsMap := range typMap {
				for el, elData := range nsMap {
					result[path.Join(typ, ns, el)] = elData
				}
			}
		} else if len(parts) == 2 {
			// Return all the entries for a type + namespace
			typ := parts[0]
			typMap, ok := c.data[typ]
			if !ok {
				continue
			}
			ns := parts[1]
			nsMap, ok := typMap[ns]
			if !ok {
				continue
			}
			for el, elData := range nsMap {
				result[path.Join(typ, ns, el)] = elData
			}
		} else if len(parts) == 3 {
			// Return all the entries for a type + namespace + name
			typ := parts[0]
			typMap, ok := c.data[typ]
			if !ok {
				continue
			}
			ns := parts[1]
			nsMap, ok := typMap[ns]
			if !ok {
				continue
			}
			el := parts[2]
			elData, ok := nsMap[el]
			if ok {
				result[path.Join(typ, ns, el)] = elData
			}
		} else {
			log.Error("Invalid key: %s", key)
		}
	}
	return result, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64,
	stopChan chan bool) (uint64, error) {

	log.Debug("WatchPrefix called: %#v, %#v, %#v", prefix, keys, waitIndex)

	c.lock.Lock()
	for _, key := range keys {
		p := path.Join(prefix, key)
		if _, ok := c.listeners[p]; ok {
			// Listener already installed
			continue
		}

		s := strings.Split(strings.Trim(p, "/"), "/")

		if len(s) != 1 && len(s) != 2 {
			log.Error("Invalid watch: %s. Only /<type>/{namespace} allowed", p)
			continue
		}
		t, ok := c.types[s[0]]
		if !ok {
			log.Error("Unknown type %s", s[0])
			continue
		}
		namespace := v1.NamespaceAll
		if len(s) == 2 {
			namespace = s[1]
		}
		log.Debug("Creating controller for type %s", s[0])
		_, controller := cache.NewInformer(
			cache.NewListWatchFromClient(
				t.restClient,
				s[0],
				namespace,
				fields.Everything()),
			t.typ,
			resyncPeriod,
			cache.ResourceEventHandlerFuncs{
				AddFunc:    c.add(s[0]),
				DeleteFunc: c.remove(s[0]),
				UpdateFunc: c.update(s[0]),
			},
		)

		log.Debug("Starting controller for type %s", s[0])
		// TODO: Make the exit work
		waitChan := make(chan struct{})
		go controller.Run(waitChan)
		c.listeners[p] = struct{}{}
	}

	c.lock.Unlock()

	// Wait for a new change
	log.Debug("Waiting for change")
	<-c.newChange

	log.Debug("Triggering an update")
	c.lastIndex++
	return c.lastIndex, nil
}

func (c *Client) add(typ string) func(obj interface{}) {
	return func(obj interface{}) {
		log.Debug("add called")
		c.addObject(typ, obj)
	}
}

func (c *Client) update(typ string) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		log.Debug("update called %#v %#v", oldObj, newObj)
		c.deleteObject(typ, oldObj)
		c.addObject(typ, newObj)
	}
}

func (c *Client) remove(typ string) func(obj interface{}) {
	return func(obj interface{}) {
		log.Debug("remove called")
		c.deleteObject(typ, obj)
	}
}

func (c *Client) addObject(typ string, obj interface{}) {
	meta, ok := getMeta(obj)
	if !ok {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	t, ok := c.data[typ]
	if !ok {
		t = map[string]map[string]string{}
		c.data[typ] = t
	}
	n, ok := t[meta.Namespace]
	if !ok {
		n = map[string]string{}
		t[meta.Namespace] = n
	}
	js, err := json.Marshal(obj)
	if err != nil {
		log.Error("Unable to marshal object: %s", err)
		return
	}
	n[meta.Name] = string(js)
	select {
	case c.newChange <- struct{}{}:
		log.Debug("Change posted")
	default:
		log.Debug("Change NOT posted")
	}
}
func (c *Client) deleteObject(typ string, obj interface{}) {
	meta, ok := getMeta(obj)
	if !ok {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	t, ok := c.data[typ]
	if !ok {
		log.Error("Delete but no old data for type %s", typ)
		return
	}
	n, ok := t[meta.Namespace]
	if !ok {
		log.Error("Delete but no old data for type/namespace %s/%s", typ, meta.Namespace)
		return
	}
	delete(n, meta.Name)
	select {
	case c.newChange <- struct{}{}:
		log.Debug("Change posted")
	default:
		log.Debug("Change NOT posted")
	}
}

func getMeta(obj interface{}) (*metav1.ObjectMeta, bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return nil, false
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	f := v.FieldByName("ObjectMeta")
	if f == reflect.Zero(f.Type()) {
		return nil, false
	}
	meta, ok := f.Interface().(metav1.ObjectMeta)
	if ok {
		if meta.Namespace == "" {
			log.Error("Namespace empty for object: %#v", obj)
		}
		if meta.Name == "" {
			log.Error("Name empty for object: %#v", obj)
		}
	}
	return &meta, ok
}
