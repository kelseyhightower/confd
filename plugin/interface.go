package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/plugin/proto"
	"google.golang.org/grpc"
)

// DatabasePlugin is the implementation of plugin.Plugin so we can
// serve/consume a plugin
type DatabasePlugin struct {
	plugin.NetRPCUnsupportedPlugin

	// Impl Injection
	Impl confd.Database
}

// GRPCServer returns an gRPC server for this plugin type.
func (p *DatabasePlugin) GRPCServer(s *grpc.Server) error {
	proto.RegisterDatabaseServer(s, &GRPCServer{Database: p.Impl})
	return nil
}

// GRPCClient returns an implementation of our interface that communicates
// over an gRPC client.
func (DatabasePlugin) GRPCClient(c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewDatabaseClient(c)}, nil
}
