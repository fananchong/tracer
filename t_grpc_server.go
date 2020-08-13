package tracer

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// RPCUnaryServerInterceptorOption 用来设置 gRPC tracer 拦截器
func RPCUnaryServerInterceptorOption(tracerName string) grpc.ServerOption {
	return grpc.UnaryInterceptor(gRPCUnaryServerInterceptor(tracerName))
}

// RPCStreamServerInterceptorOption 用来设置 gRPC tracer 拦截器
func RPCStreamServerInterceptorOption(tracerName string) grpc.ServerOption {
	return grpc.StreamInterceptor(gRPCStreamServerInterceptor(tracerName))
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
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
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

// gRPCStreamServerInterceptor
func gRPCStreamServerInterceptor(tracerName string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if tracer := Get(tracerName); tracer != nil {
			spanContext, err := extractSpanContext(ss.Context(), tracer)
			if err != nil && err != opentracing.ErrSpanContextNotFound {
				// 如果 tracer extract 失败，那么跳过追踪
				grpclog.Errorf("SpanContext Extract Error! %s", err.Error())
				return handler(srv, ss)
			}

			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC Server"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()

			err = handler(srv, newWrappedServerStream(opentracing.ContextWithSpan(ss.Context(), span), ss))
			if err != nil {
				ext.Error.Set(span, true)
				span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
			}
			return err
		}
		return handler(srv, ss)
	}
}

// wrappedServerStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type wrappedServerStream struct {
	ctx context.Context
	grpc.ServerStream
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	return w.ServerStream.SendMsg(m)
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func newWrappedServerStream(ctx context.Context, s grpc.ServerStream) grpc.ServerStream {
	return &wrappedServerStream{ctx, s}
}
