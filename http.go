/**
 * @Time : 19/05/2020 10:17 AM
 * @Author : solacowa@gmail.com
 * @File : http
 * @Software: GoLand
 */

package captcha

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

const (
	rateBucketNum = 100
)

func MakeHTTPHandler(s Service, opts []kithttp.ServerOption,
	ems []endpoint.Middleware,
	prefix string,
	encodeResponseFunc func(ctx context.Context, w http.ResponseWriter, response interface{}) (err error)) http.Handler {

	eps := NewEndpoint(s, map[string][]endpoint.Middleware{
		"Captcha": ems,
		"Refresh": ems,
		"Verify":  ems,
	}, prefix)

	r := mux.NewRouter()
	r.Handle(fmt.Sprintf("%s{captchaId}", prefix), kithttp.NewServer(
		eps.CaptchaEndpoint,
		decodeCaptchaRequest,
		encodeCaptchaResponse,
		opts...,
	)).Methods(http.MethodGet)
	r.Handle(fmt.Sprintf("%srefresh/image", prefix), kithttp.NewServer(
		eps.RefreshEndpoint,
		decodeRefreshRequest,
		encodeResponseFunc,
		opts...,
	)).Methods(http.MethodGet)

	return r

}

func decodeRefreshRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	w, _ := strconv.Atoi(r.URL.Query().Get("w"))
	h, _ := strconv.Atoi(r.URL.Query().Get("h"))

	if w < 10 {
		w = 160
	}

	if h < 10 {
		h = 80
	}

	return captchaRequest{W: w, H: h}, nil
}

func decodeCaptchaRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	captchaId, ok := vars["captchaId"]
	if !ok {
		return nil, errors.New("captchaId not exists")
	}

	w, _ := strconv.Atoi(r.URL.Query().Get("w"))
	h, _ := strconv.Atoi(r.URL.Query().Get("h"))

	if w < 10 {
		w = 160
	}

	if h < 10 {
		h = 80
	}

	return captchaRequest{CaptchaId: captchaId, W: w, H: h}, nil
}

func encodeCaptchaResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	resp, ok := response.(imageResponse)
	if !ok {
		return errors.New("response type not imageResponse")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Captcha-Id", resp.CaptchaId)

	_, _ = w.Write(resp.Content)
	return nil
}
