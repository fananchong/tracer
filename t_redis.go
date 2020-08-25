package tracer

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// RedisClient Redis 客户端
type RedisClient struct {
	*redis.Client
	tracerName string
}

// NewRedisClient RedisClient 构造函数
func NewRedisClient(ctx context.Context, tracerName string, redis *redis.Client) *RedisClient {
	rclient := &RedisClient{
		Client:     redis,
		tracerName: tracerName,
	}
	rclient.Client.WithContext(ctx)
	rclient.Client.WrapProcess(rclient.process(ctx))
	rclient.Client.WrapProcessPipeline(rclient.processPipeline(ctx))
	return rclient
}

func (rclient *RedisClient) process(ctx context.Context) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			if tracer := Get(rclient.tracerName); tracer != nil {

				var parentCtx opentracing.SpanContext
				if parent := opentracing.SpanFromContext(ctx); parent != nil {
					parentCtx = parent.Context()
				}

				spanName := strings.ToUpper(cmd.Name())
				span := tracer.StartSpan(spanName, opentracing.ChildOf(parentCtx))
				ext.DBType.Set(span, "redis")
				ext.DBStatement.Set(span, fmt.Sprintf("%v", cmd.Args()))
				defer span.Finish()

				span.LogFields(log.Object("Redis Cmd", cmd.Name()))
				span.LogFields(log.Object("Redis Cmd", cmd.Args()))
				err := oldProcess(cmd)
				if err != nil {
					ext.Error.Set(span, true)
					span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
				}

				return err
			}
			return oldProcess(cmd)
		}
	}
}

func (rclient *RedisClient) processPipeline(ctx context.Context) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			if tracer := Get(rclient.tracerName); tracer != nil {
				pipelineSpan, ctx := opentracing.StartSpanFromContext(ctx, "(pipeline)")

				ext.DBType.Set(pipelineSpan, "redis")

				for i := len(cmds); i > 0; i-- {
					cmdName := strings.ToUpper(cmds[i-1].Name())
					if cmdName == "" {
						cmdName = "(empty command)"
					}

					span, _ := opentracing.StartSpanFromContext(ctx, cmdName)
					ext.DBType.Set(span, "redis")
					ext.DBStatement.Set(span, fmt.Sprintf("%v", cmds[i-1].Args()))
					defer span.Finish()
				}

				defer pipelineSpan.Finish()

				return oldProcess(cmds)
			}
			return oldProcess(cmds)
		}
	}
}
