package main

// server1 HTTP 服务

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/fananchong/tracer"
	"github.com/fananchong/tracer/examples/proto"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

var err error
var echoClient proto.EchoClient

const tracerName = "server1"

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
	e.Use(middleware.Logger())
	e.Use(tracer.EchoMiddleware(tracerName))

	// Routes
	e.GET("/test1", test1)
	e.GET("/test2", test2)
	e.GET("/test3", test3)
	e.GET("/test4", test4)
	e.GET("/error1", error1)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func test1(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second) // tracer context
	defer cancel()
	res, err := echoClient.UnaryEcho(ctx, &proto.EchoRequest{Message: "hello, Unary"})
	fmt.Printf("UnaryEcho call returned %q, %v\n", res.GetMessage(), err)
	if err != nil {
		return c.String(http.StatusOK, fmt.Sprintf("error calling UnaryEcho: %v", err))
	}
	return c.String(http.StatusOK, res.GetMessage())
}

func test2(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second) // tracer context
	defer cancel()

	stream, err := echoClient.ServerStreamingEcho(ctx, &proto.EchoRequest{Message: "hello, ServerStreaming"})
	if err != nil {
		return c.String(http.StatusOK, fmt.Sprintf("error calling ServerStreamingEcho: %v", err))
	}

	// Read all the responses.
	var rpcStatus error
	result := fmt.Sprintf("response:\n")
	for {
		r, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		result += fmt.Sprintf(" - %s\n", r.Message)
	}
	if rpcStatus != io.EOF {
		fmt.Printf("failed to finish server streaming: %v", rpcStatus)
	}

	return c.String(http.StatusOK, result)
}

func test3(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second) // tracer context
	defer cancel()

	stream, err := echoClient.ClientStreamingEcho(ctx)
	if err != nil {
		return c.String(http.StatusOK, fmt.Sprintf("error calling ClientStreamingEcho: %v", err))
	}

	// Send all requests to the server.
	for i := 0; i < 5; i++ {
		if err := stream.Send(&proto.EchoRequest{Message: "hello, ClientStreaming"}); err != nil {
			fmt.Printf("failed to send streaming: %v\n", err)
		}
	}

	// Read the response.
	r, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("failed to CloseAndRecv: %v\n", err)
	}
	fmt.Printf("response:\n")
	fmt.Printf(" - %s\n\n", r.Message)

	return c.String(http.StatusOK, r.Message)
}

func test4(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second) // tracer context
	defer cancel()

	// Make RPC using the context with the metadata.
	stream, err := echoClient.BidirectionalStreamingEcho(ctx)
	if err != nil {
		fmt.Printf("failed to call BidirectionalStreamingEcho: %v\n", err)
	}

	go func() {
		// Send all requests to the server.
		for i := 0; i < 5; i++ {
			if err := stream.Send(&proto.EchoRequest{Message: "hello, BidirectionalStreaming"}); err != nil {
				fmt.Printf("failed to send streaming: %v\n", err)
			}
		}
		stream.CloseSend()
	}()

	// Read all the responses.
	var rpcStatus error
	result := fmt.Sprintf("response:\n")
	for {
		r, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		result += fmt.Sprintf(" - %s\n", r.Message)
	}
	if rpcStatus != io.EOF {
		fmt.Printf("failed to finish server streaming: %v", rpcStatus)
	}

	return c.String(http.StatusOK, result)
}

func error1(c echo.Context) (err error) {
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("%v\n%s", x, string(debug.Stack()))
		}
	}()
	panic("test panic!!!!!!! test test test")
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
