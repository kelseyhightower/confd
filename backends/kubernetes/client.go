/*
Package kubernetes provides a backend for confd by synthesising a
key/value-like view on top of the kubernetes API.

Using In-Cluster

The simplest way to use this backend is to run it inside a pod in your
kubernetes cluster:

	confd --backend kubernetes --watch

In this case (with no `-node` flag given) it will look at the kubernetes
service account in the default location and use this to find and speak to the
API server.

If you don't want to look at the services in a namespace other than the current
one then set a `POD_NAMESPACE` environment variable

Using Out-of-Cluster

It is also possible to use this backend from outside of the kubernetes cluster.
The easiest way is to use "kubectl proxy" to handle the authentication for you:

	kubectl proxy &
	confd --backend kubernetes --node 127.0.0.1:8001

or you can specify the credentials with "--username"/"--password" or with a
combination of "--client-ca-keys", "--client-cert", and "--client-key"

To specify the namespace in this specify the node with a query parameter -- for example:

	kubectl proxy &
	confd --backend kubernetes --node 127.0.0.1:8001?namespace=my-team


Mapping API Object to Variables

Since confd expects a key-value store and the kubernetes API doesn't expose
this directly we have to define our own pattern of variables from the API
objects.

The only API objects (and thus initial key path) supported are endpoints.

For a service called "mysvc" it will create the following variables under
"/endpoints/mysvc":

— A "ports/$port_name" variable for each named port with the port number as the
value. Ports with numbers only names are not present.

— A set of keys under "ips" for each ready pod

	/endpoints/mysvc/ips/0: 172.17.0.6
	/endpoints/mysvc/ips/1: 172.17.07

- A set of keys under "notreadyips" for each not-ready pod

	/endpoints/mysvc/notreadyips/0: 172.17.0.5

— A set of keys under "allips" that combines ready and not-ready pods

	/endpoints/mysvc/allips/0: 172.17.0.6
	/endpoints/mysvc/allips/1: 172.17.0.7
	/endpoints/mysvc/allips/2: 172.17.0.5

A complete listing of all the variables created in this example service are

	/endpoints/mysvc/ports/http: 8080
	/endpoints/mysvc/ips/0: 172.17.0.6
	/endpoints/mysvc/ips/1: 172.17.07
	/endpoints/mysvc/allips/0: 172.17.0.6
	/endpoints/mysvc/allips/1: 172.17.0.7
	/endpoints/mysvc/allips/2: 172.17.0.5
	/endpoints/mysvc/notreadyips/0: 172.17.0.5

*/
package kubernetes

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/kelseyhightower/confd/log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Client struct {
	clientset               *kubernetes.Clientset
	endpointResourceVersion string
	endpointWatcher         cache.ListerWatcher
}

// New creates a Kubernetes backend for the givne config credentials.
//
// If all of the given values are empty/thier default value then we will
// attempt to configure from in-cluster ServiceAccount provided to k8s pods
// including targeting the current namespace.
//
func New(machines []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	namespace := "default"

	// If everything is empty, try the in cluster config
	var cfg *rest.Config
	if len(machines) == 0 && cert == "" && key == "" && caCert == "" && username == "" && password == "" {
		var err error
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + api.ServiceAccountNamespaceKey)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			namespace = string(token)
		}
	} else {
		if len(machines) != 1 {
			return nil, fmt.Errorf("kubernetes backend only supports a single node, %d given", len(machines))
		}
		// Check for `?namespace=<ns>' in the machines[0]
		url, err := url.Parse(machines[0])
		if err != nil {
			return nil, fmt.Errorf("Error parsing node URL: %s", err)
		}

		if ns, ok := url.Query()["namespace"]; ok && len(ns) >= 1 {
			namespace = ns[len(ns)-1]
		}

		// IF we are given "host:port?opts" then the bit we care about will be in
		// Path. If we are given "http://host:port?opts" then it will appear in
		// host. Handle both cases.
		var host string
		if url.Host != "" {
			host = url.Host
		} else {
			host = url.Path
		}

		cfg = &rest.Config{
			Host:     host,
			Username: username,
			Password: password,
			TLSClientConfig: rest.TLSClientConfig{
				CertFile: cert,
				KeyFile:  key,
				CAFile:   caCert,
			},
		}
	}

	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		log.Info("Changing target kubernetes namespace to %q from POD_NAMESPACE environment variable", ns)
		namespace = ns
	}

	log.Info("Using kubernetes API server at %s looking at namespace %q", cfg.Host, namespace)

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		clientset:       clientset,
		endpointWatcher: cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "endpoints", namespace, nil),
	}, nil
}

type endpointMatcher interface {
	Matches(e v1.Endpoints) bool
}

type allEndpointsMatcher struct {
}

func (allEndpointsMatcher) Matches(v1.Endpoints) bool {
	// "/endpoints" case
	return true
}

type serviceNameMatcher struct {
	ServiceName string
}

func (m serviceNameMatcher) Matches(e v1.Endpoints) bool {
	return e.Name == m.ServiceName
}

