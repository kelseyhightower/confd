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

// Wrapper for service information
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
	sd, err := ProcessConsulData(*c, keys)
	if err != nil {
		return vars, err
	} else {
		MergeMaps(vars, sd)
	}
	return vars, nil
}

// Use the configured prefixes to load Consule data into
// the available keys
func ProcessConsulData(c Client, keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		data, err := ProcessConsulDataPrefix(c, key)
		if err != nil {
			return vars, err
		} else {
			MergeMaps(vars, data)
		}
	}
	return vars, nil
}

// Process a specific key with respect to Consul data
func ProcessConsulDataPrefix(c Client, key string) (map[string]string, error) {
	// Split the prefix into tokens so we can iterate through them
	tokens := strings.Split(strings.TrimPrefix(key, "/"), "/")
	// if the prefix doesn't start with /_consul, then this code won't
	// handle it, so exit out
	if tokens[0] != "_consul" {
		return make(map[string]string), nil
	}
	// if the prefix is simply /_consul, then get all Consul data that we can
	// and return it
	if cap(tokens) == 1 {
		return ProcessAllConsulServiceData(c)
	} else {
		// if the prefix starts with /_consul then pass it to some subfunctions
		// for processing. Pops off the _consul token so we can ignore it
		return ProcessConsulServiceData(c, tokens[1:])
	}
}

// Gets data for all services in consul by looking up services from
// the Consul catalog
func ProcessAllConsulServiceData(c Client) (map[string]string, error) {
	// Get all services from the Consul catalog
	services, err := ListServices(c)
	if err != nil {
		return nil, err
	}
	// Get all the health information for the services
	return GetServicesHealth(c, services)
}

// Gets service data from Consul by traversing the provide prefix that
// has been tokenized
func ProcessConsulServiceData(c Client, tokens []string) (map[string]string, error) {
	// only handle prefixes that start with /_consul/service
	if tokens[0] != "service" {
		return make(map[string]string), nil
	}
	if cap(tokens) == 1 {
		// If the key is /_consul/service then we get all service data
		// by querying the Consul catalog and iterating over the list of
		// services that are known
		return ProcessAllConsulServiceData(c)
	} else {
		// Otherwise, pop this token off and try processing for a specific service
		return ProcessConsulDataForService(c, tokens[1:])
	}
}

// Use the remaing tokens in the prefix to acquire service data from Consul
func ProcessConsulDataForService(c Client, tokens []string) (map[string]string, error) {
	// The first token is the service name
	service := tokens[0]

	// If there is a second token, than it is the tag
	tag := ""
	if cap(tokens) > 1 {
		tag = tokens[1]
	}
	// Get the Service from the Consul catalog
	srv, err := GetService(c, service)
	if err != nil {
		return nil, err
	}
	// If a service by this name exists, get its health, otherwise return empty
	if srv.Name != "" {
		if tag == "" {
			// If there is no tag, then get health for all tags
			return GetServiceHealth(c, srv)
		} else {
			// Otherwise just get the health for the specfic tag
			return GetServiceTagHealth(c, srv, tag)
		}
	} else {
		return make(map[string]string), nil
	}
}

// Get a specific service from the Consul catalog.
// Returns a zero struct if not found...Check that the service.Name != ""
func GetService(c Client, name string) (Service, error) {
	// Get all services in the catalog
	services, err := ListServices(c)
	// return error if it occurred
	if err != nil {
		return Service{}, err
	}
	// Find the service instance with the name we are looking for and return it
	// Otherwise return the zero struct
	var service Service
	for _, srv := range services {
		if srv.Name == name {
			service = srv
		}
	}
	return service, nil
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

// Retrieve data for all provided services
func GetServicesHealth(c Client, services []Service) (map[string]string, error) {
	entries := make(map[string]string)
	for _, service := range services {
		service_entries, err := GetServiceHealth(c, service)
		if err != nil {
			return entries, err
		}
		MergeMaps(entries, service_entries)
	}
	return entries, nil
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
func GetServiceHealth(c Client, service Service) (map[string]string, error) {
	entries := make(map[string]string)
	// For each service retrieve the list of health hosts from Consul
	serviceEntries, _, err := c.health.Service(service.Name, "", true, nil)
	if err != nil {
		return entries, err
	}
	// For each service host, add the JSON data to the root list
	for idx, serviceEntry := range serviceEntries {
		key := path.Join("/", "_consul", "service", service.Name, strconv.Itoa(idx))
		service_json, _ := json.Marshal(serviceEntry)
		entries[key] = string(service_json)
	}
	// For each tag registered for the service, retrieve the healthy hosts
	// for the service & tag
	tag_entries, err := GetServiceTagsHealth(c, service)
	if err != nil {
		return entries, err
	}
	MergeMaps(entries, tag_entries)
	return entries, nil
}

// Get service data for all Tags for a service
func GetServiceTagsHealth(c Client, service Service) (map[string]string, error) {
	entries := make(map[string]string)
	for _, tag := range service.Tags {
		tag_entries, err := GetServiceTagHealth(c, service, tag)
		if err != nil {
			return entries, err
		}
		MergeMaps(entries, tag_entries)
	}
	return entries, nil
}

// Get service data for a single service Tag
func GetServiceTagHealth(c Client, service Service, tag string) (map[string]string, error) {
	entries := make(map[string]string)
	serviceEntries, _, err := c.health.Service(service.Name, tag, true, nil)
	if err != nil {
		return entries, err
	}
	// Add each service hosts for the tag, add the JSON data to the tag list
	for idx, serviceEntry := range serviceEntries {
		key := path.Join("/", "_consul", "service", service.Name, tag, strconv.Itoa(idx))
		service_json, _ := json.Marshal(serviceEntry)
		entries[key] = string(service_json)
	}
	return entries, nil
}

// Merge two maps together. Data from the 2nd arg will override the data
// in the original map.
func MergeMaps(base map[string]string, data map[string]string) {
	for k, v := range data {
		base[k] = v
	}
}
