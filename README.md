# go-kit 简易图型验证码


图形生成包使用的是 [github.com/dchest/captcha](github.com/dchest/captcha)，为了方便go-kit使用，我对期进行了一层封装。
封装之后的包在go-kit的架构里只用两行代码就可以搞定图型验证码的生成,刷新,验证等功能

## 安装

```go
$ go get github.com/icowan/kit-captcha
```

## 使用

在您的应用启动http服务的初始化代码参考: 

```go 

import (
	"context"
	"net/http"
	"os"
	
	"github.com/dchest/captcha"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
)

func main() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.StdlibWriter{})

    var traceKey = "trace-id"

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(level.Error(logger))),
		kithttp.ServerBefore(func(ctx context.Context, request *http.Request) context.Context {
			ctx = context.WithValue(ctx, "context-trace-key", traceKey)
			return ctx
		}),
	}

	var ems []endpoint.Middleware

	svc := New(logger, captcha.NewMemoryStore(
		captcha.CollectNum,
		captcha.Expiration,
	), traceKey)

    // 不想看到日志的可以不加
	svc = NewLoggingServer(logger, svc)

	var prefix = "/captcha/"

	mux := http.NewServeMux()
	mux.Handle(prefix, MakeHTTPHandler(logger, svc, opts, ems, prefix, func(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
		return kithttp.EncodeJSONResponse(ctx, w, response)
	}))

	http.Handle("/", accessControl(mux, logger))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		os.Exit(1)
	}
}

func accessControl(h http.Handler, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			return
		}
		_ = level.Info(logger).Log("remote-addr", r.RemoteAddr, "uri", r.RequestURI, "method", r.Method, "length", r.ContentLength)

		h.ServeHTTP(w, r)
	})
}

```

服务启动: `go run main.go`


### 获取图型验证码ID

```
$ curl http://localhost:8080/captcha/refresh/image
{"captcha_id":"IOAUynIXiqfUs56dQgfg","captcha_url":"/captcha/IOAUynIXiqfUs56dQgfg"}
```

### 生成图形验证码

拿到上面获取到的captcha_id在浏览器打开 `http://localhost:8080/captcha/IOAUynIXiqfUs56dQgfg` 就能展示出图形验证码。

![](http://source.qiniu.cnd.nsini.com/images/2020/09/9e/8d/3d/20200930-947c5cc13679d32a3106b9a78dd9e2eb.jpeg?imageView2/2/w/1280/interlace/0/q/70)

可以传参数:

- `w`: 图片的宽度
- `h`: 图片的高度

默认生成的图片是 160 x 80，您可以自己定义如:

`http://localhost:8080/captcha/IOAUynIXiqfUs56dQgfg?w=320&h=120`

![](http://source.qiniu.cnd.nsini.com/images/2020/09/84/54/4e/20200930-01792ce4837ae3794b30bb9e54c8448c.jpeg?imageView2/2/w/1280/interlace/0/q/70)


### 验证图形验证码

结合到您的服务上进行验证:

```go
svc := New(logger, captcha.NewMemoryStore(
    captcha.CollectNum,
    captcha.Expiration,
), "trace-id")

// 不想看到日志的可以不加
svc = NewLoggingServer(logger, svc)

// 使用验证 一般在中间件使用
if !svc.VerifyCaptcha(ctx, req.Query.Get("captchaId"), req.Query.Get("verifyCode")) {
    fmt.Println("验证码错误")
} 
```

### 使用第三方存储

默认使用内存进行存储验证码信息，多个节点建议使用第三方存储方案。

我这里使用了一个单点或集群Redis都支持的包: [github.com/icowan/redis-client](github.com/icowan/redis-client)，不喜欢的可以按照自己的需求实现一个就行，您自定义的存储方案只需要实现以下两个接口就行:

```go
type Store interface {
	// Set sets the digits for the captcha id.
	Set(id string, digits []byte)

	// Get returns stored digits for the captcha id. Clear indicates
	// whether the captcha must be deleted from the store.
	Get(id string, clear bool) (digits []byte)
}
```

以下是我使用Redis存储方案的参考: 

**redisstorage.go**

```go

import (
	"time"

	"github.com/dchest/captcha"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	redisclient "github.com/icowan/redis-client"
)

type store struct {
	rds        redisclient.RedisClient
	expiration time.Duration
	logger     log.Logger
	prefix     string
}

func (s *store) Set(id string, digits []byte) {
	err := s.rds.Set(s.pre(id), string(digits), s.expiration)
	if err != nil {
		_ = level.Error(s.logger).Log("rds", "set", "id", id, "err", err.Error())
	}
}

func (s *store) Get(id string, clear bool) (digits []byte) {
	v, err := s.rds.Get(s.pre(id))
	if err != nil {
		_ = level.Error(s.logger).Log("rds", "get", "id", id, "clear", clear, "err", err.Error())
	}
	if clear {
		//_ = s.rds.Del(s.pre(id))
	}
	return []byte(v)
}

func (s *store) pre(id string) string {
	return s.prefix + id
}

func NewStore(rds redisclient.RedisClient, logger log.Logger, expiration time.Duration) captcha.Store {
	return &store{
		rds:        rds,
		logger:     logger,
		expiration: expiration,
		prefix:     "captcha:",
	}
}
```

**使用存储方案参考**

```go
import redisclient "github.com/icowan/redis-client"

rdsClient := redisclient.NewRedisClient(...)
svc := New(logger, NewStore(rdsClient, logger, time.Minute*5), "trace-id")
```

