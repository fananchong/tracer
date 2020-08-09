package tracer

import (
	"github.com/fananchong/tracer/internal/jaeger"
	"github.com/opentracing/opentracing-go"
)

// ITracer tracer 访问接口
type ITracer interface {
	Enable(name string) (err error)
	Disable(name string)
	Get(name string) opentracing.Tracer
}

// Enable 打开 tracer
func Enable(name string) (err error) {
	return DefaultTracer.Enable(name)
}

// Disable 关闭 tracer
func Disable(name string) {
	DefaultTracer.Disable(name)
}

// Get 获取 tracer
func Get(name string) opentracing.Tracer {
	return DefaultTracer.Get(name)
}

// DefaultTracer tracer 具体实例
var DefaultTracer ITracer

// Usejaeger 使用 jaeger 做为 tracer
func Usejaeger() {
	DefaultTracer = jaeger.New()
}

func init() {
	Usejaeger()
}
