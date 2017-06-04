package plugin

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
)

// DatabaseRPC is an implementation that talks over RPC
type DatabaseRPC struct{ client *rpc.Client }

func (g *DatabaseRPC) GetValues(keys []string) (resp map[string]string, err error) {
	err = g.client.Call("Plugin.GetValues", keys, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (g *DatabaseRPC) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (resp uint64, err error) {
	err = g.client.Call("Plugin.WatchPrefix", new(interface{}), &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// DatabaseRPCServer is the RPC server that DatabaseRPC talks to, conforming to
// the requirements of net/rpc
type DatabaseRPCServer struct {
	// This is the real implementation
	Impl confd.Database
}

func (s *DatabaseRPCServer) GetValues(keys []string, resp *map[string]string) (err error) {
	*resp, err = s.Impl.GetValues(keys)
	return err
}

func (s *DatabaseRPCServer) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool, resp *uint64) (err error) {
	*resp, err = s.Impl.WatchPrefix(prefix, keys, waitIndex, stopChan)
	return err
}

// DatabasePlugin is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a GreeterRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return GreeterRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type DatabasePlugin struct {
	// Impl Injection
	Impl confd.Database
}

func (p *DatabasePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &DatabaseRPCServer{Impl: p.Impl}, nil
}

func (DatabasePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &DatabaseRPC{client: c}, nil
}
