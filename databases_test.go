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

var mockHttpClientDatabases = &httpClientMock{
	m: map[string]map[reqType]resp{
		urlPrefix + "projects/validProjectID/databases": {
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
							`[
  {
    "created_at": "2022-07-27T09:13:06.855Z",
    "id": 0,
    "name": "main",
    "owner_id": 0,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
  }
]`,
						),
					),
					ContentLength: 1000,
				}, nil
			},

			// update project
			post: func(req *http.Request) (*http.Response, error) {
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
				var db DatabaseRequest
				if err := json.Unmarshal(b, &db); err != nil {
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
    "owner_id": %d,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
  }`,
								db.Database.Name, db.Database.OwnerID,
							),
						),
					),
				}, nil
			},
		},
		urlPrefix + "projects/invalidProjectID/databases": {
			get:  objNotFoundResponse,
			post: objNotFoundResponse,
		},

		urlPrefix + "projects/validProjectID/databases/0": {
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
    "created_at": "2022-07-27T09:13:06.855Z",
    "id": 0,
    "name": "main",
    "owner_id": 1,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
}`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
			put: func(req *http.Request) (*http.Response, error) {
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
				var db DatabaseRequest
				if err := json.Unmarshal(b, &db); err != nil {
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
    "owner_id": %d,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
  }`,
								db.Database.Name, db.Database.OwnerID,
							),
						),
					),
				}, nil
			},
			del: func(req *http.Request) (*http.Response, error) {
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
    "name": "main",
    "owner_id": 1,
    "project_id": "validProjectID",
    "updated_at": "2022-07-27T09:13:06.855Z"
}`,
						),
					),
					ContentLength: 1000,
				}, nil
			},
		},
		urlPrefix + "projects/validProjectID/databases/1": {
			get: objNotFoundResponse,
			put: objNotFoundResponse,
			del: objNotFoundResponse,
		},
	},
}

func Test_client_ListDatabases(t *testing.T) {
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
		want    []DatabaseResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			want: []DatabaseResponse{
				{
					CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
					ID:        0,
					Name:      "main",
					OwnerID:   0,
					ProjectID: "validProjectID",
					UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				},
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
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
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
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
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
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
				got, err := c.ListDatabases(tt.args.projectID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("ListDatabases() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListDatabases() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateDatabase(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	type args struct {
		projectID string
		cfg       DatabaseRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    DatabaseResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID: "validProjectID",
				cfg: DatabaseRequest{
					Database: struct {
						Name    string `json:"name"`
						OwnerID int    `json:"owner_id"`
					}{
						Name:    "main",
						OwnerID: 1,
					},
				},
			},
			want: DatabaseResponse{
				CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				ID:        0,
				Name:      "main",
				OwnerID:   1,
				ProjectID: "validProjectID",
				UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    DatabaseResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    DatabaseResponse{},
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
				got, err := c.CreateDatabase(tt.args.projectID, tt.args.cfg)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("CreateDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ReadDatabase(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	type args struct {
		projectID  string
		databaseID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    DatabaseResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID:  "validProjectID",
				databaseID: 0,
			},
			want: DatabaseResponse{
				CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				ID:        0,
				Name:      "main",
				OwnerID:   1,
				ProjectID: "validProjectID",
				UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    DatabaseResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    DatabaseResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: db not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", databaseID: 1},
			want:    DatabaseResponse{},
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
				got, err := c.ReadDatabase(tt.args.projectID, tt.args.databaseID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("ReadDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ReadDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_UpdateDatabase(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	type args struct {
		projectID  string
		databaseID int
		cfg        DatabaseRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    DatabaseResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID:  "validProjectID",
				databaseID: 0,
				cfg: DatabaseRequest{
					Database: struct {
						Name    string `json:"name"`
						OwnerID int    `json:"owner_id"`
					}{Name: "foo", OwnerID: 1},
				},
			},
			want: DatabaseResponse{
				CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				ID:        0,
				Name:      "foo",
				OwnerID:   1,
				ProjectID: "validProjectID",
				UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    DatabaseResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    DatabaseResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: db not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", databaseID: 1},
			want:    DatabaseResponse{},
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
				got, err := c.UpdateDatabase(tt.args.projectID, tt.args.databaseID, tt.args.cfg)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UpdateDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteDatabase(t *testing.T) {
	type fields struct {
		options Options
		baseURL string
	}
	type args struct {
		projectID  string
		databaseID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		envVars map[string]string
		want    DatabaseResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{
				projectID:  "validProjectID",
				databaseID: 0,
			},
			want: DatabaseResponse{
				CreatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
				ID:        0,
				Name:      "main",
				OwnerID:   1,
				ProjectID: "validProjectID",
				UpdatedAt: mustParseTime("2022-07-27T09:13:06.855Z"),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: internal error",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args: args{projectID: "validProjectID"},
			envVars: map[string]string{
				"SDK_TEST_INTERNAL_ERROR": "TRUE",
			},
			want:    DatabaseResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt response content",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID"},
			envVars: map[string]string{"SDK_TEST_RESPONSE_FAIL": "TRUE"},
			want:    DatabaseResponse{},
			wantErr: true,
		},
		{
			name: "unhappy path: db not found",
			fields: fields{
				options: Options{
					APIKey:     "validApiKey",
					HTTPClient: mockHttpClientDatabases,
				},
				baseURL: baseURL,
			},
			args:    args{projectID: "validProjectID", databaseID: 1},
			want:    DatabaseResponse{},
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
				got, err := c.DeleteDatabase(tt.args.projectID, tt.args.databaseID)

				t.Cleanup(
					func() {
						for k := range tt.envVars {
							_ = os.Unsetenv(k)
						}
					},
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
