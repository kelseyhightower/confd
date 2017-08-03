package plugin

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/plugin/proto"
	"google.golang.org/grpc"
)

// DatabasePlugin is the implementation of plugin.Plugin so we can
// serve/consume a plugin
type DatabasePlugin struct {
	// Impl Injection
	Impl confd.Database
}

// Server returns an RPC server for this plugin type.
func (p *DatabasePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Database: p.Impl}, nil
}

// Client returns an implementation of our interface that communicates
// over an RPC client.
func (DatabasePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// Server returns an gRPC server for this plugin type.
func (p *DatabasePlugin) GRPCServer(s *grpc.Server) error {
	proto.RegisterDatabaseServer(s, &GRPCServer{Database: p.Impl})
	return nil
}

// Client returns an implementation of our interface that communicates
// over an gRPC client.
func (DatabasePlugin) GRPCClient(c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewDatabaseClient(c)}, nil
}
