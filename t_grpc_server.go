package tracer

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// RPCServerOption
func RPCServerOption(tracerName string) grpc.ServerOption {
	return grpc.UnaryInterceptor(gRPCUnaryServerInterceptor(tracerName))
}

// gRPCUnaryServerInterceptor gRPC 服务器端，一元拦截器
func gRPCUnaryServerInterceptor(tracerName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if tracer := Get(tracerName); tracer != nil {
			spanContext, err := extractSpanContext(ctx, tracer)
			if err != nil && err != opentracing.ErrSpanContextNotFound {
				// 如果 tracer extract 失败，那么跳过追踪
				grpclog.Errorf("SpanContext Extract Error! %s", err.Error())
				return handler(ctx, req)
			}

			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: tracerName + " gRPC"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()

			ctx = opentracing.ContextWithSpan(ctx, span)
			span.LogFields(log.Object("gRPC request", req))
			resp, err = handler(ctx, req)
			if err == nil {
				span.LogFields(log.Object("gRPC response", resp))
			} else {
				ext.Error.Set(span, true)
				span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
			}
			return resp, err
		}
		return handler(ctx, req)
	}
}
