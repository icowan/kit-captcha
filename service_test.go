/**
 * @Time: 2020/9/30 14:05
 * @Author: solacowa@gmail.com
 * @File: service_test
 * @Software: GoLand
 */

package captcha

import (
	"context"
	"github.com/dchest/captcha"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"
	"os"
)

func ExampleMakeHTTPHandler() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.StdlibWriter{})

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(level.Error(logger))),
		kithttp.ServerBefore(kithttp.PopulateRequestContext),
		kithttp.ServerBefore(func(ctx context.Context, request *http.Request) context.Context {
			ctx = context.WithValue(ctx, "context-trace-key", "trace-id")
			return ctx
		}),
	}

	var ems []endpoint.Middleware

	svc := New(logger, captcha.NewMemoryStore(
		captcha.CollectNum,
		captcha.Expiration,
	), "trace-id")

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
