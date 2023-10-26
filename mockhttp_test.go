package sdk

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_newObjectPath(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want objPath
	}{
		{
			name: "projects",
			args: args{"/projects"},
			want: objPath{"/projects", false},
		},
		{
			name: "trailing slash",
			args: args{"/////"},
			want: objPath{"", false},
		},
		{
			name: "project",
			args: args{"/projects/fooBar"},
			want: objPath{"/projects/{project_id}", false},
		},
		{
			name: "project's branches",
			args: args{"/projects/fooBar/branches"},
			want: objPath{"/projects/{project_id}/branches", false},
		},
		{
			name: "project's branch",
			args: args{"/projects/fooBar/branches/qux"},
			want: objPath{"/projects/{project_id}/branches/{branch_id}", false},
		},
		{
			name: "project branch's endpoints",
			args: args{"/projects/fooBar/branches/qux/endpoints"},
			want: objPath{"/projects/{project_id}/branches/{branch_id}/endpoints", false},
		},
		{
			name: "project not found",
			args: args{"/projects/notFound/branches/qux/endpoints"},
			want: objPath{"/projects/{project_id}/branches/{branch_id}/endpoints", true},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := parsePath(tt.args.s); got != tt.want {
					t.Errorf("parsePath(%s) = %v, want %v", tt.args.s, got, tt.want)
				}
			},
		)
	}
}

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
