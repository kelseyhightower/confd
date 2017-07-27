package plugin

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
)

// DatabaseRPC is an implementation that talks over RPC
type DatabaseRPC struct {
	client *rpc.Client
}

func (g *DatabaseRPC) Configure(config map[string]interface{}) error {
	args := &DatabaseConfigureArgs{
		Config: config,
	}
	var resp DatabaseConfigureResponse
	err := g.client.Call("Plugin.Configure", args, &resp)
	return err
}

func (g *DatabaseRPC) GetValues(keys []string) (map[string]string, error) {
	args := &DatabaseGetValuesArgs{
		Keys: keys,
	}
	var resp DatabaseGetValuesResponse
	err := g.client.Call("Plugin.GetValues", args, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Values, nil
}

func (g *DatabaseRPC) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	args := &DatabaseWatchPrefixArgs{
		Prefix:    prefix,
		Keys:      keys,
		WaitIndex: waitIndex,
		StopChan:  stopChan,
	}
	var resp DatabaseWatchPrefixResponse
	err := g.client.Call("Plugin.WatchPrefix", args, &resp)
	if err != nil {
		return resp.Index, err
	}

	return resp.Index, nil
}

// DatabaseRPCServer is the RPC server that DatabaseRPC talks to, conforming to
// the requirements of net/rpc
type DatabaseRPCServer struct {
	Database confd.Database
}

type DatabaseConfigureArgs struct {
	Config map[string]interface{}
}

type DatabaseConfigureResponse struct{}

type DatabaseGetValuesArgs struct {
	Keys []string
}

type DatabaseGetValuesResponse struct {
	Values map[string]string
}

type DatabaseWatchPrefixArgs struct {
	Prefix    string
	Keys      []string
	WaitIndex uint64
	StopChan  chan bool
}

type DatabaseWatchPrefixResponse struct {
	Index uint64
}

func (s *DatabaseRPCServer) Configure(
	args *DatabaseConfigureArgs,
	resp *DatabaseConfigureResponse) error {
	err := s.Database.Configure(args.Config)
	*resp = DatabaseConfigureResponse{}
	return err
}

func (s *DatabaseRPCServer) GetValues(
	args *DatabaseGetValuesArgs,
	resp *DatabaseGetValuesResponse) error {
	values, err := s.Database.GetValues(args.Keys)
	*resp = DatabaseGetValuesResponse{
		Values: values,
	}
	return err
}

func (s *DatabaseRPCServer) WatchPrefix(
	args *DatabaseWatchPrefixArgs,
	resp *DatabaseWatchPrefixResponse) error {
	index, err := s.Database.WatchPrefix(args.Prefix, args.Keys, args.WaitIndex, args.StopChan)
	*resp = DatabaseWatchPrefixResponse{
		Index: index,
	}
	return err
}

// DatabasePlugin is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a GreeterRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return GreeterRPC for this.
type DatabasePlugin struct {
	// Impl Injection
	Impl confd.Database
}

func (p *DatabasePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &DatabaseRPCServer{Database: p.Impl}, nil
}

func (DatabasePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &DatabaseRPC{client: c}, nil
}
