package tracer

import (
	"fmt"
	"io"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

type tracerWrap struct {
	tracer opentracing.Tracer
	closer io.Closer
	vaild  bool
}

var tracers sync.Map

// EnableTracer 打开 tracer
func EnableTracer(name string) (err error) {
	// 如果 tracer 已经存在，则激活，并返回
	if x, ok := tracers.Load(name); ok {
		if t := x.(*tracerWrap); !t.vaild {
			tracers.Store(name, &tracerWrap{t.tracer, t.closer, true})
		}
		return
	}
	// 如果 tracer 不存在，则创建，并返回
	cfg, err := config.FromEnv()
	cfg.ServiceName = name
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1
	if err != nil {
		fmt.Printf("cannot parse jaeger env vars: %s\n", err.Error())
		return
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		fmt.Printf("cannot initialize jaeger tracer: %s\n", err.Error())
		return
	}
	tracers.Store(name, &tracerWrap{tracer, closer, true})
	return
}

// DisableTracer 关闭 tracer
func DisableTracer(name string) {
	if x, ok := tracers.Load(name); ok {
		t := x.(*tracerWrap)
		tracers.Store(name, &tracerWrap{t.tracer, t.closer, false})
	}
}

// GetTracer 获取 tracer
func GetTracer(name string) opentracing.Tracer {
	if x, ok := tracers.Load(name); ok {
		if t := x.(*tracerWrap); t.vaild {
			return t.tracer
		}
	}
	return nil
}
