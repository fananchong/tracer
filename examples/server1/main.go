package main

// server1 HTTP 服务

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fananchong/tracer"
	"github.com/fananchong/tracer/examples/proto"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

var err error
var echoClient proto.EchoClient

const tracerName = "test"

func main() {

	// Init tracer
	if err = tracer.Enable(tracerName); err != nil {
		panic(err)
	}

	// Init gRPC client
	var conn *grpc.ClientConn
	if conn, echoClient, err = newEchoClient("127.0.0.1:8888"); err != nil {
		panic(err)
	}
	defer conn.Close()

	// Echo instance
	e := echo.New()

	// Use Middleware
	e.Use(tracer.EchoMiddleware(tracerName))

	// Routes
	e.GET("/test1", test1)
	e.GET("/error1", error1)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func test1(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := echoClient.UnaryEcho(ctx, &proto.EchoRequest{Message: "hello, xxxx"})
	fmt.Printf("UnaryEcho call returned %q, %v\n", res.GetMessage(), err)
	if err != nil {
		return c.String(http.StatusOK, fmt.Sprintf("error calling UnaryEcho: %v", err))
	}
	return c.String(http.StatusOK, res.GetMessage())
}

func error1(c echo.Context) (err error) {
	defer func() {
		if x := recover(); x != nil {
			switch x.(type) {
			case error:
				err = x.(error)
			default:
				err = fmt.Errorf("%v", x)
			}
		}
	}()
	panic("test panic!!!!!!! test test test")
}

func newEchoClient(addr string) (conn *grpc.ClientConn, client proto.EchoClient, err error) {
	if conn, err = grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock()); err == nil {
		client = proto.NewEchoClient(conn)
	} else {
		log.Fatalf("did not connect: %v", err)
	}
	return
}
