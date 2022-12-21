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
)

type logging struct {
	logger  log.Logger
	traceId string
	next    Service
}

func (s *logging) GenCaptchaId(ctx context.Context) string {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			s.traceId, ctx.Value(s.traceId),
			"method", "GenCaptchaId",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.GenCaptchaId(ctx)
}

func (s *logging) VerifyCaptcha(ctx context.Context, captchaId, verify string) bool {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			s.traceId, ctx.Value(s.traceId),
			"method", "VerifyCaptcha", "captchaId", captchaId, "verify", verify,
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.next.VerifyCaptcha(ctx, captchaId, verify)
}

func NewLogging(logger log.Logger, traceId string) Middleware {
	logger = log.With(logger, "pkg.git", "logging")
	return func(next Service) Service {
		return &logging{
			logger:  logger,
			next:    next,
			traceId: traceId,
		}
	}
}

func (s *logging) Image(ctx context.Context, captchaId string, w, h int) []byte {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			s.traceId, ctx.Value(s.traceId),
			"method", "Image",
			"captchaId", captchaId,
			"w", w,
			"h", h,
			"took", time.Since(begin),
			"err", "null",
		)
	}(time.Now())
	return s.next.Image(ctx, captchaId, w, h)
}

func (s *logging) Refresh(ctx context.Context, w, h int) (captchaId string, res []byte) {
	defer func(begin time.Time) {
		_ = s.logger.Log(
			s.traceId, ctx.Value(s.traceId),
			"method", "Refresh",
			"w", w,
			"h", h,
			"captchaId", captchaId,
			"took", time.Since(begin),
			"err", "null",
		)
	}(time.Now())
	return s.next.Refresh(ctx, w, h)
}
