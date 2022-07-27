package sdk

import (
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

var mockHttpClientProjects = &httpClientMock{
	m: map[string]map[reqType]resp{
		// create and list end point
		urlPrefix + "projects": {
			// create project
			post: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				if os.Getenv("SDK_TEST_INTERNAL_ERROR") == "TRUE" {
					return &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
					}, nil
				}
				if os.Getenv("SDK_TEST_RESPONSE_FAIL") == "TRUE" {
					return &http.Response{
						StatusCode:    200,
						ContentLength: 1000,
						Body:          io.NopCloser(strings.NewReader(`{`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 201,
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
    "name": "foo",
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

			// list projects
			get: func(req *http.Request) (*http.Response, error) {
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
							`[{
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
    "name": "foo",
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
  }]`,
						),
					),
				}, nil
			},
		},

		// read project info end point
		urlPrefix + "projects/validProjectID": {
			get: func(req *http.Request) (*http.Response, error) {
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
    "id": "validProjectID",
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
    "name": "foo",
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
					ContentLength: 1000,
				}, nil
			},

			// update project
			patch: func(req *http.Request) (*http.Response, error) {
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
    "id": "validProjectID",
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
    "name": "bar",
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

		// test in case the project does not exist
		urlPrefix + "projects/invalidProjectID": {
			get: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader(`{"message":"object not found","code":""}`)),
				}, nil
			},

			patch: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader(`{"message":"object not found","code":""}`)),
				}, nil
			},
		},

		// del end point
		urlPrefix + "projects/validProjectID/delete": {
			post: func(req *http.Request) (*http.Response, error) {
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
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`{
  "id": 281749,
  "created_at": "2022-07-26T08:56:59.366068468Z",
  "updated_at": "2022-07-26T08:56:59.366068468Z",
  "project_id": "validProjectID",
  "uuid": "ab8421dd-fd07-4bf7-bd8f-4952f430eb1d",
  "action": "stop_compute",
  "status": "running",
  "error": "",
  "failures_count": 0,
  "retry_at": "0001-01-01T00:00:00Z"
}`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
		},
		urlPrefix + "projects/invalidProjectID/delete": {
			post: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader(`{"message":"project not found","code":""}`)),
				}, nil
			},
		},

		// start project end point
		urlPrefix + "projects/validProjectID/start": {
			post: func(req *http.Request) (*http.Response, error) {
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
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`{
  "id": 281749,
  "created_at": "2022-07-26T08:56:59.366068468Z",
  "updated_at": "2022-07-26T08:56:59.366068468Z",
  "project_id": "validProjectID",
  "uuid": "ab8421dd-fd07-4bf7-bd8f-4952f430eb1d",
  "action": "start_compute",
  "status": "running",
  "error": "",
  "failures_count": 0,
  "retry_at": "0001-01-01T00:00:00Z"
}`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
		},
		urlPrefix + "projects/invalidProjectID/start": {
			post: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader(`{"message":"project not found","code":""}`)),
				}, nil
			},
		},

		// stop project end point
		urlPrefix + "projects/validProjectID/stop": {
			post: func(req *http.Request) (*http.Response, error) {
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
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader(`{"message":"internal error","code":""}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(
						strings.NewReader(
							`{
  "id": 281749,
  "created_at": "2022-07-26T08:56:59.366068468Z",
  "updated_at": "2022-07-26T08:56:59.366068468Z",
  "project_id": "validProjectID",
  "uuid": "ab8421dd-fd07-4bf7-bd8f-4952f430eb1d",
  "action": "stop_compute",
  "status": "running",
  "error": "",
  "failures_count": 0,
  "retry_at": "0001-01-01T00:00:00Z"
}`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
		},
		urlPrefix + "projects/invalidProjectID/stop": {
			post: func(req *http.Request) (*http.Response, error) {
				if resp := authErrorResp(req); resp != nil {
					return resp, nil
				}
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader(`{"message":"project not found","code":""}`)),
				}, nil
			},
		},
	},
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
		envVars map[string]string
		want    ProjectInfo
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				settings: ProjectSettingsRequestCreate{
					Name: "foo",
				},
			},
			want: ProjectInfo{
				CreatedAt:    mustParseTime("2022-07-24T11:18:12.322513Z"),
				CurrentState: "active",
				Databases: []Database{
					{
						ID:      1,
						Name:    "main",
						OwnerID: 1,
					},
				},
				Deleted:        false,
				ID:             "quiet-river-711967",
				InstanceHandle: "",
				InstanceTypeID: "0",
				MaxProjectSize: 0,
				Name:           "foo",
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
				UpdatedAt: mustParseTime("2022-07-24T11:18:18.389868Z"),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				settings: ProjectSettingsRequestCreate{
					Name: "foo",
				},
			},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    ProjectInfo{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				settings: ProjectSettingsRequestCreate{
					Name: "foo",
				},
			},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    ProjectInfo{},
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
				got, err := c.CreateProject(tt.args.settings)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

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
		envVars map[string]string
		want    ProjectStatus
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
			},
			want: ProjectStatus{
				Action:        "stop_compute",
				CreatedAt:     mustParseTime("2022-07-26T08:56:59.366068468Z"),
				Error:         "",
				FailuresCount: 0,
				ID:            281749,
				ProjectID:     "validProjectID",
				Status:        "running",
				UpdatedAt:     mustParseTime("2022-07-26T08:56:59.366068468Z"),
				UUID:          "ab8421dd-fd07-4bf7-bd8f-4952f430eb1d",
			},
			wantErr: false,
		},
		{
			name: "unhappy path: valid projectID, internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
			},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    ProjectStatus{},
			wantErr: true,
		},
		{
			name: "unhappy path: projectID not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "invalidProjectID",
			},
			want:    ProjectStatus{},
			wantErr: true,
		},
		{
			name: "unhappy path: valid projectID, corrupted response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
			},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    ProjectStatus{},
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
				got, err := c.DeleteProject(tt.args.projectID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf(
						"DeleteProject() error = %v, wantErr %v; %s", err, tt.wantErr,
						os.Getenv("SDK_TEST_INTERNAL_ERROR"),
					)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjects(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	tests := []struct {
		name    string
		fields  fields
		envVars map[string]string
		want    []ProjectInfo
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			want: []ProjectInfo{
				{
					CreatedAt:    mustParseTime("2022-07-24T11:18:12.322513Z"),
					CurrentState: "active",
					Databases: []Database{
						{
							ID:      1,
							Name:    "main",
							OwnerID: 1,
						},
					},
					Deleted:        false,
					ID:             "quiet-river-711967",
					InstanceHandle: "",
					InstanceTypeID: "0",
					MaxProjectSize: 0,
					Name:           "foo",
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
					UpdatedAt: mustParseTime("2022-07-24T11:18:18.389868Z"),
				},
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
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
				got, err := c.ListProjects()

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjects() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjects() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ReadProject(t *testing.T) {
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
		envVars map[string]string
		want    ProjectInfo
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			want: ProjectInfo{
				CreatedAt:    mustParseTime("2022-07-24T11:18:12.322513Z"),
				CurrentState: "active",
				Databases: []Database{
					{
						ID:      1,
						Name:    "main",
						OwnerID: 1,
					},
				},
				Deleted:        false,
				ID:             "validProjectID",
				InstanceHandle: "",
				InstanceTypeID: "0",
				MaxProjectSize: 0,
				Name:           "foo",
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
				UpdatedAt: mustParseTime("2022-07-24T11:18:18.389868Z"),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    ProjectInfo{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    ProjectInfo{},
			wantErr: true,
		},
		{
			name: "unhappy path: project is not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "invalidProjectID"},
			want:    ProjectInfo{},
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
				got, err := c.ReadProject(tt.args.projectID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("ReadProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ReadProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_UpdateProject(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	type args struct {
		projectID string
		settings  ProjectSettingsRequestUpdate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    ProjectInfo
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
				settings: ProjectSettingsRequestUpdate{
					Project: struct {
						InstanceTypeId string                 `json:"instance_type_id"`
						Name           string                 `json:"name"`
						PoolerEnabled  bool                   `json:"pooler_enabled"`
						Settings       map[string]interface{} `json:"settings"`
					}{Name: "bar"},
				},
			},
			want: ProjectInfo{
				CreatedAt:    mustParseTime("2022-07-24T11:18:12.322513Z"),
				CurrentState: "active",
				Databases: []Database{
					{
						ID:      1,
						Name:    "main",
						OwnerID: 1,
					},
				},
				Deleted:        false,
				ID:             "validProjectID",
				InstanceHandle: "",
				InstanceTypeID: "0",
				MaxProjectSize: 0,
				Name:           "bar",
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
				UpdatedAt: mustParseTime("2022-07-24T11:18:18.389868Z"),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    ProjectInfo{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    ProjectInfo{},
			wantErr: true,
		},
		{
			name: "unhappy path: project is not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "invalidProjectID"},
			want:    ProjectInfo{},
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
				got, err := c.UpdateProject(tt.args.projectID, tt.args.settings)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UpdateProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_StartProject(t *testing.T) {
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
		envVars map[string]string
		want    ProjectStatus
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			want: ProjectStatus{
				Action:        "start_compute",
				CreatedAt:     mustParseTime("2022-07-26T08:56:59.366068468Z"),
				Error:         "",
				FailuresCount: 0,
				ID:            281749,
				ProjectID:     "validProjectID",
				RetryAt:       mustParseTime("0001-01-01T00:00:00Z"),
				Status:        "running",
				UpdatedAt:     mustParseTime("2022-07-26T08:56:59.366068468Z"),
				UUID:          "ab8421dd-fd07-4bf7-bd8f-4952f430eb1d",
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    ProjectStatus{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    ProjectStatus{},
			wantErr: true,
		},
		{
			name: "unhappy path: project is not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "invalidProjectID"},
			want:    ProjectStatus{},
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
				got, err := c.StartProject(tt.args.projectID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("StartProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("StartProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_StopProject(t *testing.T) {
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
		envVars map[string]string
		want    ProjectStatus
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			want: ProjectStatus{
				Action:        "stop_compute",
				CreatedAt:     mustParseTime("2022-07-26T08:56:59.366068468Z"),
				Error:         "",
				FailuresCount: 0,
				ID:            281749,
				ProjectID:     "validProjectID",
				RetryAt:       mustParseTime("0001-01-01T00:00:00Z"),
				Status:        "running",
				UpdatedAt:     mustParseTime("2022-07-26T08:56:59.366068468Z"),
				UUID:          "ab8421dd-fd07-4bf7-bd8f-4952f430eb1d",
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    ProjectStatus{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    ProjectStatus{},
			wantErr: true,
		},
		{
			name: "unhappy path: project is not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientProjects,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "invalidProjectID"},
			want:    ProjectStatus{},
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
				got, err := c.StopProject(tt.args.projectID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("StopProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("StopProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
