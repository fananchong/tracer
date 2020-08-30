package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/fananchong/tracer"
	"github.com/fananchong/tracer/examples/proto"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedEchoServer
}

func (s *server) TestRedis(ctx context.Context, in *proto.EchoRequest) (*proto.EchoResponse, error) {
	fmt.Printf("TestRedis called with message %q\n", in.GetMessage())
	rclient := tracer.NewRedisClient(ctx, tracerName, rdb)
	rclient.Set("data", "TestRedis", 60*time.Second)
	rclient.Set("data2", "TestRedis2", 60*time.Second)
	rclient.Set("data3", "TestRedis3", 60*time.Second)
	rclient.Set("data4", "TestRedis4", 60*time.Second)
	data := rclient.Get("data").String()
	resq := fmt.Sprintf("%s %d", data, rand.Int())
	fmt.Printf("resq %s\n", resq)
	return &proto.EchoResponse{Message: resq}, nil
}

func (s *server) TestMySQL(ctx context.Context, in *proto.EchoRequest) (*proto.EchoResponse, error) {
	tracer.MySQLPingWrap(ctx, tracerName)
	resq := fmt.Sprintf("MySQL %d", rand.Int())
	fmt.Printf("resq %s\n", resq)
	return &proto.EchoResponse{Message: resq}, nil
}

const tracerName = "server3"

var rdb *redis.Client

func main() {

	// Init tracer
	if err := tracer.Enable(tracerName); err != nil {
		panic(err)
	}

	// Init Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "123456",
	})
	defer rdb.Close()

	// Init gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":9999"))
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