func varsFromV1Endpoint(e v1.Endpoints) map[string]string {
	vars := make(map[string]string)

	addHostVars := func(idx int, kind string, addr v1.EndpointAddress) {
		varName := fmt.Sprintf("/endpoints/%s/%s/%d", e.Name, kind, idx)
		vars[varName] = addr.IP
	}

	for _, subset := range e.Subsets {

		portPrefix := fmt.Sprintf("/endpoints/%s/port/", e.Name)
		for _, port := range subset.Ports {
			if port.Name != "" {
				vars[portPrefix+port.Name] = strconv.Itoa(int(port.Port))
			} else {
				vars[portPrefix+strconv.Itoa(int(port.Port))] = strconv.Itoa(int(port.Port))
			}
		}

		for n, node := range subset.Addresses {
			addHostVars(n, "ips", node)
			// We want allips to include ready and not-ready IPs
			addHostVars(n, "allips", node)
		}
		for n, node := range subset.NotReadyAddresses {
			addHostVars(n+len(subset.Addresses), "allips", node)
			addHostVars(n, "notreadyips", node)
		}
	}
	return vars
}

func newEndpointMatchFromKeyParts(parts []string) (endpointMatcher, bool) {
	// Types of path we might be given:
	// - /endpoints (all endpoints!)
	// - /endpoints/mysvc (this service)

	if parts[0] != "endpoints" {
		panic("parts must start with \"endpoints\"!")
	}

	if len(parts) == 1 {
		return allEndpointsMatcher{}, true
	}

	matcher := serviceNameMatcher{ServiceName: parts[1]}

	return matcher, true
}

func (k *Client) buildMatchers(keys []string) ([]endpointMatcher, error) {
	var matchers []endpointMatcher
	for _, key := range keys {
		key = strings.TrimPrefix(key, "/")
		parts := strings.Split(key, "/")

		switch parts[0] {
		case "endpoints":
			matcher, ok := newEndpointMatchFromKeyParts(parts)
			if ok {
				matchers = append(matchers, matcher)
			}
		default:
			return nil, fmt.Errorf("Unknown key type %q", parts[0])
		}
	}
	return matchers, nil
}

func (k *Client) GetValues(keys []string) (map[string]string, error) {
	log.Debug("Getting keys: %+v", keys)
	vars := make(map[string]string)

	endpointMatchers, err := k.buildMatchers(keys)
	if err != nil {
		return nil, err
	}

	if len(endpointMatchers) > 0 {
		err = k.setEndpointValues(&vars, endpointMatchers)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("Got vars %#+v", vars)
	return vars, nil
}

func (k *Client) setEndpointValues(vars *map[string]string, matchers []endpointMatcher) error {

	genericList, err := k.endpointWatcher.List(api.ListOptions{})
	if err != nil {
		return err
	}

	list, ok := genericList.(*v1.EndpointsList)
	if !ok {
		return fmt.Errorf("Expected a *v1.EndpointsList but got %T", genericList)
	}

	// Store the version so if we are in Watch mode it will restart from the same
	// place so we don't miss any changes
	k.endpointResourceVersion = list.GetResourceVersion()

	for _, ep := range list.Items {
		for _, matcher := range matchers {
			if !matcher.Matches(ep) {
				log.Debug("Endpoint %+v didn't match %#+v", ep.Name, matcher)
				continue
			}
			log.Debug("Endpoint %+v matched", ep.Name)
			for k, v := range varsFromV1Endpoint(ep) {
				(*vars)[k] = v
			}
		}
	}

	return nil
}

func (k *Client) WatchPrefix(prefix string, keys []string, lastIndex uint64, stopChan chan bool) (uint64, error) {
	endpointMatchers, err := k.buildMatchers(keys)
	if err != nil {
		return lastIndex, err
	}

	listWatcher := k.endpointWatcher

	if k.endpointResourceVersion == "" {
		// We don't yet have a resource version so this is the first time through.
		// So yes, something has changed (from nothing to whatever the current
		// state is)
		return lastIndex, nil
	}

	opts := api.ListOptions{
		ResourceVersion: k.endpointResourceVersion,
	}
	epWatcher, err := listWatcher.Watch(opts)
	if err != nil {
		return lastIndex, err
	}

	for {
		select {
		case <-stopChan:
			epWatcher.Stop()
			return lastIndex, nil
		case e := <-epWatcher.ResultChan():
			switch obj := e.Object.(type) {
			case nil:
				// Timeout or other error. Just try again from where we last were.
				epWatcher.Stop()
				epWatcher, err = listWatcher.Watch(opts)
				if err != nil {
					return lastIndex, err
				}

			case *unversioned.Status:
				// If we get anything we don't understand we should clear the
				// ResourceVersion so that we come in with a fresh one and start the
				// watch again
				k.endpointResourceVersion = ""
				if obj.Status == unversioned.StatusFailure && obj.Reason == unversioned.StatusReasonGone {
					// This happens every so often and is not an error.
					log.Info("Restarting watch after getting Gone reason: %s", obj.Message)
					return lastIndex, nil
				} else {
					return lastIndex, fmt.Errorf("Kubernetes API returned an error %#+v", e.Object)
				}
			default:
				k.endpointResourceVersion = ""
				return lastIndex, fmt.Errorf("Expected a *v1.Endpoints but got %T, %#+v", e.Object, e.Object)
			case *v1.Endpoints:

				for _, matcher := range endpointMatchers {
					if matcher.Matches(*obj) {
						k.endpointResourceVersion = opts.ResourceVersion
						epWatcher.Stop()
						return lastIndex, nil
					}
				}
			}
		}
	}
}
