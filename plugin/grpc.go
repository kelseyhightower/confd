package plugin

import (
	context "golang.org/x/net/context"

	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/plugin/proto"
)

// GRPCClient is an implementation of Database that talks over gRPC
type GRPCClient struct {
	client proto.DatabaseClient
}

func (c *GRPCClient) Configure(config map[string]string) error {
	args := &proto.ConfigureRequest{
		Config: config,
	}
	_, err := c.client.Configure(context.Background(), args)
	return err
}

func (c *GRPCClient) GetValues(keys []string) (map[string]string, error) {
	args := &proto.GetValuesRequest{
		Keys: keys,
	}
	resp, err := c.client.GetValues(context.Background(), args)
	if err != nil {
		return nil, err
	}

	return resp.Values, nil
}

func (c *GRPCClient) WatchPrefix(prefix string, keys []string, waitIndex uint64) (uint64, error) {
	args := &proto.WatchPrefixRequest{
		Prefix:    prefix,
		Keys:      keys,
		WaitIndex: waitIndex,
	}
	resp, err := c.client.WatchPrefix(context.Background(), args)
	if err != nil {
		return 0, err
	}

	return resp.Index, nil
}

// GRPCServer is the GRPC server that GRPCClient talks to.
type GRPCServer struct {
	Database confd.Database
}

func (s *GRPCServer) Configure(
	ctx context.Context,
	req *proto.ConfigureRequest) (*proto.ConfigureResponse, error) {
	err := s.Database.Configure(req.Config)
	resp := proto.ConfigureResponse{}
	return &resp, err
}

func (s *GRPCServer) GetValues(
	ctx context.Context,
	req *proto.GetValuesRequest) (*proto.GetValuesResponse, error) {
	values, err := s.Database.GetValues(req.Keys)
	resp := proto.GetValuesResponse{
		Values: values,
	}
	return &resp, err
}

func (s *GRPCServer) WatchPrefix(
	ctx context.Context,
	req *proto.WatchPrefixRequest) (*proto.WatchPrefixResponse, error) {
	index, err := s.Database.WatchPrefix(req.Prefix, req.Keys, req.WaitIndex)
	resp := proto.WatchPrefixResponse{
		Index: index,
	}
	return &resp, err
}
