package k8s

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/confd/log"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s/resources"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s/thirdparty"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	"k8s.io/client-go/kubernetes"
	clientapi "k8s.io/client-go/pkg/api"
	kerrors "k8s.io/client-go/pkg/api/errors"
	kapiv1 "k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/schema"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ipPool         = "/calico/v1/ipam/v4/pool"
	global         = "/calico/bgp/v1/global"
	globalASN      = "/calico/bgp/v1/global/as_num"
	globalNodeMesh = "/calico/bgp/v1/global/node_mesh"
	allNodes       = "/calico/bgp/v1/host"
	globalLogging  = "/calico/bgp/v1/global/loglevel"
)

var (
	singleNode = regexp.MustCompile("^/calico/bgp/v1/host/([a-zA-Z0-9._-]*)$")
	ipBlock    = regexp.MustCompile("^/calico/ipam/v2/host/([a-zA-Z0-9._-]*)/ipv4/block")
)

type Client struct {
	clientSet *kubernetes.Clientset
	tprClient *rest.RESTClient
}

func NewK8sClient(kubeconfig string) (*Client, error) {

	log.Debug("Building k8s client")

	// Set an explicit path to the kubeconfig if one
	// was provided.
	loadingRules := clientcmd.ClientConfigLoadingRules{}
	if kubeconfig != "" {
		log.Debug(fmt.Sprintf("Using kubeconfig: \n%s", kubeconfig))
		loadingRules.ExplicitPath = kubeconfig
	}

	// A kubeconfig file was provided.  Use it to load a config, passing through
	// any overrides.
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}

	// Create the clientset
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	log.Debug(fmt.Sprintf("Created k8s clientSet: %+v", cs))

	tprClient, err := buildTPRClient(config)
	if err != nil {
		return nil, err
	}
	kubeClient := &Client{
		clientSet: cs,
		tprClient: tprClient,
	}

	return kubeClient, nil
}

