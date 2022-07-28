package sdk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

var mockHttpClientValidateAPI = &httpClientMock{
	m: map[string]map[reqType]resp{
		urlPrefix + "users/me": {
			get: func(req *http.Request) (*http.Response, error) {
				token := req.Header.Get("Authorization")
				if token == "" || token == "Bearer invalidApiKey" {
					return &http.Response{
						Status:     "",
						StatusCode: 302,
					}, fmt.Errorf("not authoraised")
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`{
  "id": "ec82e2ee-0500-4f9f-b925-43ecdd1c3e89",
  "email": "admin@dkisler.com",
  "login": "kislerdm",
  "name": "Dmitry Kisler",
  "image": "https://avatars.githubusercontent.com/u/13434797?v=4",
  "projects_limit": 3,
  "auth_accounts": [
    {
      "provider": "github",
      "email": "admin@dkisler.com",
      "name": "Dmitry Kisler",
      "login": "kislerdm",
      "image": "https://avatars.githubusercontent.com/u/13434797?v=4"
    }
  ]
}`,
						),
					),
					ContentLength: 370,
				}, nil
			},
		},
	},
}

func TestNewClient(t *testing.T) {
	type args struct {
		ctx    context.Context
		optFns []func(*options)
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
				optFns: []func(*options){
					WithAPIKey("validApiKey"),
					WithHTTPClient(mockHttpClientValidateAPI),
				},
			},
			envVarKey: "",
			want: &client{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientValidateAPI,
				},
				baseURL: baseURL,
			},
			wantErr: false,
		},
		{
			name: "happy path - apiKey from env var",
			args: args{
				ctx: nil,
				optFns: []func(*options){
					WithHTTPClient(mockHttpClientValidateAPI),
				},
			},
			envVarKey: "validApiKey",
			want: &client{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientValidateAPI,
				},
				baseURL: baseURL,
			},
			wantErr: false,
		},
		{
			name: "unhappy path - missing apiKey",
			args: args{
				ctx: nil,
				optFns: []func(*options){
					WithHTTPClient(mockHttpClientValidateAPI),
				},
			},
			envVarKey: "",
			want:      nil,
			wantErr:   true,
		},
		{
			name: "unhappy path - invalid API key",
			args: args{
				ctx: nil,
				optFns: []func(*options){
					WithHTTPClient(mockHttpClientValidateAPI),
				},
			},
			envVarKey: "invalidApiKey",
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
