/**
 * @Time : 19/05/2020 9:38 AM
 * @Author : solacowa@gmail.com
 * @File : service
 * @Software: GoLand
 */

package captcha

import (
	"bytes"
	"context"
	"github.com/dchest/captcha"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type Service interface {
	// 获取图形验证码ID
	GenCaptchaId(ctx context.Context) string

	// 验证图形验证码
	VerifyCaptcha(ctx context.Context, captchaId, verify string) bool

	// 生成图片
	Image(ctx context.Context, captchaId string, w, h int) []byte

	// 刷新验证码图片
	Refresh(ctx context.Context, w, h int) (captchaId string, res []byte)
}

type service struct {
	logger  log.Logger
	store   captcha.Store
	traceId string
}

func (s *service) Refresh(ctx context.Context, w, h int) (captchaId string, res []byte) {
	captchaId = s.GenCaptchaId(ctx)
	res = s.Image(ctx, captchaId, w, h)
	return
}

func (s *service) Image(ctx context.Context, captchaId string, w, h int) []byte {
	logger := log.With(s.logger, s.traceId, ctx.Value(s.traceId))
	var content bytes.Buffer
	captcha.SetCustomStore(s.store)
	err := captcha.WriteImage(&content, captchaId, w, h)
	if err != nil {
		_ = level.Error(logger).Log("captcha", "WriteImage", "err", err.Error())
		return nil
	}
	return content.Bytes()
}

func (s *service) VerifyCaptcha(ctx context.Context, captchaId, verify string) bool {
	captcha.SetCustomStore(s.store)
	return captcha.VerifyString(captchaId, verify)
}

func (s *service) GenCaptchaId(ctx context.Context) string {
	captcha.SetCustomStore(s.store)
	captchaId := captcha.NewLen(6)
	return captchaId
}

func New(logger log.Logger, store captcha.Store, traceId string) Service {
	logger = log.With(logger, "service", "kit-captcha")
	return &service{
		logger:  logger,
		store:   store,
		traceId: traceId,
	}
}
