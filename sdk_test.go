package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	type args struct {
		cfg Config
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "happy path, default http client",
			args: args{
				cfg: Config{
					Key: "foo",
				},
			},
			want: &Client{
				cfg: Config{
					Key:        "foo",
					HTTPClient: &http.Client{Timeout: defaultTimeout},
				},
				baseURL: baseURL,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: missing api key",
			args: args{
				cfg: Config{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "happy path: custom http client",
			args: args{
				cfg: Config{
					Key:        "bar",
					HTTPClient: &http.Client{Timeout: 1 * time.Minute},
				},
			},
			want: &Client{
				cfg: Config{
					Key:        "bar",
					HTTPClient: &http.Client{Timeout: 1 * time.Minute},
				},
				baseURL: baseURL,
			},
			wantErr: false,
		},
		{
			name: "happy path: custom http client and key",
			args: args{
				cfg: Config{
					Key:        "bar",
					HTTPClient: &http.Client{Timeout: 1 * time.Minute},
				},
			},
			want: &Client{
				cfg: Config{
					Key:        "bar",
					HTTPClient: &http.Client{Timeout: 1 * time.Minute},
				},
				baseURL: baseURL,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewClient(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewClient() got = %v, want %v", got, tt.want)
				}

			},
		)
	}
}

type mockPayload struct {
	Foo string `json:"foo"`
}

type mockHttp struct {
	reqHeaders http.Header
	respBody   mockPayload
	err        Error
}

func (m *mockHttp) Do(req *http.Request) (*http.Response, error) {
	m.reqHeaders = req.Header

	if m.err.HTTPCode > 299 {
		o, _ := json.Marshal(m.err.errorResp)
		return &http.Response{
			StatusCode: m.err.HTTPCode,
			Request:    req,
			Body:       io.NopCloser(bytes.NewReader(o)),
		}, nil
	}

	var (
		err error
		r   []byte
	)
	if req.Body != nil {
		buf, err := io.ReadAll(req.Body)
		defer func() { _ = req.Body.Close() }()
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(buf, &m.respBody); err != nil {
			return nil, err
		}
		m.respBody.Foo = "resp:" + strings.TrimPrefix(m.respBody.Foo, "req:")
	}

	if m.respBody.Foo != "" {
		r, err = json.Marshal(m.respBody)
		if err != nil {
			return nil, err
		}
	}

	return &http.Response{
		StatusCode:    m.err.HTTPCode,
		Body:          io.NopCloser(bytes.NewReader(r)),
		ContentLength: int64(len(r)),
		Request:       req,
	}, nil
}

func Test_client_requestHandler(t *testing.T) {
	type fields struct {
		cfg     Config
		baseURL string
	}
	type args struct {
		url             string
		t               string
		reqPayload      interface{}
		responsePayload interface{}
	}

	var respPayload mockPayload

	tests := []struct {
		name               string
		fields             fields
		args               args
		wantRequestHeaders http.Header
		wantResp           mockPayload
		wantErr            error
	}{
		{
			name: "happy path: post w payload",
			fields: fields{
				cfg: Config{
					Key: "foo",
					HTTPClient: &mockHttp{
						err: Error{HTTPCode: http.StatusOK},
					},
				},
				baseURL: "",
			},
			args: args{
				url:             "",
				t:               "POST",
				reqPayload:      mockPayload{Foo: "req:bar"},
				responsePayload: &respPayload,
			},
			wantRequestHeaders: http.Header{
				"Accept":        []string{"application/json"},
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer foo"},
			},
			wantResp: mockPayload{Foo: "resp:bar"},
			wantErr:  nil,
		},
		{
			name: "happy path: get w/o payload",
			fields: fields{
				cfg: Config{
					Key: "bar",
					HTTPClient: &mockHttp{
						err:      Error{HTTPCode: http.StatusOK},
						respBody: mockPayload{Foo: "resp:"},
					},
				},
				baseURL: "",
			},
			args: args{
				url:             "",
				t:               "GET",
				responsePayload: &respPayload,
			},
			wantRequestHeaders: http.Header{
				"Accept":        []string{"application/json"},
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer bar"},
			},
			wantResp: mockPayload{Foo: "resp:"},
			wantErr:  nil,
		},
		{
			name: "unhappy path: get w/o payload",
			fields: fields{
				cfg: Config{
					Key: "bar",
					HTTPClient: &mockHttp{
						err: Error{
							HTTPCode: http.StatusNotFound,
							errorResp: errorResp{
								Code:    "foo",
								Message: "bar",
							},
						},
					},
				},
				baseURL: "",
			},
			args: args{
				url:             "",
				t:               "GET",
				responsePayload: &respPayload,
			},
			wantRequestHeaders: http.Header{
				"Accept":        []string{"application/json"},
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer bar"},
			},
			wantResp: mockPayload{},
			wantErr: Error{
				HTTPCode: http.StatusNotFound,
				errorResp: errorResp{
					Code:    "foo",
					Message: "bar",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := &Client{
					cfg:     tt.fields.cfg,
					baseURL: tt.fields.baseURL,
				}
				respPayload = mockPayload{}

				if err := c.requestHandler(
					tt.args.url, tt.args.t, tt.args.reqPayload, tt.args.responsePayload,
				); err != tt.wantErr {
					t.Errorf("requestHandler() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !reflect.DeepEqual(tt.wantRequestHeaders, (tt.fields.cfg.HTTPClient).(*mockHttp).reqHeaders) {
					t.Errorf("missing expected request headers")
				}

				if !reflect.DeepEqual(tt.wantResp, respPayload) {
					t.Errorf("response payload does not match expectations")
				}

			},
		)
	}
}

type faultyReader struct{}

func (f faultyReader) Read(_ []byte) (n int, err error) {
	return 1, errors.New("foo")
}

func Test_convertErrorResponse(t *testing.T) {
	type args struct {
		res *http.Response
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy path: not found",
			args: args{
				res: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`{"code":"","message":"not found"}`)),
				},
			},
			wantErr: Error{
				HTTPCode: http.StatusNotFound,
				errorResp: errorResp{
					Message: "not found",
				},
			},
		},
		{
			name: "unhappy path: faulty body content",
			args: args{
				res: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(faultyReader{}),
				},
			},
			wantErr: Error{
				HTTPCode: http.StatusNotFound,
				errorResp: errorResp{
					Message: "cannot read response bytes",
				},
			},
		},
		{
			name: "unhappy path: faulty json",
			args: args{
				res: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`{`)),
				},
			},
			wantErr: Error{
				HTTPCode: http.StatusNotFound,
				errorResp: errorResp{
					Message: "unexpected end of JSON input",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := convertErrorResponse(tt.args.res); !reflect.DeepEqual(tt.wantErr, err) {
					t.Errorf("convertErrorResponse() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestError_Error(t *testing.T) {
	type fields struct {
		HTTPCode  int
		errorResp errorResp
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "404",
			fields: fields{
				HTTPCode: http.StatusNotFound,
				errorResp: errorResp{
					Message: "object not found",
				},
			},
			want: "[HTTP Code: 404][Error Code: ] object not found",
		},
		{
			name: "406",
			fields: fields{
				HTTPCode: http.StatusNotAcceptable,
				errorResp: errorResp{
					Code:    "foo",
					Message: "bar",
				},
			},
			want: "[HTTP Code: 406][Error Code: foo] bar",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := Error{
					HTTPCode:  tt.fields.HTTPCode,
					errorResp: tt.fields.errorResp,
				}
				if got := e.Error(); got != tt.want {
					t.Errorf("Error() = %s, want %s", got, tt.want)
				}
			},
		)
	}
}

func TestError_httpResp(t *testing.T) {
	type fields struct {
		HTTPCode  int
		errorResp errorResp
	}
	tests := []struct {
		name   string
		fields fields
		want   *http.Response
	}{
		{
			name: "",
			fields: fields{
				HTTPCode: http.StatusNotFound,
				errorResp: errorResp{
					Code:    "foo",
					Message: "object not found",
				},
			},
			want: &http.Response{
				Status:        "foo",
				StatusCode:    http.StatusNotFound,
				Body:          io.NopCloser(bytes.NewReader([]byte(`{"code":"foo","message":"object not found"}`))),
				ContentLength: int64(len(`{"code":"foo","message":"object not found"}`)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := Error{
					HTTPCode:  tt.fields.HTTPCode,
					errorResp: tt.fields.errorResp,
				}
				if got := e.httpResp(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("httpResp() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListApiKeys(t *testing.T) {
	deserializeResp := func(s string) []ApiKeysListResponseItem {
		var v []ApiKeysListResponseItem
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	tests := []struct {
		name    string
		apiKey  string
		want    []ApiKeysListResponseItem
		wantErr bool
	}{
		{
			name:    "happy path",
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/api_keys"]["GET"].Content),
			wantErr: false,
		},
		{
			name:    "unhappy path",
			apiKey:  "invalidApiKey",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListApiKeys()
				if (err != nil) != tt.wantErr {
					t.Errorf("ListApiKeys() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListApiKeys() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateApiKey(t *testing.T) {
	deserializeResp := func(s string) ApiKeyCreateResponse {
		var v ApiKeyCreateResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cfg ApiKeyCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ApiKeyCreateResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: ApiKeyCreateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/api_keys"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cfg: ApiKeyCreateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    ApiKeyCreateResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.CreateApiKey(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateApiKey() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateApiKey() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_RevokeApiKey(t *testing.T) {
	deserializeResp := func(s string) ApiKeyRevokeResponse {
		var v ApiKeyRevokeResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		keyID int64
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ApiKeyRevokeResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				keyID: 1,
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/api_keys/{key_id}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				keyID: 1,
			},
			apiKey:  "invalidApiKey",
			want:    ApiKeyRevokeResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.RevokeApiKey(tt.args.keyID)
				if (err != nil) != tt.wantErr {
					t.Errorf("RevokeApiKey() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RevokeApiKey() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProjectOperation(t *testing.T) {
	deserializeResp := func(s string) OperationResponse {
		var v OperationResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID   string
		operationID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    OperationResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:   "foo",
				operationID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/operations/{operation_id}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:   "foo",
				operationID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    OperationResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetProjectOperation(tt.args.projectID, tt.args.operationID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectOperation() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectOperation() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjects(t *testing.T) {
	deserializeResp := func(s string) ListProjectsRespObj {
		var v ListProjectsRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cursor *string
		limit  *int
		search *string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ListProjectsRespObj
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cursor: createPointer("foo"),
				limit:  createPointer(1),
				search: createPointer("foo"),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cursor: createPointer("foo"),
				limit:  createPointer(1),
				search: createPointer("foo"),
			},
			apiKey:  "invalidApiKey",
			want:    ListProjectsRespObj{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjects(tt.args.cursor, tt.args.limit, tt.args.search)
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

func Test_client_CreateProject(t *testing.T) {
	deserializeResp := func(s string) CreatedProject {
		var v CreatedProject
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cfg ProjectCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    CreatedProject
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: ProjectCreateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cfg: ProjectCreateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    CreatedProject{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.CreateProject(tt.args.cfg)
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

func Test_client_ListSharedProjects(t *testing.T) {
	deserializeResp := func(s string) ListSharedProjectsRespObj {
		var v ListSharedProjectsRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cursor *string
		limit  *int
		search *string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ListSharedProjectsRespObj
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cursor: createPointer("foo"),
				limit:  createPointer(1),
				search: createPointer("foo"),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/shared"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cursor: createPointer("foo"),
				limit:  createPointer(1),
				search: createPointer("foo"),
			},
			apiKey:  "invalidApiKey",
			want:    ListSharedProjectsRespObj{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListSharedProjects(tt.args.cursor, tt.args.limit, tt.args.search)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListSharedProjects() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListSharedProjects() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProject(t *testing.T) {
	deserializeResp := func(s string) ProjectResponse {
		var v ProjectResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ProjectResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    ProjectResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetProject(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_UpdateProject(t *testing.T) {
	deserializeResp := func(s string) UpdateProjectRespObj {
		var v UpdateProjectRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		cfg       ProjectUpdateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    UpdateProjectRespObj
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				cfg:       ProjectUpdateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}"]["PATCH"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				cfg:       ProjectUpdateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    UpdateProjectRespObj{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.UpdateProject(tt.args.projectID, tt.args.cfg)
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

func Test_client_DeleteProject(t *testing.T) {
	deserializeResp := func(s string) ProjectResponse {
		var v ProjectResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ProjectResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    ProjectResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
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

func Test_client_ListProjectOperations(t *testing.T) {
	deserializeResp := func(s string) ListOperations {
		var v ListOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		cursor    *string
		limit     *int
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ListOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				cursor:    createPointer("foo"),
				limit:     createPointer(1),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/operations"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				cursor:    createPointer("foo"),
				limit:     createPointer(1),
			},
			apiKey:  "invalidApiKey",
			want:    ListOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjectOperations(tt.args.projectID, tt.args.cursor, tt.args.limit)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectOperations() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectOperations() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjectBranches(t *testing.T) {
	deserializeResp := func(s string) BranchesResponse {
		var v BranchesResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchesResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    BranchesResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjectBranches(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectBranches() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectBranches() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateProjectBranch(t *testing.T) {
	deserializeResp := func(s string) CreatedBranch {
		var v CreatedBranch
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		cfg       *BranchCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    CreatedBranch
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				cfg:       nil,
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				cfg:       nil,
			},
			apiKey:  "invalidApiKey",
			want:    CreatedBranch{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.CreateProjectBranch(tt.args.projectID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateProjectBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateProjectBranch() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProjectBranch(t *testing.T) {
	deserializeResp := func(s string) BranchResponse {
		var v BranchResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    BranchResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetProjectBranch(tt.args.projectID, tt.args.branchID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectBranch() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_UpdateProjectBranch(t *testing.T) {
	deserializeResp := func(s string) BranchOperations {
		var v BranchOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		cfg       BranchUpdateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       BranchUpdateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}"]["PATCH"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       BranchUpdateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    BranchOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.UpdateProjectBranch(tt.args.projectID, tt.args.branchID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateProjectBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UpdateProjectBranch() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteProjectBranch(t *testing.T) {
	deserializeResp := func(s string) BranchOperations {
		var v BranchOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    BranchOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.DeleteProjectBranch(tt.args.projectID, tt.args.branchID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteProjectBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteProjectBranch() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_RestoreProjectBranch(t *testing.T) {
	deserializeResp := func(s string) BranchOperations {
		var v BranchOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		cfg       BranchRestoreRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       BranchRestoreRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/restore"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       BranchRestoreRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    BranchOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.RestoreProjectBranch(tt.args.projectID, tt.args.branchID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("RestoreProjectBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RestoreProjectBranch() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_SetPrimaryProjectBranch(t *testing.T) {
	deserializeResp := func(s string) BranchOperations {
		var v BranchOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/set_as_primary"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    BranchOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.SetPrimaryProjectBranch(tt.args.projectID, tt.args.branchID)
				if (err != nil) != tt.wantErr {
					t.Errorf("SetPrimaryProjectBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("SetPrimaryProjectBranch() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjectBranchEndpoints(t *testing.T) {
	deserializeResp := func(s string) EndpointsResponse {
		var v EndpointsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/endpoints"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    EndpointsResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjectBranchEndpoints(tt.args.projectID, tt.args.branchID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectBranchEndpoints() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectBranchEndpoints() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjectBranchDatabases(t *testing.T) {
	deserializeResp := func(s string) DatabasesResponse {
		var v DatabasesResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    DatabasesResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/databases"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    DatabasesResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjectBranchDatabases(tt.args.projectID, tt.args.branchID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectBranchDatabases() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectBranchDatabases() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateProjectBranchDatabase(t *testing.T) {
	deserializeResp := func(s string) DatabaseOperations {
		var v DatabaseOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		cfg       DatabaseCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    DatabaseOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       DatabaseCreateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/databases"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       DatabaseCreateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    DatabaseOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.CreateProjectBranchDatabase(tt.args.projectID, tt.args.branchID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateProjectBranchDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateProjectBranchDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProjectBranchDatabase(t *testing.T) {
	deserializeResp := func(s string) DatabaseResponse {
		var v DatabaseResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID    string
		branchID     string
		databaseName string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    DatabaseResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:    "foo",
				branchID:     "foo",
				databaseName: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/databases/{database_name}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:    "foo",
				branchID:     "foo",
				databaseName: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    DatabaseResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetProjectBranchDatabase(tt.args.projectID, tt.args.branchID, tt.args.databaseName)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectBranchDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectBranchDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_UpdateProjectBranchDatabase(t *testing.T) {
	deserializeResp := func(s string) DatabaseOperations {
		var v DatabaseOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID    string
		branchID     string
		databaseName string
		cfg          DatabaseUpdateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    DatabaseOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:    "foo",
				branchID:     "foo",
				databaseName: "foo",
				cfg:          DatabaseUpdateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/databases/{database_name}"]["PATCH"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:    "foo",
				branchID:     "foo",
				databaseName: "foo",
				cfg:          DatabaseUpdateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    DatabaseOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.UpdateProjectBranchDatabase(tt.args.projectID, tt.args.branchID, tt.args.databaseName, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateProjectBranchDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UpdateProjectBranchDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteProjectBranchDatabase(t *testing.T) {
	deserializeResp := func(s string) DatabaseOperations {
		var v DatabaseOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID    string
		branchID     string
		databaseName string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    DatabaseOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:    "foo",
				branchID:     "foo",
				databaseName: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/databases/{database_name}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:    "foo",
				branchID:     "foo",
				databaseName: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    DatabaseOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.DeleteProjectBranchDatabase(tt.args.projectID, tt.args.branchID, tt.args.databaseName)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteProjectBranchDatabase() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteProjectBranchDatabase() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjectBranchRoles(t *testing.T) {
	deserializeResp := func(s string) RolesResponse {
		var v RolesResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    RolesResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/roles"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    RolesResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjectBranchRoles(tt.args.projectID, tt.args.branchID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectBranchRoles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectBranchRoles() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateProjectBranchRole(t *testing.T) {
	deserializeResp := func(s string) RoleOperations {
		var v RoleOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		cfg       RoleCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    RoleOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       RoleCreateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/roles"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				cfg:       RoleCreateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    RoleOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.CreateProjectBranchRole(tt.args.projectID, tt.args.branchID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateProjectBranchRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateProjectBranchRole() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProjectBranchRole(t *testing.T) {
	deserializeResp := func(s string) RoleResponse {
		var v RoleResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		roleName  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    RoleResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/roles/{role_name}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    RoleResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetProjectBranchRole(tt.args.projectID, tt.args.branchID, tt.args.roleName)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectBranchRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectBranchRole() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteProjectBranchRole(t *testing.T) {
	deserializeResp := func(s string) RoleOperations {
		var v RoleOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		roleName  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    RoleOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/roles/{role_name}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    RoleOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.DeleteProjectBranchRole(tt.args.projectID, tt.args.branchID, tt.args.roleName)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteProjectBranchRole() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteProjectBranchRole() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProjectBranchRolePassword(t *testing.T) {
	deserializeResp := func(s string) RolePasswordResponse {
		var v RolePasswordResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		roleName  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    RolePasswordResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/roles/{role_name}/reveal_password"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    RolePasswordResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetProjectBranchRolePassword(tt.args.projectID, tt.args.branchID, tt.args.roleName)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectBranchRolePassword() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectBranchRolePassword() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ResetProjectBranchRolePassword(t *testing.T) {
	deserializeResp := func(s string) RoleOperations {
		var v RoleOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		roleName  string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    RoleOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/roles/{role_name}/reset_password"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				roleName:  "foo",
			},
			apiKey:  "invalidApiKey",
			want:    RoleOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ResetProjectBranchRolePassword(tt.args.projectID, tt.args.branchID, tt.args.roleName)
				if (err != nil) != tt.wantErr {
					t.Errorf("ResetProjectBranchRolePassword() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ResetProjectBranchRolePassword() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjectEndpoints(t *testing.T) {
	deserializeResp := func(s string) EndpointsResponse {
		var v EndpointsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    EndpointsResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjectEndpoints(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectEndpoints() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectEndpoints() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateProjectEndpoint(t *testing.T) {
	deserializeResp := func(s string) EndpointOperations {
		var v EndpointOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		cfg       EndpointCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				cfg:       EndpointCreateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				cfg:       EndpointCreateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    EndpointOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.CreateProjectEndpoint(tt.args.projectID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateProjectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateProjectEndpoint() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProjectEndpoint(t *testing.T) {
	deserializeResp := func(s string) EndpointResponse {
		var v EndpointResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID  string
		endpointID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints/{endpoint_id}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    EndpointResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetProjectEndpoint(tt.args.projectID, tt.args.endpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectEndpoint() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_UpdateProjectEndpoint(t *testing.T) {
	deserializeResp := func(s string) EndpointOperations {
		var v EndpointOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID  string
		endpointID string
		cfg        EndpointUpdateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
				cfg:        EndpointUpdateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints/{endpoint_id}"]["PATCH"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
				cfg:        EndpointUpdateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    EndpointOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.UpdateProjectEndpoint(tt.args.projectID, tt.args.endpointID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateProjectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UpdateProjectEndpoint() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteProjectEndpoint(t *testing.T) {
	deserializeResp := func(s string) EndpointOperations {
		var v EndpointOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID  string
		endpointID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints/{endpoint_id}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    EndpointOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.DeleteProjectEndpoint(tt.args.projectID, tt.args.endpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteProjectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteProjectEndpoint() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_StartProjectEndpoint(t *testing.T) {
	deserializeResp := func(s string) EndpointOperations {
		var v EndpointOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID  string
		endpointID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints/{endpoint_id}/start"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    EndpointOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.StartProjectEndpoint(tt.args.projectID, tt.args.endpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("StartProjectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("StartProjectEndpoint() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_SuspendProjectEndpoint(t *testing.T) {
	deserializeResp := func(s string) EndpointOperations {
		var v EndpointOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID  string
		endpointID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EndpointOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints/{endpoint_id}/suspend"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:  "foo",
				endpointID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    EndpointOperations{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.SuspendProjectEndpoint(tt.args.projectID, tt.args.endpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("SuspendProjectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("SuspendProjectEndpoint() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListProjectsConsumption(t *testing.T) {
	deserializeResp := func(s string) ListProjectsConsumptionRespObj {
		var v ListProjectsConsumptionRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cursor *string
		limit  *int
		from   *time.Time
		to     *time.Time
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ListProjectsConsumptionRespObj
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cursor: createPointer("foo"),
				limit:  createPointer(1),
				from:   createPointer(time.Time{}),
				to:     createPointer(time.Time{}),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/consumption/projects"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cursor: createPointer("foo"),
				limit:  createPointer(1),
				from:   createPointer(time.Time{}),
				to:     createPointer(time.Time{}),
			},
			apiKey:  "invalidApiKey",
			want:    ListProjectsConsumptionRespObj{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.ListProjectsConsumption(tt.args.cursor, tt.args.limit, tt.args.from, tt.args.to)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectsConsumption() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectsConsumption() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetCurrentUserInfo(t *testing.T) {
	deserializeResp := func(s string) CurrentUserInfoResponse {
		var v CurrentUserInfoResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	tests := []struct {
		name    string
		apiKey  string
		want    CurrentUserInfoResponse
		wantErr bool
	}{
		{
			name:    "happy path",
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/users/me"]["GET"].Content),
			wantErr: false,
		},
		{
			name:    "unhappy path",
			apiKey:  "invalidApiKey",
			want:    CurrentUserInfoResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewClient(Config{tt.apiKey, NewMockHTTPClient()})
				if err != nil {
					panic(err)
				}
				got, err := c.GetCurrentUserInfo()
				if (err != nil) != tt.wantErr {
					t.Errorf("GetCurrentUserInfo() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetCurrentUserInfo() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestTypes(t *testing.T) {
	// GIVEN
	// the types are defined correctly

	// WHEN
	// check fields' types

	// THEN
	// optional fields of the complex types
	// are expected to be pointers to the underlying structs
	if reflect.TypeOf(EndpointCreateRequestEndpoint{}.Settings).Kind() != reflect.Ptr {
		t.Errorf("EndpointCreateRequestEndpoint{}.Settings must be pointer")
	}

	if reflect.TypeOf(EndpointUpdateRequestEndpoint{}.Settings).Kind() != reflect.Ptr {
		t.Errorf("EndpointUpdateRequestEndpoint{}.Settings must be pointer")
	}
}

type dummyType interface {
	int | int64 | int32 | bool | string | float64 | float32 | time.Time
}

func createPointer[V dummyType](v V) *V {
	return &v
}
