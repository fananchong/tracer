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
	ctx        context.Context
}

// NewRedisClient RedisClient 构造函数
func NewRedisClient(ctx context.Context, tracerName string, redis *redis.Client) *RedisClient {
	rclient := &RedisClient{
		tracerName: tracerName,
	}
	rclient.Client = redis.WithContext(ctx)
	rclient.Client.WrapProcess(rclient.process())
	return rclient
}

func (rclient *RedisClient) process() func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {

			if tracer := Get(rclient.tracerName); tracer != nil {

				var parentCtx opentracing.SpanContext
				if parent := opentracing.SpanFromContext(rclient.Client.Context()); parent != nil {
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
