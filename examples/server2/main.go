package main

import (
	"context"
	"fmt"
	"io"
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

func (s *server) ServerStreamingEcho(in *proto.EchoRequest, stream proto.Echo_ServerStreamingEchoServer) (err error) {
	fmt.Printf("--- ServerStreamingEcho ---\n")
	fmt.Printf("request received: %v\n", in)
	// Read requests and send responses.
	for i := 0; i < 5; i++ {
		fmt.Printf("echo message %v\n", in.Message)
		err := stream.Send(&proto.EchoResponse{Message: in.Message})
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *server) ClientStreamingEcho(stream proto.Echo_ClientStreamingEchoServer) (err error) {
	fmt.Printf("--- ClientStreamingEcho ---\n")

	// Read requests and send responses.
	var message string
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("echo last received message\n")
			return stream.SendAndClose(&proto.EchoResponse{Message: message})
		}
		message = in.Message
		fmt.Printf("request received: %v, building echo\n", in)
		if err != nil {
			return err
		}
	}
}
func (s *server) BidirectionalStreamingEcho(stream proto.Echo_BidirectionalStreamingEchoServer) (err error) {
	fmt.Printf("--- BidirectionalStreamingEcho ---\n")

	// Read requests and send responses.
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("request received %v, sending echo\n", in)
		if err := stream.Send(&proto.EchoResponse{Message: in.Message}); err != nil {
			return err
		}
	}
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
		tracer.RPCUnaryServerInterceptorOption(tracerName),  // server tracer
		tracer.RPCStreamServerInterceptorOption(tracerName), // server tracer
	)
	proto.RegisterEchoServer(s, &server{})
	s.Serve(lis)
}
