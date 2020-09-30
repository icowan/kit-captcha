/**
 * @Time : 19/05/2020 10:21 AM
 * @Author : solacowa@gmail.com
 * @File : logger
 * @Software: GoLand
 */

package captcha

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type loggingServer struct {
	logger log.Logger
	Service
}

func NewLoggingServer(logger log.Logger, s Service) Service {
	return &loggingServer{
		logger:  level.Info(logger),
		Service: s,
	}
}

func (s *loggingServer) Image(ctx context.Context, captchaId string, w, h int) []byte {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			ctx.Value("context-trace-key"), ctx.Value(ctx.Value("context-trace-key")),
			"method", "Image",
			"captchaId", captchaId,
			"w", w,
			"h", h,
			"took", time.Since(begin),
			"err", "null",
		)
	}(time.Now())
	return s.Service.Image(ctx, captchaId, w, h)
}

func (s *loggingServer) Refresh(ctx context.Context, w, h int) (captchaId string, res []byte) {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			ctx.Value("context-trace-key"), ctx.Value(ctx.Value("context-trace-key")),
			"method", "Refresh",
			"w", w,
			"h", h,
			"captchaId", captchaId,
			"took", time.Since(begin),
			"err", "null",
		)
	}(time.Now())
	return s.Service.Refresh(ctx, w, h)
}
