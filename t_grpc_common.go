package tracer

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/metadata"
)

func injectSpanContext(ctx context.Context, tracer opentracing.Tracer, span opentracing.Span) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, metadataCarrier(md))
	if err != nil {
		span.LogFields(log.String("event", "Tracer.Inject() failed"), log.Error(err))
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func extractSpanContext(ctx context.Context, tracer opentracing.Tracer) (opentracing.SpanContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	return tracer.Extract(opentracing.HTTPHeaders, metadataCarrier(md))
}

// metadataCarrier

type metadataCarrier metadata.MD

func (md metadataCarrier) Set(key, val string) {
	key = strings.ToLower(key)
	md[key] = append(md[key], val)
}

func (md metadataCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range md {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
