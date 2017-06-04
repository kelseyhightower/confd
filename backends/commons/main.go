package commons

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
)

// Here is an implementation that talks over RPC
type StoreClientRPC struct{ client *rpc.Client }

func (g *StoreClientRPC) GetValues(keys []string) (resp map[string]string, err error) {
	err = g.client.Call("Plugin.GetValues", keys, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (g *StoreClientRPC) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (resp uint64, err error) {
	err = g.client.Call("Plugin.WatchPrefix", new(interface{}), &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

// Here is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type StoreClientRPCServer struct {
	// This is the real implementation
	Impl StoreClient
}

func (s *StoreClientRPCServer) GetValues(keys []string, resp *map[string]string) (err error) {
	*resp, err = s.Impl.GetValues(keys)
	return err
}

func (s *StoreClientRPCServer) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool, resp *uint64) (err error) {
	*resp, err = s.Impl.WatchPrefix(prefix, keys, waitIndex, stopChan)
	return err
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a GreeterRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return GreeterRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type StoreClientPlugin struct {
	// Impl Injection
	Impl StoreClient
}

func (p *StoreClientPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &StoreClientRPCServer{Impl: p.Impl}, nil
}

func (StoreClientPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &StoreClientRPC{client: c}, nil
}
