package sdk

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_authErrorResp(t *testing.T) {
	type args struct {
		req *http.Request
	}

	errResp := func(msg string) *http.Response {
		o := Error{HTTPCode: http.StatusForbidden}
		o.Message = msg
		return o.httpResp()
	}

	tests := []struct {
		name string
		args args
		want *http.Response
	}{
		{
			name: "auth successful",
			args: args{
				req: &http.Request{
					Header: http.Header{
						"Authorization": []string{"Bearer validKey"},
					},
				},
			},
			want: nil,
		},
		{
			name: "auth not successful",
			args: args{
				req: &http.Request{
					Header: http.Header{
						"Authorization": []string{"Bearer invalidApiKey"},
					},
				},
			},
			want: errResp("authorization failed"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := authErrorResp(tt.args.req); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("authErrorResp() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
