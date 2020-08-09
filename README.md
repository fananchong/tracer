# tracer

使用 opentracing ，封装常见的 golang 库，做全链路追踪



## 例子

模拟以下场景，做下全链路追踪练习

```
                                                                 +-------------+
                                                                 |    Redis    |
                                                                 +------+------+
                                                                        ^
                                                                        |
+-------------+     http        +-------------+     gRPC         +-------------+
|   server1   +---------------->+   server2   +----------------->+   server3   |
+-------------+                 +-------------+                  +-------------+
                                                                        |
                                                                        v
                                                                 +------+------+
                                                                 |    MySQL    |
                                                                 +-------------+
```


## jaeger

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


## 参考

- [https://github.com/opentracing-contrib](https://github.com/opentracing-contrib)
