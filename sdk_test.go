package sdk

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type resp func(*http.Request) (*http.Response, error)

type httpClientMock struct {
	m map[string]resp
}

func (h *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	return h.m[req.URL.EscapedPath()](req)
}

const urlPrefix = "/api/v1/"

var mockHttpClient = &httpClientMock{
	m: map[string]resp{
		urlPrefix + "users/me": func(req *http.Request) (*http.Response, error) {
			token := req.Header.Get("Authorization")
			if token == "" || token == "Bearer fail" {
				return &http.Response{
					Status:     "",
					StatusCode: 302,
				}, fmt.Errorf("not authoraised")
			}
			return &http.Response{
				StatusCode: 200,
				Body:       nil,
			}, nil
		},
	},
}

func TestNewClient(t *testing.T) {
	type args struct {
		ctx    context.Context
		optFns []func(*Options)
	}
	tests := []struct {
		name      string
		args      args
		envVarKey string
		want      Client
		wantErr   bool
	}{
		{
			name: "happy path",
			args: args{
				ctx: nil,
				optFns: []func(*Options){
					WithAPIKey("foobar"),
					WithHTTPClient(mockHttpClient),
				},
			},
			envVarKey: "",
			want: &client{
				options: Options{
					APIKey:     "foobar",
					HTTPClient: mockHttpClient,
				},
				baseURL: baseURL,
			},
			wantErr: false,
		},
		{
			name: "happy path - apiKey from env var",
			args: args{
				ctx: nil,
				optFns: []func(*Options){
					WithHTTPClient(mockHttpClient),
				},
			},
			envVarKey: "foobar",
			want: &client{
				options: Options{
					APIKey:     "foobar",
					HTTPClient: mockHttpClient,
				},
				baseURL: baseURL,
			},
			wantErr: false,
		},
		{
			name: "unhappy path - missing apiKey",
			args: args{
				ctx: nil,
				optFns: []func(*Options){
					WithHTTPClient(mockHttpClient),
				},
			},
			envVarKey: "",
			want:      nil,
			wantErr:   true,
		},
		{
			name: "unhappy path - wrong apiKey",
			args: args{
				ctx: nil,
				optFns: []func(*Options){
					WithHTTPClient(mockHttpClient),
				},
			},
			envVarKey: "fail",
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_ = os.Setenv("NEON_API_KEY", tt.envVarKey)

				got, err := NewClient(tt.args.ctx, tt.args.optFns...)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewClient() got = %v, want %v", got, tt.want)
				}
			},
		)

		t.Cleanup(
			func() {
				_ = os.Unsetenv("NEON_API_KEY")
			},
		)
	}
}
