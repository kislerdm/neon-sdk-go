package sdk

import (
	"io"
	"net/http"
	"strings"
	"time"
)

type resp func(*http.Request) (*http.Response, error)

type httpClientMock struct {
	m map[string]map[reqType]resp
}

func (h *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	p := req.URL.Path
	m := reqType(req.Method)
	return h.m[p][m](req)
}

func authErrorResp(req *http.Request) *http.Response {
	token := req.Header.Get("Authorization")
	if token == "" || token == "Bearer invalidApiKey" {
		return &http.Response{
			Status:     "",
			StatusCode: 403,
			Body: io.NopCloser(
				strings.NewReader(`{"message":"authorization failed","code":""}`),
			),
		}
	}
	return nil
}

const urlPrefix = "/api/v1/"

func mustParseTime(s string) time.Time {
	o, _ := time.Parse(time.RFC3339Nano, s)
	return o
}

func objNotFoundResponse(req *http.Request) (*http.Response, error) {
	if resp := authErrorResp(req); resp != nil {
		return resp, nil
	}
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader(`{"message":"object not found","code":""}`)),
	}, nil
}
