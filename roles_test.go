package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

var mockHttpClientRoles = &httpClientMock{
	m: map[string]map[string]resp{
		urlPrefix + "projects/validProjectID/roles": {
			"GET": func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				if os.Getenv("SDK_TEST_RESPONSE_FAIL") == "TRUE" {
					return &http.Response{
						StatusCode:    200,
						ContentLength: 1000,
						Body:          io.NopCloser(strings.NewReader(`{`)),
					}, nil
				}
				if os.Getenv("SDK_TEST_INTERNAL_ERROR") == "TRUE" {
					return &http.Response{
						StatusCode:    500,
						ContentLength: 1000,
						Body:          io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`[
  {
    "created_at": "2022-07-27T09:13:06.855Z",
    "id": 0,
    "name": "validRole",
    "password": "xxx",
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
  }
]`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
			"POST": func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				if os.Getenv("SDK_TEST_INTERNAL_ERROR") == "TRUE" {
					return &http.Response{
						StatusCode:    500,
						Body:          io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
						ContentLength: 1000,
					}, nil
				}
				if os.Getenv("SDK_TEST_RESPONSE_FAIL") == "TRUE" {
					return &http.Response{
						StatusCode:    200,
						ContentLength: 100,
						Body:          io.NopCloser(strings.NewReader(`{`)),
					}, nil
				}

				b, err := io.ReadAll(req.Body)
				if err != nil {
					panic(err)
				}
				var r RoleRequest
				if err := json.Unmarshal(b, &r); err != nil {
					panic(err)
				}

				return &http.Response{
					StatusCode:    200,
					ContentLength: 1000,
					Body: io.NopCloser(
						strings.NewReader(
							fmt.Sprintf(
								`{
    "created_at": "2022-07-27T09:13:06.855Z",
    "id": 0,
    "name": "%s",
    "password": "xxx",
	"protected": true,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
  }`, r.Role.Name,
							),
						),
					),
				}, nil
			},
		},
		urlPrefix + "projects/invalidProjectID/roles": {
			"GET":  objNotFoundResponse,
			"POST": objNotFoundResponse,
		},

		urlPrefix + "projects/validProjectID/roles/validRole": {
			"GET": func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				if os.Getenv("SDK_TEST_RESPONSE_FAIL") == "TRUE" {
					return &http.Response{
						StatusCode:    200,
						ContentLength: 1000,
						Body:          io.NopCloser(strings.NewReader(`{`)),
					}, nil
				}
				if os.Getenv("SDK_TEST_INTERNAL_ERROR") == "TRUE" {
					return &http.Response{
						StatusCode:    500,
						ContentLength: 1000,
						Body:          io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`{
    "created_at": "2022-07-27T09:13:06.855Z",
    "id": 0,
    "name": "validRole",
    "password": "xxx",
	"protected": true,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
}`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
			"DELETE": func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				if os.Getenv("SDK_TEST_RESPONSE_FAIL") == "TRUE" {
					return &http.Response{
						StatusCode:    200,
						ContentLength: 1000,
						Body:          io.NopCloser(strings.NewReader(`{`)),
					}, nil
				}
				if os.Getenv("SDK_TEST_INTERNAL_ERROR") == "TRUE" {
					return &http.Response{
						StatusCode:    500,
						ContentLength: 1000,
						Body:          io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`{
    "created_at": "2022-07-27T09:13:06.855Z",
    "id": 0,
    "name": "validRole",
    "password": "xxx",
	"protected": true,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
}`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
		},
		urlPrefix + "projects/validProjectID/roles/invalidRole": {
			"GET":    objNotFoundResponse,
			"DELETE": objNotFoundResponse,
		},

		urlPrefix + "projects/validProjectID/roles/validRole/reset_password": {
			"POST": func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				if os.Getenv("SDK_TEST_INTERNAL_ERROR") == "TRUE" {
					return &http.Response{
						StatusCode:    500,
						Body:          io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
						ContentLength: 1000,
					}, nil
				}
				if os.Getenv("SDK_TEST_RESPONSE_FAIL") == "TRUE" {
					return &http.Response{
						StatusCode:    200,
						ContentLength: 100,
						Body:          io.NopCloser(strings.NewReader(`{`)),
					}, nil
				}
				return &http.Response{
					StatusCode:    200,
					ContentLength: 1000,
					Body: io.NopCloser(
						strings.NewReader(
							`{
	"operation_id": 1,
	"password": "xxx"
}`,
						),
					),
				}, nil
			},
		},
		urlPrefix + "projects/validProjectID/roles/invalidRole/reset_password": {
			"POST": objNotFoundResponse,
		},
	},
}

func Test_client_ListRoles(t *testing.T) {
	type fields struct {
		options options
		baseURL string
	}
	type args struct {
		projectID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    []RoleResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			want: []RoleResponse{
				{
					CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
					ID:        0,
					Name:      "validRole",
					ProjectID: "validProjectID",
					Password:  "xxx",
					UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				},
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: invalid projectID",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "invalidProjectID"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for k, v := range tt.envVars {
					_ = os.Setenv(k, v)
				}

				c := &client{
					options: tt.fields.options,
					baseURL: tt.fields.baseURL,
				}
				got, err := c.ListRoles(tt.args.projectID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("ListRoles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListRoles() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateRole(t *testing.T) {
	type fields struct {
		options options
		baseURL string
	}
	type args struct {
		projectID string
		cfg       RoleRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    RoleResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
				cfg: RoleRequest{
					Role: struct {
						Name string `json:"name"`
					}{
						Name: "foo",
					},
				},
			},
			want: RoleResponse{
				CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				Name:      "foo",
				Password:  "xxx",
				ProjectID: "validProjectID",
				ID:        0,
				Protected: true,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    RoleResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    RoleResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: project not found",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "invalidProjectID"},
			want:    RoleResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for k, v := range tt.envVars {
					_ = os.Setenv(k, v)
				}

				c := &client{
					options: tt.fields.options,
					baseURL: tt.fields.baseURL,
				}
				got, err := c.CreateRole(tt.args.projectID, tt.args.cfg)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("CreateRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateRole() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ReadRole(t *testing.T) {
	type fields struct {
		options options
		baseURL string
	}
	type args struct {
		projectID string
		roleName  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    RoleResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
				roleName:  "validRole",
			},
			want: RoleResponse{
				CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				Name:      "validRole",
				Password:  "xxx",
				ProjectID: "validProjectID",
				ID:        0,
				Protected: true,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID", roleName: "validRole"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    RoleResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", roleName: "validRole"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    RoleResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: role not found",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", roleName: "invalidRole"},
			want:    RoleResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for k, v := range tt.envVars {
					_ = os.Setenv(k, v)
				}

				c := &client{
					options: tt.fields.options,
					baseURL: tt.fields.baseURL,
				}
				got, err := c.ReadRole(tt.args.projectID, tt.args.roleName)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("ReadRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ReadRole() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteRole(t *testing.T) {
	type fields struct {
		options options
		baseURL string
	}
	type args struct {
		projectID string
		roleName  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    RoleResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID", roleName: "validRole"},
			want: RoleResponse{
				CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				Name:      "validRole",
				Password:  "xxx",
				ProjectID: "validProjectID",
				ID:        0,
				Protected: true,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID", roleName: "validRole"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    RoleResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", roleName: "validRole"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    RoleResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: role not found",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", roleName: "invalidRole"},
			want:    RoleResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for k, v := range tt.envVars {
					_ = os.Setenv(k, v)
				}

				c := &client{
					options: tt.fields.options,
					baseURL: tt.fields.baseURL,
				}
				got, err := c.DeleteRole(tt.args.projectID, tt.args.roleName)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteRole() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ResetRolePassword(t *testing.T) {
	type fields struct {
		options options
		baseURL string
	}
	type args struct {
		projectID string
		roleName  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    RolePasswordResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
				roleName:  "validRole",
			},
			want: RolePasswordResponse{
				Password:    "xxx",
				OperationID: 1,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID", roleName: "validRole"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    RolePasswordResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", roleName: "validRole"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    RolePasswordResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: role not found",
			fields: fields{
				options: options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientRoles,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", roleName: "invalidRole"},
			want:    RolePasswordResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for k, v := range tt.envVars {
					_ = os.Setenv(k, v)
				}

				c := &client{
					options: tt.fields.options,
					baseURL: tt.fields.baseURL,
				}
				got, err := c.ResetRolePassword(tt.args.projectID, tt.args.roleName)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("ResetRolePassword() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ResetRolePassword() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
