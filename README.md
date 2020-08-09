# tracer

使用 opentracing ，封装常见的 golang 库，做全链路追踪

## 主要特点

- 可以开启或关闭 opentracing
- 详细的例子，演练模拟各种正常、异常下链路追踪情况

## 例子

模拟以下情景，做下全链路追踪练习

```
      +--------------------------------+
      |                                |                         +-------------+
      |                                |                         |    Redis    |
      |                                |                         +------+------+
      |                                |                                ^
      v                                |                                |
+-----+-------+     http        +-------------+     gRPC         +-------------+
|   server1   +---------------->+   server2   +----------------->+   server3   +---+
+-------------+                 +-------------+                  +-------------+   |
                                       ^                                |          |
                                       |                                v          |
                                       |                         +------+------+   |
                                       |                         |    MySQL    |   |
                                       |                         +-------------+   |
                                       |                                           |
                                       |                                           |
                                       +-------------------------------------------+
```

演练以下情景：
- 正常情况
  ```
  server1 --> HTTP --> server2
                          |--> HTTP --> server1
                          |-->  gRPC --> server3
                                            |--> MySQL
                                            |--> Redis
                                            |--> gRPC --> server2
  ```
  - 包括 HTTP 嵌套调用正常
  - 包括 gRPC 嵌套调用正常

- 各种异常
  - HTTP 执行 panic
  - gRPC 执行 panic
  - MySQL 执行失败
  - Redis 执行失败
  - 服务间调用死循环


## HTTP

目前 HTTP 使用 [github.com/labstack/echo](github.com/labstack/echo)

接入 tracer 代码：

```go
e.Use(tracer.EchoMiddleware(tracerName))
```

## gRPC

包括一元 RPC 调用追踪、流 RPC 调用追踪

接入 tracer 代码：

```go
s := grpc.NewServer(
	tracer.RPCServerOption(tracerName), // server tracer
)
```

```go
conn, err = grpc.Dial(addr,
	tracer.RPCClientOption(tracerName), // client tracer
)
```

## Redis

TODO

## MySQL

TODO

## Jaeger

jaeger 安装，参考： [https://www.jaegertracing.io/docs/1.18/getting-started/](https://www.jaegertracing.io/docs/1.18/getting-started/)

类似以下命令：

```vim
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.18
```

UI 界面： http://localhost:16686/

## Zipkin

TODO


## 集成报警

TODO


## 主要用途

- 开发环境，可以一直开着 tracer ，并通过集成报警，实时通知程序服务异常
- 生产环境，排查错误，需要时，打开 tracer ，协助分析问题


## 参考

- [https://github.com/opentracing-contrib](https://github.com/opentracing-contrib)
