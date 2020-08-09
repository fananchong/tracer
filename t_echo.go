package tracer

import (
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// EchoMiddleware echo 的中间件
func EchoMiddleware(tracerName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if tracer := Get(tracerName); tracer != nil {
				r := c.Request()

				carrier := opentracing.HTTPHeadersCarrier(r.Header)
				spanContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
				if err != nil && err != opentracing.ErrSpanContextNotFound {
					// 如果 tracer extract 失败，那么跳过追踪
					c.Logger().Errorf("SpanContext Extract Error! %s", err.Error())
					return next(c)
				}
				span := tracer.StartSpan(
					"HTTP "+r.Method+" "+r.URL.Path,
					ext.RPCServerOption(spanContext),
					opentracing.Tag{Key: string(ext.Component), Value: tracerName + " HTTP"},
				)
				defer span.Finish()

				ext.HTTPMethod.Set(span, r.Method)
				ext.HTTPUrl.Set(span, r.URL.String())

				r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))
				c.SetRequest(r)

				if err = tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
					span.LogFields(log.String("event", "Tracer.Inject() failed"), log.Error(err))
				}

				if err = next(c); err != nil {
					ext.Error.Set(span, true)
					span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
					c.Error(err)
				}
				ext.HTTPStatusCode.Set(span, uint16(c.Response().Status))
				return err
			}
			return next(c)
		}
	}
}
