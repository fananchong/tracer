package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"github.com/fananchong/tracer"
	"github.com/fananchong/tracer/examples/proto"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedEchoServer
}

func (s *server) UnaryEcho(ctx context.Context, in *proto.EchoRequest) (*proto.EchoResponse, error) {
	fmt.Printf("UnaryEcho called with message %q\n", in.GetMessage())
	return &proto.EchoResponse{Message: fmt.Sprintf("%s %d", in.Message, rand.Int())}, nil
}

func (s *server) ServerStreamingEcho(*proto.EchoRequest, proto.Echo_ServerStreamingEchoServer) (err error) {
	return
}
func (s *server) ClientStreamingEcho(proto.Echo_ClientStreamingEchoServer) (err error) {
	return
}
func (s *server) BidirectionalStreamingEcho(proto.Echo_BidirectionalStreamingEchoServer) (err error) {
	return
}

const tracerName = "test"

func main() {

	// Init tracer
	if err := tracer.Enable(tracerName); err != nil {
		panic(err)
	}

	// Init gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":8888"))
	if err != nil {
		panic(fmt.Errorf("failed to listen: %v", err))
	}
	fmt.Printf("server listening at %v\n", lis.Addr())
	s := grpc.NewServer(
		tracer.RPCServerOption(tracerName), // server tracer
	)
	proto.RegisterEchoServer(s, &server{})
	s.Serve(lis)
}
