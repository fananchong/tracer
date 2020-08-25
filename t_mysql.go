package tracer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql" //
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// MySQL 封装例子
// MySQL 客户端库太多了，因此这里演练下 trace 过程的例子
// 项目中的 MySQL 封装层，套用下本例子的使用即可

// MySQLPingWrap ping mysql for test
func MySQLPingWrap(ctx context.Context, tracerName string) {
	if tracer := Get(tracerName); tracer != nil {

		var parentCtx opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentCtx = parent.Context()
		}

		spanName := strings.ToUpper("Ping")
		span := tracer.StartSpan(spanName, opentracing.ChildOf(parentCtx))
		ext.DBType.Set(span, "MySQL")
		ext.DBStatement.Set(span, fmt.Sprintf("%v", "ping"))
		defer span.Finish()
		if err := ping(); err != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}

	} else {
		ping()
	}
}

func ping() (err error) {
	var db *sql.DB
	db, err = sql.Open("mysql", "fan:123456@tcp(172.26.150.124:3306)/test")
	if err != nil {
		return
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	return
}
