package plugin

import (
	"io"

	"github.com/kelseyhightower/confd/confd"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/plugin/proto"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/transport"
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

func (c *GRPCClient) WatchPrefix(prefix string, keys []string, results chan string) error {
	args := &proto.WatchPrefixRequest{
		Prefix: prefix,
		Keys:   keys,
	}
	s, err := c.client.WatchPrefix(context.Background(), args)
	if err != nil {
		return err
	}
	for {
		_, err := s.Recv()
		if err == io.EOF {
			log.Debug("caught EOF on client")
			return nil
		}
		if err != nil {
			status, _ := status.FromError(err)
			if status.Message() == grpc.ErrClientConnClosing.Error() || status.Message() == transport.ErrConnClosing.Desc {
				return nil
			}
			return err
		}
		results <- ""
	}
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
	req *proto.WatchPrefixRequest,
	stream proto.Database_WatchPrefixServer) error {
	results := make(chan string)
	go s.Database.WatchPrefix(req.Prefix, req.Keys, results)
	go func() {
		ctx := stream.Context()
		<-ctx.Done()
		close(results)
	}()
	for range results {
		err := stream.Send(&proto.WatchPrefixResponse{})
		if err != nil {
			return err
		}
	}
	return nil
}
