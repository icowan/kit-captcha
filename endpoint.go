/**
 * @Time : 19/05/2020 10:20 AM
 * @Author : solacowa@gmail.com
 * @File : endpoint
 * @Software: GoLand
 */

package captcha

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
)

type captchaRequest struct {
	CaptchaId string
	W         int
	H         int
}

type imageResponse struct {
	Content   []byte
	CaptchaId string
}

type (
	VerifiedRequest struct {
		CaptchaId string `json:"captcha_id"`
		Verify    string `json:"verify"`
	}
	VerifiedResponse struct {
		Data bool `json:"data"`
	}
)

type GenerateResponse struct {
	CaptchaId  string `json:"captcha_id"`
	CaptchaUrl string `json:"captcha_url"`
}

type Endpoints struct {
	CaptchaEndpoint endpoint.Endpoint
	RefreshEndpoint endpoint.Endpoint
	ImageEndpoint   endpoint.Endpoint
	VerifyEndpoint  endpoint.Endpoint
}

func NewEndpoint(s Service, mdw map[string][]endpoint.Middleware, prefix string) Endpoints {
	eps := Endpoints{
		CaptchaEndpoint: makeCaptchaEndpoint(s),
		RefreshEndpoint: makeRefreshEndpoint(s, prefix),
		VerifyEndpoint:  makeVerifyEndpoint(s),
	}

	for _, m := range mdw["Captcha"] {
		eps.CaptchaEndpoint = m(eps.CaptchaEndpoint)
	}
	for _, m := range mdw["Refresh"] {
		eps.RefreshEndpoint = m(eps.RefreshEndpoint)
	}
	for _, m := range mdw["Verify"] {
		eps.VerifyEndpoint = m(eps.VerifyEndpoint)
	}

	return eps
}

func makeRefreshEndpoint(s Service, prefix string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		captchaId := s.GenCaptchaId(ctx)
		return GenerateResponse{
			CaptchaId:  captchaId,
			CaptchaUrl: fmt.Sprintf("%s%s", prefix, captchaId),
		}, nil
	}
}

func makeCaptchaEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(captchaRequest)
		data := s.Image(ctx, req.CaptchaId, req.W, req.H)
		return imageResponse{
			CaptchaId: req.CaptchaId,
			Content:   data,
		}, err
	}
}

func makeVerifyEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(VerifiedRequest)
		verified := s.VerifyCaptcha(ctx, req.CaptchaId, req.Verify)
		if !verified {
			err = errors.New("captcha not verified")
		}
		return VerifiedResponse{Data: verified}, err
	}
}
