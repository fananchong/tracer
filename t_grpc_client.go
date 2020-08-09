package tracer

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
)

// RPCClientOption
func RPCClientOption(tracerName string) grpc.DialOption {
	return grpc.WithUnaryInterceptor(gRPCUnaryClientInterceptor(tracerName))
}

func gRPCUnaryClientInterceptor(tracerName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if tracer := Get(tracerName); tracer != nil {
			var err error
			var parentCtx opentracing.SpanContext
			if parent := opentracing.SpanFromContext(ctx); parent != nil {
				parentCtx = parent.Context()
			}

			span := tracer.StartSpan(
				method,
				opentracing.ChildOf(parentCtx),
				opentracing.Tag{Key: string(ext.Component), Value: tracerName + " gRPC"},
				ext.SpanKindRPCClient,
			)
			defer span.Finish()
			ctx = injectSpanContext(ctx, tracer, span)
			span.LogFields(log.Object("gRPC request", req))
			err = invoker(ctx, method, req, resp, cc, opts...)
			if err == nil {
				span.LogFields(log.Object("gRPC response", resp))
			} else {
				ext.Error.Set(span, true)
				span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
			}
			return err
		}
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}
