package tracer

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type mwOptions struct {
	opNameFunc   func(r *http.Request) string
	spanObserver func(span opentracing.Span, r *http.Request)
	urlTagFunc   func(u *url.URL) string
	tracerName   string
}

// EchoMiddleware echo 的中间件
func EchoMiddleware(tracerName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if tracer := Get(tracerName); tracer != nil {
				return withTracer(tracerName, tracer, next, c)
			}
			return next(c)
		}
	}
}

func withTracer(tracerName string, tracer opentracing.Tracer, next echo.HandlerFunc, c echo.Context) error {
	r := c.Request()

	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	ctx, _ := tracer.Extract(opentracing.HTTPHeaders, carrier)
	sp := tracer.StartSpan("HTTP "+r.Method+" "+r.URL.Path, ext.RPCServerOption(ctx))
	defer sp.Finish()

	ext.HTTPMethod.Set(sp, r.Method)
	ext.HTTPUrl.Set(sp, r.URL.String())
	ext.Component.Set(sp, tracerName)

	r = r.WithContext(opentracing.ContextWithSpan(r.Context(), sp))
	c.SetRequest(r)

	if err := tracer.Inject(sp.Context(), opentracing.HTTPHeaders, carrier); err != nil {
		return fmt.Errorf("SpanContext Inject Error! %s", err.Error())
	}

	var err error
	if err = next(c); err != nil {
		sp.SetTag("error", true)
		sp.SetTag("errormsg", err.Error())
		c.Error(err)
	} else {
		sp.SetTag("error", false)
	}
	ext.HTTPStatusCode.Set(sp, uint16(c.Response().Status))
	return err
}
