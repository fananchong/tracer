package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

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

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":8888"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("server listening at %v\n", lis.Addr())

	s := grpc.NewServer()
	proto.RegisterEchoServer(s, &server{})

	s.Serve(lis)
}
