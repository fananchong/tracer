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

	/**
	用来测试循环调用，同时，验证 server2 --rpc--> server3 --rpc--> server2 span 正常

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second) // tracer context
	defer cancel()
	res, err := echoClient.UnaryEcho(ctx, &proto.EchoRequest{Message: "hello, Unary"})
	fmt.Printf("UnaryEcho call returned %q, %v\n", res.GetMessage(), err)
	if err != nil {
		fmt.Printf("error calling UnaryEcho: %v", err)
	}

	结果类似：

	3.21s
	server1: HTTP GET /test101dbe34
	6872 Spans2062 Errors
	server1 (2)server2 (2061)server3 (4809)

	*/

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
var err error
var echoClient proto.EchoClient

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

	// 初始化一个调用 server2 的客户端（实际应该虽然不会这么去做，这里主要演练一遍，这种请下是否正常）
	var conn *grpc.ClientConn
	go func() {
		if conn, echoClient, err = newEchoClient("127.0.0.1:8888"); err != nil {
			panic(err)
		}
	}()
	defer conn.Close()

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

func newEchoClient(addr string) (conn *grpc.ClientConn, client proto.EchoClient, err error) {
	conn, err = grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		tracer.RPCUnaryClientInterceptorOption(tracerName),  // client tracer
		tracer.RPCStreamClientInterceptorOption(tracerName), // client tracer
	)
	if err == nil {
		client = proto.NewEchoClient(conn)
	} else {
		fmt.Printf("did not connect: %v", err)
	}
	return
}
