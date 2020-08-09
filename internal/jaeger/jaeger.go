package jaeger

import (
	"fmt"
	"io"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// Jaeger 管理 jaeger tracer 实例
type Jaeger struct {
	tracers sync.Map
}

func New() *Jaeger {
	return &Jaeger{}
}

// Enable 打开 tracer
func (j *Jaeger) Enable(name string) (err error) {
	// 如果 tracer 已经存在，则激活，并返回
	if x, ok := j.tracers.Load(name); ok {
		if t := x.(*tracerWrap); !t.vaild {
			j.tracers.Store(name, &tracerWrap{t.tracer, t.closer, true})
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
	j.tracers.Store(name, &tracerWrap{tracer, closer, true})
	return
}

// Disable 关闭 tracer
func (j *Jaeger) Disable(name string) {
	if x, ok := j.tracers.Load(name); ok {
		t := x.(*tracerWrap)
		j.tracers.Store(name, &tracerWrap{t.tracer, t.closer, false})
	}
}

// Get 获取 tracer
func (j *Jaeger) Get(name string) opentracing.Tracer {
	if x, ok := j.tracers.Load(name); ok {
		if t := x.(*tracerWrap); t.vaild {
			return t.tracer
		}
	}
	return nil
}

type tracerWrap struct {
	tracer opentracing.Tracer
	closer io.Closer
	vaild  bool
}
