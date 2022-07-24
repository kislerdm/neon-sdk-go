package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
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
	if token == "" || token == "Bearer fail" {
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

var mockHttpClient = &httpClientMock{
	m: map[string]map[reqType]resp{
		urlPrefix + "users/me": {
			get: func(req *http.Request) (*http.Response, error) {
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
		urlPrefix + "projects": {
			post: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}

				var v ProjectSettingsRequestCreate
				buf, _ := io.ReadAll(req.Body)
				_ = json.Unmarshal(buf, &v)
				switch v.Name {
				case "fail":
					return &http.Response{
						StatusCode: 500,
						Body: io.NopCloser(
							strings.NewReader(`{"message":"internal error","code":""}`),
						),
					}, nil
				case "fail-response":
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{`)),
					}, nil
				}

				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`{
    "id": "quiet-river-711967",
    "parent_id": null,
    "roles": [
      {
        "id": 1,
        "name": "foo",
        "password": ""
      }
    ],
    "databases": [
      {
        "id": 1,
        "name": "main",
        "owner_id": 1
      }
    ],
    "name": "quiet-river-711967",
    "created_at": "2022-07-24T11:18:12.322513Z",
    "updated_at": "2022-07-24T11:18:18.389868Z",
    "region_id": "us-west-2",
    "instance_handle": "",
    "instance_type_id": "0",
    "region_name": "US West (Oregon)",
    "platform_name": "Serverless",
    "platform_id": "serverless",
    "settings": {},
    "pending_state": null,
    "current_state": "active",
    "deleted": false,
    "size": 0,
    "max_project_size": 0,
    "pooler_enabled": false
  }`,
						),
					),
				}, nil
			},
		},
		urlPrefix + "projects/fail/delete": {
			post: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
				}, nil
			},
		},
		urlPrefix + "projects/fail-response/delete": {
			post: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{`)),
				}, nil
			},
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

func mustParse(s string) time.Time {
	o, _ := time.Parse(time.RFC3339, s)
	return o
}

func Test_client_CreateProject(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	type args struct {
		settings ProjectSettingsRequestCreate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ProjectInfo
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				options: Options{
					APIKey:     "foo",
					HTTPClient: mockHttpClient,
				},
				baseURL: baseURL,
			},
			args: args{
				settings: ProjectSettingsRequestCreate{
					Name: "foo",
				},
			},
			want: ProjectInfo{
				CreatedAt:    mustParse("2022-07-24T11:18:12.322513Z"),
				CurrentState: "active",
				Databases: []Database{
					{
						ID:      1,
						Name:    "main",
						OwnerId: 1,
					},
				},
				Deleted:        false,
				ID:             "quiet-river-711967",
				InstanceHandle: "",
				InstanceTypeID: "0",
				MaxProjectSize: 0,
				Name:           "quiet-river-711967",
				ParentID:       "",
				PendingState:   "",
				PlatformID:     "serverless",
				PlatformName:   "Serverless",
				PoolerEnabled:  false,
				RegionID:       "us-west-2",
				RegionName:     "US West (Oregon)",
				Roles: []Role{
					{
						ID:       1,
						Name:     "foo",
						Password: "",
					},
				},
				Settings:  AdditionalOptions{},
				Size:      0,
				UpdatedAt: mustParse("2022-07-24T11:18:18.389868Z"),
			},
			wantErr: false,
		},
		{
			name: "invalid - server error",
			fields: fields{
				options: Options{
					APIKey:     "foo",
					HTTPClient: mockHttpClient,
				},
				baseURL: baseURL,
			},
			args: args{
				settings: ProjectSettingsRequestCreate{
					Name: "fail",
				},
			},
			want:    ProjectInfo{},
			wantErr: true,
		},
		{
			name: "invalid - faulty response",
			fields: fields{
				options: Options{
					APIKey:     "foo",
					HTTPClient: mockHttpClient,
				},
				baseURL: baseURL,
			},
			args: args{
				settings: ProjectSettingsRequestCreate{
					Name: "fail-response",
				},
			},
			want:    ProjectInfo{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := &client{
					options: tt.fields.options,
					baseURL: tt.fields.baseURL,
				}
				got, err := c.CreateProject(tt.args.settings)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteProject(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	type args struct {
		projectID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ProjectDeleteResponse
		wantErr bool
	}{
		//{
		//	name: "valid",
		//	fields: fields{
		//		options: Options{
		//			APIKey:     "foo",
		//			HTTPClient: mockHttpClient,
		//		},
		//		baseURL: baseURL,
		//	},
		//	args: args{
		//		projectID: "foo",
		//	},
		//	want:    ProjectDeleteResponse{},
		//	wantErr: false,
		//},
		{
			name: "invalid - server error",
			fields: fields{
				options: Options{
					APIKey:     "foo",
					HTTPClient: mockHttpClient,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "fail",
			},
			want:    ProjectDeleteResponse{},
			wantErr: true,
		},
		{
			name: "invalid - faulty response",
			fields: fields{
				options: Options{
					APIKey:     "foo",
					HTTPClient: mockHttpClient,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "fail-response",
			},
			want:    ProjectDeleteResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := &client{
					options: tt.fields.options,
					baseURL: tt.fields.baseURL,
				}
				got, err := c.DeleteProject(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
