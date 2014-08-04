package consul

import (
	"encoding/json"
	"path"
	"strconv"
	"strings"

	"github.com/armon/consul-api"
)

// Client provides a wrapper around the consulkv client
type Client struct {
	client  *consulapi.KV
	health  *consulapi.Health
	catalog *consulapi.Catalog
}

type Service struct {
	Name string
	Tags []string
}

// NewConsulClient returns a new client to Consul for the given address
func NewConsulClient(nodes []string) (*Client, error) {
	conf := consulapi.DefaultConfig()
	if len(nodes) > 0 {
		conf.Address = nodes[0]
	}
	client, err := consulapi.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &Client{client.KV(), client.Health(), client.Catalog()}, nil
}

// GetValues queries Consul for keys
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		key := strings.TrimPrefix(key, "/")
		pairs, _, err := c.client.List(key, nil)
		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[path.Join("/", p.Key)] = string(p.Value)
		}
	}
	services, _ := ListServices(*c)
	service_vars, _ := GetServicesHealth(*c, services)
	for k, v := range service_vars {
		vars[k] = v
	}
	return vars, nil
}

// ListServices queries Consul for all registered services
func ListServices(c Client) ([]Service, error) {
	services := make([]Service, 0)
	srvs, _, err := c.catalog.Services(nil)
	if err != nil {
		return services, err
	}
	for name, tags := range srvs {
		services = append(services, Service{name, tags})
	}
	return services, nil
}

// GetServicesHealth queries Consul for the active host data for each
// service in the input. Typically this is returned value from ListServices.
// Currently Consul is queried only for healthy hosts for the service.
//
// Returned data from Consul is stored as a JSON blob under the
// /_consul/service/<name>/<idx> where <name> is the configured
// name for the service and <idx> is the entry index in result set from Consul.
// The /_consul prefix was selected to avoid potential clashes with the
// key-value store.
//
// Additionally, Consul service tag data is evaluated and the service JSON is
// stored under a second key with the format /_consul/service/<name>/<tag>/<idx>
// where <name> and <idx> are the same as above and <tag> is the applied tag
// from Consul. This is down to allow templating on either the service or a
// subset of instances of that service with a particular tag.
//
// Currently Consul datacenters are not supported via this method.
func GetServicesHealth(c Client, services []Service) (map[string]string, error) {
	entries := make(map[string]string)
	for _, entry := range services {
		// For each service retrieve the list of health hosts from Consul
		serviceEntries, _, err := c.health.Service(entry.Name, "", true, nil)
		if err != nil {
			return entries, err
		}
		// For each service host, add the JSON data to the root list
		for idx, serviceEntry := range serviceEntries {
			key := path.Join("/", "_consul", "service", entry.Name, strconv.Itoa(idx))
			service_json, _ := json.Marshal(serviceEntry)
			entries[key] = string(service_json)
		}
		// For each tag registered for the service, retrieve the healthy hosts
		// for the service & tag
		for _, tag := range entry.Tags {
			serviceEntries, _, err := c.health.Service(entry.Name, tag, true, nil)
			if err != nil {
				return entries, err
			}
			// Add each service hosts for the tag, add the JSON data to the tag list
			for idx, serviceEntry := range serviceEntries {
				key := path.Join("/", "_consul", "service", entry.Name, tag, strconv.Itoa(idx))
				service_json, _ := json.Marshal(serviceEntry)
				entries[key] = string(service_json)
			}
		}
	}
	return entries, nil
}
