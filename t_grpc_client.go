package tracer

import (
	"context"
	"io"
	"runtime"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// RPCUnaryClientInterceptorOption 用来设置 gRPC tracer 拦截器
func RPCUnaryClientInterceptorOption(tracerName string) grpc.DialOption {
	return grpc.WithUnaryInterceptor(gRPCUnaryClientInterceptor(tracerName))
}

// RPCStreamClientInterceptorOption 用来设置 gRPC tracer 拦截器
func RPCStreamClientInterceptorOption(tracerName string) grpc.DialOption {
	return grpc.WithStreamInterceptor(gRPCStreamClientInterceptor(tracerName))
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
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
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

func gRPCStreamClientInterceptor(tracerName string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if tracer := Get(tracerName); tracer != nil {
			var err error
			var parentCtx opentracing.SpanContext
			if parent := opentracing.SpanFromContext(ctx); parent != nil {
				parentCtx = parent.Context()
			}

			span := tracer.StartSpan(
				method,
				opentracing.ChildOf(parentCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCClient,
			)
			ctx = injectSpanContext(ctx, tracer, span)
			w, err := streamer(ctx, desc, cc, method, opts...)
			if err != nil {
				span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
				ext.Error.Set(span, true)
				span.Finish()
				return w, err
			}
			return createClientStream(w, method, desc, span), nil
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func createClientStream(w grpc.ClientStream, method string, desc *grpc.StreamDesc, span opentracing.Span) grpc.ClientStream {
	otcs := newWrappedClientStream(w, desc, span)

	go func() {
		select {
		case <-otcs.finishChan:
			// The client span is being finished by another code path; hence, no
			// action is necessary.
		case <-w.Context().Done():
			otcs.finish(w.Context().Err())
		}
	}()

	// The `ClientStream` interface allows one to omit calling `Recv` if it's
	// known that the result will be `io.EOF`. See
	// http://stackoverflow.com/q/42915337
	// In such cases, there's nothing that triggers the span to finish. We,
	// therefore, set a finalizer so that the span and the context goroutine will
	// at least be cleaned up when the garbage collector is run.
	runtime.SetFinalizer(otcs, func(otcs *wrappedClientStream) {
		otcs.finish(nil)
	})
	return otcs
}

type wrappedClientStream struct {
	grpc.ClientStream
	desc       *grpc.StreamDesc
	span       opentracing.Span
	once       sync.Once
	finishChan chan struct{}
}

func newWrappedClientStream(w grpc.ClientStream, desc *grpc.StreamDesc, span opentracing.Span) *wrappedClientStream {
	return &wrappedClientStream{
		ClientStream: w,
		desc:         desc,
		span:         span,
		finishChan:   make(chan struct{}),
	}
}

func (w *wrappedClientStream) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	if err != nil {
		w.finish(err)
	}
	return md, err
}

func (w *wrappedClientStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)
	if err != nil {
		w.finish(err)
	}
	return err
}

func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if err == io.EOF {
		w.finish(nil)
		return err
	} else if err != nil {
		w.finish(err)
		return err
	}
	if !w.desc.ServerStreams {
		w.finish(nil)
	}
	return err
}

func (w *wrappedClientStream) CloseSend() error {
	err := w.ClientStream.CloseSend()
	if err != nil {
		w.finish(err)
	}
	return err
}

func (w *wrappedClientStream) finish(err error) {
	w.once.Do(func() {
		close(w.finishChan)
		defer w.span.Finish()
		if err != nil {
			w.span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
			ext.Error.Set(w.span, true)
		}
	})
}
