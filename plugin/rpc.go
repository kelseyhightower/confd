package plugin

import (
	"net/rpc"

	"github.com/kelseyhightower/confd/confd"
)

// RPCClient is an implementation that talks over RPC
type RPCClient struct {
	client *rpc.Client
}

func (g *RPCClient) Configure(config map[string]string) error {
	args := &DatabaseConfigureArgs{
		Config: config,
	}
	var resp DatabaseConfigureResponse
	err := g.client.Call("Plugin.Configure", args, &resp)
	return err
}

func (g *RPCClient) GetValues(keys []string) (map[string]string, error) {
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

func (g *RPCClient) WatchPrefix(prefix string, keys []string, waitIndex uint64) (uint64, error) {
	args := &DatabaseWatchPrefixArgs{
		Prefix:    prefix,
		Keys:      keys,
		WaitIndex: waitIndex,
	}
	var resp DatabaseWatchPrefixResponse
	err := g.client.Call("Plugin.WatchPrefix", args, &resp)
	if err != nil {
		return resp.Index, err
	}

	return resp.Index, nil
}

// RPCServer is the RPC server that RPCClient talks to, conforming to
// the requirements of net/rpc
type RPCServer struct {
	Database confd.Database
}

type DatabaseConfigureArgs struct {
	Config map[string]string
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
}

type DatabaseWatchPrefixResponse struct {
	Index uint64
}

func (s *RPCServer) Configure(
	args *DatabaseConfigureArgs,
	resp *DatabaseConfigureResponse) error {
	err := s.Database.Configure(args.Config)
	*resp = DatabaseConfigureResponse{}
	return err
}

func (s *RPCServer) GetValues(
	args *DatabaseGetValuesArgs,
	resp *DatabaseGetValuesResponse) error {
	values, err := s.Database.GetValues(args.Keys)
	*resp = DatabaseGetValuesResponse{
		Values: values,
	}
	return err
}

func (s *RPCServer) WatchPrefix(
	args *DatabaseWatchPrefixArgs,
	resp *DatabaseWatchPrefixResponse) error {
	index, err := s.Database.WatchPrefix(args.Prefix, args.Keys, args.WaitIndex)
	*resp = DatabaseWatchPrefixResponse{
		Index: index,
	}
	return err
}