// GetValues takes the etcd like keys and route it to the appropriate k8s API endpoint.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	var kvps = make(map[string]string)
	for _, key := range keys {
		log.Debug(fmt.Sprintf("Getting key %s", key))
		if m := singleNode.FindStringSubmatch(key); m != nil {
			host := m[len(m)-1]
			kNode, err := c.clientSet.Nodes().Get(host, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			err = populateNodeDetails(kNode, kvps)
			if err != nil {
				return nil, err
			}
			// Find the podCIDR assigned to individual Nodes
		} else if m := ipBlock.FindStringSubmatch(key); m != nil {
			host := m[len(m)-1]
			kNode, err := c.clientSet.Nodes().Get(host, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			cidr := kNode.Spec.PodCIDR
			parts := strings.Split(cidr, "/")
			cidr = strings.Join(parts, "-")
			kvps[key+"/"+cidr] = "{}"
		}
		switch key {
		case global:
			// Default to "info" until this makes it into k8s.
			kvps[globalLogging] = "info"
			// Default to 64512
			kvps[globalASN] = "64512"
			// Default to true until peering info is available in k8s.
			kvps[globalNodeMesh] = `{"enabled": true}`
		case globalNodeMesh:
			// This is needed as there are calls to 'global' and directly to 'global/node_mesh'
			// Default to true until peering configuration is available in k8s.
			kvps[globalNodeMesh] = `{"enabled": true}`
		case ipPool:
			tprs := thirdparty.IpPoolList{}
			err := c.tprClient.Get().
				Resource("ippools").
				Namespace("kube-system").
				Do().Into(&tprs)

			// Ignore not found errors, as this simply means ippools does
			// not exist.
			if err != nil {
				if !kerrors.IsNotFound(err) {
					return nil, err
				}
			}

			for _, tpr := range tprs.Items {
				kvp := resources.ThirdPartyToIPPool(&tpr)
				cidr := kvp.Key.(model.IPPoolKey).CIDR

				if cidr.Version() == 4 {
					kvps[ipPool+"/"+tpr.Metadata.Name] = tpr.Spec.Value
				}
			}
		case allNodes:
			nodes, err := c.clientSet.Nodes().List(kapiv1.ListOptions{})
			if err != nil {
				return nil, err
			}

			for _, kNode := range nodes.Items {
				err := populateNodeDetails(&kNode, kvps)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	log.Debug(fmt.Sprintf("%v", kvps))
	return kvps, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {

	if waitIndex == 0 {
		switch prefix {
		case global:
			// We only have defaults for this, and they won't change in the
			// API at this time, so we can safely assume we won't be refreshing.
			time.Sleep(10 * time.Second)
			return waitIndex, nil
		case globalNodeMesh:
			// This is currently not changeable in k8s.
			time.Sleep(10 * time.Second)
			return waitIndex, nil
		case allNodes:
			// Get all nodes.  The k8s client does not expose a way to watch a single Node.
			nodes, err := c.clientSet.Nodes().List(kapiv1.ListOptions{})
			if err != nil {
				return 0, err
			}
			ver := nodes.ListMeta.ResourceVersion

			return convertResourceVersionToUint(ver, prefix)
		case ipPool:
			tprs := thirdparty.IpPoolList{}
			err := c.tprClient.Get().
				Name("ippool").
				Namespace("kube-system").
				Do().Into(&tprs)
			if err != nil {
				if !kerrors.IsNotFound(err) {
					return 0, err
				}
			}

			ver := tprs.Metadata.ResourceVersion
			return convertResourceVersionToUint(ver, prefix)
		default:
			// We aren't tracking this key, default to 10 second refresh.
			time.Sleep(60 * time.Second)
			log.Debug(fmt.Sprintf("Receieved unknown key: %v", prefix))
			return waitIndex + 1, nil
		}
	}

	switch prefix {
	case global:
		// These are currently not changeable in k8s.
		time.Sleep(10 * time.Second)
		return waitIndex, nil
	case globalNodeMesh:
		// This is currently not changeable in k8s.
		time.Sleep(10 * time.Second)
		return waitIndex, nil
	case allNodes:
		w, err := c.clientSet.Nodes().Watch(kapiv1.ListOptions{})
		if err != nil {
			return waitIndex, err
		}
		event := <-w.ResultChan()
		ver := event.Object.(*kapiv1.NodeList).ListMeta.ResourceVersion
		w.Stop()
		log.Debug(fmt.Sprintf("%d : %s", waitIndex, ver))

		return convertResourceVersionToUint(ver, prefix)
	case ipPool:
		w, err := c.tprClient.Get().
			Name("ippool").
			Namespace("kube-system").
			Watch()
		if err != nil {
			return waitIndex, err
		}
		event := <-w.ResultChan()
		ver := event.Object.(*thirdparty.IpPoolList).Metadata.ResourceVersion
		w.Stop()

		return convertResourceVersionToUint(ver, prefix)
	default:
		// We aren't tracking this key, default to 10 second refresh.
		time.Sleep(60 * time.Second)
		log.Debug(fmt.Sprintf("Receieved unknown key: %v", prefix))
		return waitIndex + 1, nil
	}
	return waitIndex, nil
}

// buildTPRClient builds a RESTClient configured to interact with Calico ThirdPartyResources.
func buildTPRClient(baseConfig *rest.Config) (*rest.RESTClient, error) {
	// Generate config using the base config.
	cfg := baseConfig
	cfg.GroupVersion = &schema.GroupVersion{
		Group:   "projectcalico.org",
		Version: "v1",
	}
	cfg.APIPath = "/apis"
	cfg.ContentType = runtime.ContentTypeJSON
	cfg.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: clientapi.Codecs}

	cli, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}

	// We also need to register resources.
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				*cfg.GroupVersion,
				&thirdparty.GlobalConfig{},
				&thirdparty.GlobalConfigList{},
				&thirdparty.IpPool{},
				&thirdparty.IpPoolList{},
				&kapiv1.ListOptions{},
				&kapiv1.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(clientapi.Scheme)

	return cli, nil
}

// populateNodeDetails populates the given kvps map with values we track from the k8s Node object.
func populateNodeDetails(kNode *kapiv1.Node, kvps map[string]string) error {
	cNode, err := resources.K8sNodeToCalico(kNode)
	if err != nil {
		log.Error("Failed to parse k8s Node into Calico Node")
		return err
	}
	node := cNode.Value.(*model.Node)
	nodeKey := allNodes + "/" + kNode.Name

	if node.FelixIPv4 != nil {
		kvps[nodeKey+"/ip_addr_v4"] = node.FelixIPv4.String()
	}
	if node.BGPIPv4Net != nil {
		kvps[nodeKey+"/network_v4"] = node.BGPIPv4Net.String()
	}

	return nil
}

// convertResourceVersionToUint converts the k8s string resource version to a uint64 expected by confd.
func convertResourceVersionToUint(rv string, prefix string) (uint64, error) {
	i, err := strconv.ParseUint(rv, 10, 64)
	if err != nil {
		log.Error(fmt.Sprintf("Could not convert '%s' resource version %s to uint64", prefix, rv))
		return 0, err
	}
	return i, nil
}
