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

func Test_client_AddProjectJWKS(t *testing.T) {
	deserializeResp := func(s string) JWKSCreationOperation {
		var v JWKSCreationOperation
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		cfg       AddProjectJWKSRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    JWKSCreationOperation
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				cfg:       AddProjectJWKSRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/jwks"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				cfg:       AddProjectJWKSRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    JWKSCreationOperation{},
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
				got, err := c.AddProjectJWKS(tt.args.projectID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("AddProjectJWKS() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("AddProjectJWKS() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_AssignOrganizationVPCEndpoint(t *testing.T) {
	type args struct {
		orgID         string
		regionID      string
		vpcEndpointID string
		cfg           VPCEndpointAssignment
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID:         "foo",
				regionID:      "foo",
				vpcEndpointID: "foo",
				cfg:           VPCEndpointAssignment{},
			},
			apiKey:  "foo",
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID:         "foo",
				regionID:      "foo",
				vpcEndpointID: "foo",
				cfg:           VPCEndpointAssignment{},
			},
			apiKey:  "invalidApiKey",
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
				err = c.AssignOrganizationVPCEndpoint(tt.args.orgID, tt.args.regionID, tt.args.vpcEndpointID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("AssignOrganizationVPCEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}

func Test_client_AssignProjectVPCEndpoint(t *testing.T) {
	type args struct {
		projectID     string
		vpcEndpointID string
		cfg           VPCEndpointAssignment
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:     "foo",
				vpcEndpointID: "foo",
				cfg:           VPCEndpointAssignment{},
			},
			apiKey:  "foo",
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:     "foo",
				vpcEndpointID: "foo",
				cfg:           VPCEndpointAssignment{},
			},
			apiKey:  "invalidApiKey",
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
				err = c.AssignProjectVPCEndpoint(tt.args.projectID, tt.args.vpcEndpointID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("AssignProjectVPCEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}

func Test_client_CountProjectBranches(t *testing.T) {
	deserializeResp := func(s string) CountProjectBranchesRespObj {
		var v CountProjectBranchesRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		search    *string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    CountProjectBranchesRespObj
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				search:    createPointer("foo"),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/count"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				search:    createPointer("foo"),
			},
			apiKey:  "invalidApiKey",
			want:    CountProjectBranchesRespObj{},
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
				got, err := c.CountProjectBranches(tt.args.projectID, tt.args.search)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountProjectBranches() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CountProjectBranches() got = %v, want %v", got, tt.want)
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

func Test_client_CreateNeonAuthIntegration(t *testing.T) {
	deserializeResp := func(s string) NeonAuthCreateIntegrationResponse {
		var v NeonAuthCreateIntegrationResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cfg NeonAuthCreateIntegrationRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    NeonAuthCreateIntegrationResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: NeonAuthCreateIntegrationRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/auth/create"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cfg: NeonAuthCreateIntegrationRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    NeonAuthCreateIntegrationResponse{},
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
				got, err := c.CreateNeonAuthIntegration(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateNeonAuthIntegration() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateNeonAuthIntegration() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateNeonAuthNewUser(t *testing.T) {
	deserializeResp := func(s string) NeonAuthCreateNewUserResponse {
		var v NeonAuthCreateNewUserResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cfg NeonAuthCreateNewUserRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    NeonAuthCreateNewUserResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: NeonAuthCreateNewUserRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/auth/user"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cfg: NeonAuthCreateNewUserRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    NeonAuthCreateNewUserResponse{},
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
				got, err := c.CreateNeonAuthNewUser(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateNeonAuthNewUser() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateNeonAuthNewUser() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateNeonAuthProviderSDKKeys(t *testing.T) {
	deserializeResp := func(s string) NeonAuthCreateIntegrationResponse {
		var v NeonAuthCreateIntegrationResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cfg NeonAuthCreateAuthProviderSDKKeysRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    NeonAuthCreateIntegrationResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: NeonAuthCreateAuthProviderSDKKeysRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/auth/keys"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cfg: NeonAuthCreateAuthProviderSDKKeysRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    NeonAuthCreateIntegrationResponse{},
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
				got, err := c.CreateNeonAuthProviderSDKKeys(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateNeonAuthProviderSDKKeys() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateNeonAuthProviderSDKKeys() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateOrgApiKey(t *testing.T) {
	deserializeResp := func(s string) OrgApiKeyCreateResponse {
		var v OrgApiKeyCreateResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID string
		cfg   OrgApiKeyCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    OrgApiKeyCreateResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID: "foo",
				cfg:   OrgApiKeyCreateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/api_keys"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID: "foo",
				cfg:   OrgApiKeyCreateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    OrgApiKeyCreateResponse{},
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
				got, err := c.CreateOrgApiKey(tt.args.orgID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateOrgApiKey() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateOrgApiKey() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_CreateOrganizationInvitations(t *testing.T) {
	deserializeResp := func(s string) OrganizationInvitationsResponse {
		var v OrganizationInvitationsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID string
		cfg   OrganizationInvitesCreateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    OrganizationInvitationsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID: "foo",
				cfg:   OrganizationInvitesCreateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/invitations"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID: "foo",
				cfg:   OrganizationInvitesCreateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    OrganizationInvitationsResponse{},
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
				got, err := c.CreateOrganizationInvitations(tt.args.orgID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateOrganizationInvitations() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("CreateOrganizationInvitations() got = %v, want %v", got, tt.want)
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
		cfg       *CreateProjectBranchReqObj
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

func Test_client_DeleteNeonAuthIntegration(t *testing.T) {
	type args struct {
		projectID    string
		authProvider NeonAuthSupportedAuthProvider
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:    "foo",
				authProvider: "foo",
			},
			apiKey:  "foo",
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:    "foo",
				authProvider: "foo",
			},
			apiKey:  "invalidApiKey",
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
				err = c.DeleteNeonAuthIntegration(tt.args.projectID, tt.args.authProvider)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteNeonAuthIntegration() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}

func Test_client_DeleteOrganizationVPCEndpoint(t *testing.T) {
	type args struct {
		orgID         string
		regionID      string
		vpcEndpointID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID:         "foo",
				regionID:      "foo",
				vpcEndpointID: "foo",
			},
			apiKey:  "foo",
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID:         "foo",
				regionID:      "foo",
				vpcEndpointID: "foo",
			},
			apiKey:  "invalidApiKey",
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
				err = c.DeleteOrganizationVPCEndpoint(tt.args.orgID, tt.args.regionID, tt.args.vpcEndpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteOrganizationVPCEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
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

func Test_client_DeleteProjectJWKS(t *testing.T) {
	deserializeResp := func(s string) JWKS {
		var v JWKS
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		jwksID    string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    JWKS
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				jwksID:    "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/jwks/{jwks_id}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				jwksID:    "foo",
			},
			apiKey:  "invalidApiKey",
			want:    JWKS{},
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
				got, err := c.DeleteProjectJWKS(tt.args.projectID, tt.args.jwksID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteProjectJWKS() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DeleteProjectJWKS() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_DeleteProjectVPCEndpoint(t *testing.T) {
	type args struct {
		projectID     string
		vpcEndpointID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:     "foo",
				vpcEndpointID: "foo",
			},
			apiKey:  "foo",
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:     "foo",
				vpcEndpointID: "foo",
			},
			apiKey:  "invalidApiKey",
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
				err = c.DeleteProjectVPCEndpoint(tt.args.projectID, tt.args.vpcEndpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteProjectVPCEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}

func Test_client_GetActiveRegions(t *testing.T) {
	deserializeResp := func(s string) ActiveRegionsResponse {
		var v ActiveRegionsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	tests := []struct {
		name    string
		apiKey  string
		want    ActiveRegionsResponse
		wantErr bool
	}{
		{
			name:    "happy path",
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/regions"]["GET"].Content),
			wantErr: false,
		},
		{
			name:    "unhappy path",
			apiKey:  "invalidApiKey",
			want:    ActiveRegionsResponse{},
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
				got, err := c.GetActiveRegions()
				if (err != nil) != tt.wantErr {
					t.Errorf("GetActiveRegions() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetActiveRegions() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetAuthDetails(t *testing.T) {
	deserializeResp := func(s string) AuthDetailsResponse {
		var v AuthDetailsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	tests := []struct {
		name    string
		apiKey  string
		want    AuthDetailsResponse
		wantErr bool
	}{
		{
			name:    "happy path",
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/auth"]["GET"].Content),
			wantErr: false,
		},
		{
			name:    "unhappy path",
			apiKey:  "invalidApiKey",
			want:    AuthDetailsResponse{},
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
				got, err := c.GetAuthDetails()
				if (err != nil) != tt.wantErr {
					t.Errorf("GetAuthDetails() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetAuthDetails() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetAvailablePreloadLibraries(t *testing.T) {
	deserializeResp := func(s string) AvailablePreloadLibraries {
		var v AvailablePreloadLibraries
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
		want    AvailablePreloadLibraries
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/available_preload_libraries"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    AvailablePreloadLibraries{},
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
				got, err := c.GetAvailablePreloadLibraries(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetAvailablePreloadLibraries() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetAvailablePreloadLibraries() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetConnectionURI(t *testing.T) {
	deserializeResp := func(s string) ConnectionURIResponse {
		var v ConnectionURIResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID    string
		branchID     *string
		endpointID   *string
		databaseName string
		roleName     string
		pooled       *bool
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ConnectionURIResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:    "foo",
				branchID:     createPointer("foo"),
				endpointID:   createPointer("foo"),
				databaseName: "foo",
				roleName:     "foo",
				pooled:       createPointer(true),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/connection_uri"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:    "foo",
				branchID:     createPointer("foo"),
				endpointID:   createPointer("foo"),
				databaseName: "foo",
				roleName:     "foo",
				pooled:       createPointer(true),
			},
			apiKey:  "invalidApiKey",
			want:    ConnectionURIResponse{},
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
				got, err := c.GetConnectionURI(tt.args.projectID, tt.args.branchID, tt.args.endpointID, tt.args.databaseName, tt.args.roleName, tt.args.pooled)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetConnectionURI() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetConnectionURI() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetConsumptionHistoryPerAccount(t *testing.T) {
	deserializeResp := func(s string) ConsumptionHistoryPerAccountResponse {
		var v ConsumptionHistoryPerAccountResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		from             time.Time
		to               time.Time
		granularity      ConsumptionHistoryGranularity
		orgID            *string
		includeV1Metrics *bool
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ConsumptionHistoryPerAccountResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				from:             time.Time{},
				to:               time.Time{},
				granularity:      "foo",
				orgID:            createPointer("foo"),
				includeV1Metrics: createPointer(true),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/consumption_history/account"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				from:             time.Time{},
				to:               time.Time{},
				granularity:      "foo",
				orgID:            createPointer("foo"),
				includeV1Metrics: createPointer(true),
			},
			apiKey:  "invalidApiKey",
			want:    ConsumptionHistoryPerAccountResponse{},
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
				got, err := c.GetConsumptionHistoryPerAccount(tt.args.from, tt.args.to, tt.args.granularity, tt.args.orgID, tt.args.includeV1Metrics)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetConsumptionHistoryPerAccount() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetConsumptionHistoryPerAccount() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetConsumptionHistoryPerProject(t *testing.T) {
	deserializeResp := func(s string) GetConsumptionHistoryPerProjectRespObj {
		var v GetConsumptionHistoryPerProjectRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cursor           *string
		limit            *int
		projectIDs       []string
		from             time.Time
		to               time.Time
		granularity      ConsumptionHistoryGranularity
		orgID            *string
		includeV1Metrics *bool
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    GetConsumptionHistoryPerProjectRespObj
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cursor:           createPointer("foo"),
				limit:            createPointer(1),
				projectIDs:       []string{"foo"},
				from:             time.Time{},
				to:               time.Time{},
				granularity:      "foo",
				orgID:            createPointer("foo"),
				includeV1Metrics: createPointer(true),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/consumption_history/projects"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cursor:           createPointer("foo"),
				limit:            createPointer(1),
				projectIDs:       []string{"foo"},
				from:             time.Time{},
				to:               time.Time{},
				granularity:      "foo",
				orgID:            createPointer("foo"),
				includeV1Metrics: createPointer(true),
			},
			apiKey:  "invalidApiKey",
			want:    GetConsumptionHistoryPerProjectRespObj{},
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
				got, err := c.GetConsumptionHistoryPerProject(tt.args.cursor, tt.args.limit, tt.args.projectIDs, tt.args.from, tt.args.to, tt.args.granularity, tt.args.orgID, tt.args.includeV1Metrics)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetConsumptionHistoryPerProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetConsumptionHistoryPerProject() got = %v, want %v", got, tt.want)
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

func Test_client_GetCurrentUserOrganizations(t *testing.T) {
	deserializeResp := func(s string) OrganizationsResponse {
		var v OrganizationsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	tests := []struct {
		name    string
		apiKey  string
		want    OrganizationsResponse
		wantErr bool
	}{
		{
			name:    "happy path",
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/users/me/organizations"]["GET"].Content),
			wantErr: false,
		},
		{
			name:    "unhappy path",
			apiKey:  "invalidApiKey",
			want:    OrganizationsResponse{},
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
				got, err := c.GetCurrentUserOrganizations()
				if (err != nil) != tt.wantErr {
					t.Errorf("GetCurrentUserOrganizations() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetCurrentUserOrganizations() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetOrganization(t *testing.T) {
	deserializeResp := func(s string) Organization {
		var v Organization
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    Organization
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    Organization{},
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
				got, err := c.GetOrganization(tt.args.orgID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetOrganization() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetOrganization() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetOrganizationInvitations(t *testing.T) {
	deserializeResp := func(s string) OrganizationInvitationsResponse {
		var v OrganizationInvitationsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    OrganizationInvitationsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/invitations"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    OrganizationInvitationsResponse{},
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
				got, err := c.GetOrganizationInvitations(tt.args.orgID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetOrganizationInvitations() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetOrganizationInvitations() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetOrganizationMember(t *testing.T) {
	deserializeResp := func(s string) Member {
		var v Member
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID    string
		memberID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    Member
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID:    "foo",
				memberID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/members/{member_id}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID:    "foo",
				memberID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    Member{},
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
				got, err := c.GetOrganizationMember(tt.args.orgID, tt.args.memberID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetOrganizationMember() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetOrganizationMember() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetOrganizationMembers(t *testing.T) {
	deserializeResp := func(s string) OrganizationMembersResponse {
		var v OrganizationMembersResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    OrganizationMembersResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/members"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    OrganizationMembersResponse{},
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
				got, err := c.GetOrganizationMembers(tt.args.orgID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetOrganizationMembers() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetOrganizationMembers() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetOrganizationVPCEndpointDetails(t *testing.T) {
	deserializeResp := func(s string) VPCEndpointDetails {
		var v VPCEndpointDetails
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID         string
		regionID      string
		vpcEndpointID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    VPCEndpointDetails
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID:         "foo",
				regionID:      "foo",
				vpcEndpointID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/vpc/region/{region_id}/vpc_endpoints/{vpc_endpoint_id}"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID:         "foo",
				regionID:      "foo",
				vpcEndpointID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    VPCEndpointDetails{},
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
				got, err := c.GetOrganizationVPCEndpointDetails(tt.args.orgID, tt.args.regionID, tt.args.vpcEndpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetOrganizationVPCEndpointDetails() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetOrganizationVPCEndpointDetails() got = %v, want %v", got, tt.want)
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

func Test_client_GetProjectBranch(t *testing.T) {
	deserializeResp := func(s string) GetProjectBranchRespObj {
		var v GetProjectBranchRespObj
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
		want    GetProjectBranchRespObj
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
			want:    GetProjectBranchRespObj{},
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

func Test_client_GetProjectBranchSchema(t *testing.T) {
	deserializeResp := func(s string) BranchSchemaResponse {
		var v BranchSchemaResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		branchID  string
		dbName    string
		lsn       *string
		timestamp *time.Time
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchSchemaResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				dbName:    "foo",
				lsn:       createPointer("foo"),
				timestamp: createPointer(time.Time{}),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/schema"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				branchID:  "foo",
				dbName:    "foo",
				lsn:       createPointer("foo"),
				timestamp: createPointer(time.Time{}),
			},
			apiKey:  "invalidApiKey",
			want:    BranchSchemaResponse{},
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
				got, err := c.GetProjectBranchSchema(tt.args.projectID, tt.args.branchID, tt.args.dbName, tt.args.lsn, tt.args.timestamp)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectBranchSchema() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectBranchSchema() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_GetProjectBranchSchemaComparison(t *testing.T) {
	deserializeResp := func(s string) BranchSchemaCompareResponse {
		var v BranchSchemaCompareResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID     string
		branchID      string
		baseBranchID  *string
		dbName        string
		lsn           *string
		timestamp     *time.Time
		baseLsn       *string
		baseTimestamp *time.Time
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    BranchSchemaCompareResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID:     "foo",
				branchID:      "foo",
				baseBranchID:  createPointer("foo"),
				dbName:        "foo",
				lsn:           createPointer("foo"),
				timestamp:     createPointer(time.Time{}),
				baseLsn:       createPointer("foo"),
				baseTimestamp: createPointer(time.Time{}),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/compare_schema"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID:     "foo",
				branchID:      "foo",
				baseBranchID:  createPointer("foo"),
				dbName:        "foo",
				lsn:           createPointer("foo"),
				timestamp:     createPointer(time.Time{}),
				baseLsn:       createPointer("foo"),
				baseTimestamp: createPointer(time.Time{}),
			},
			apiKey:  "invalidApiKey",
			want:    BranchSchemaCompareResponse{},
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
				got, err := c.GetProjectBranchSchemaComparison(tt.args.projectID, tt.args.branchID, tt.args.baseBranchID, tt.args.dbName, tt.args.lsn, tt.args.timestamp, tt.args.baseLsn, tt.args.baseTimestamp)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectBranchSchemaComparison() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectBranchSchemaComparison() got = %v, want %v", got, tt.want)
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

func Test_client_GetProjectJWKS(t *testing.T) {
	deserializeResp := func(s string) ProjectJWKSResponse {
		var v ProjectJWKSResponse
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
		want    ProjectJWKSResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/jwks"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    ProjectJWKSResponse{},
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
				got, err := c.GetProjectJWKS(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetProjectJWKS() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetProjectJWKS() got = %v, want %v", got, tt.want)
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

func Test_client_ListNeonAuthIntegrations(t *testing.T) {
	deserializeResp := func(s string) ListNeonAuthIntegrationsResponse {
		var v ListNeonAuthIntegrationsResponse
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
		want    ListNeonAuthIntegrationsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/auth/integrations"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    ListNeonAuthIntegrationsResponse{},
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
				got, err := c.ListNeonAuthIntegrations(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListNeonAuthIntegrations() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListNeonAuthIntegrations() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListOrgApiKeys(t *testing.T) {
	deserializeResp := func(s string) []OrgApiKeysListResponseItem {
		var v []OrgApiKeysListResponseItem
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    []OrgApiKeysListResponseItem
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/api_keys"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID: "foo",
			},
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
				got, err := c.ListOrgApiKeys(tt.args.orgID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListOrgApiKeys() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListOrgApiKeys() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_ListOrganizationVPCEndpoints(t *testing.T) {
	deserializeResp := func(s string) VPCEndpointsResponse {
		var v VPCEndpointsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID    string
		regionID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    VPCEndpointsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID:    "foo",
				regionID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/vpc/region/{region_id}/vpc_endpoints"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID:    "foo",
				regionID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    VPCEndpointsResponse{},
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
				got, err := c.ListOrganizationVPCEndpoints(tt.args.orgID, tt.args.regionID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListOrganizationVPCEndpoints() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListOrganizationVPCEndpoints() got = %v, want %v", got, tt.want)
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

func Test_client_ListProjectBranches(t *testing.T) {
	deserializeResp := func(s string) ListProjectBranchesRespObj {
		var v ListProjectBranchesRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		search    *string
		sortBy    *string
		cursor    *string
		sortOrder *string
		limit     *int
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    ListProjectBranchesRespObj
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				search:    createPointer("foo"),
				sortBy:    createPointer("foo"),
				cursor:    createPointer("foo"),
				sortOrder: createPointer("foo"),
				limit:     createPointer(1),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				search:    createPointer("foo"),
				sortBy:    createPointer("foo"),
				cursor:    createPointer("foo"),
				sortOrder: createPointer("foo"),
				limit:     createPointer(1),
			},
			apiKey:  "invalidApiKey",
			want:    ListProjectBranchesRespObj{},
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
				got, err := c.ListProjectBranches(tt.args.projectID, tt.args.search, tt.args.sortBy, tt.args.cursor, tt.args.sortOrder, tt.args.limit)
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

func Test_client_ListProjectVPCEndpoints(t *testing.T) {
	deserializeResp := func(s string) VPCEndpointsResponse {
		var v VPCEndpointsResponse
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
		want    VPCEndpointsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/vpc_endpoints"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    VPCEndpointsResponse{},
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
				got, err := c.ListProjectVPCEndpoints(tt.args.projectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListProjectVPCEndpoints() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ListProjectVPCEndpoints() got = %v, want %v", got, tt.want)
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
		cursor  *string
		limit   *int
		search  *string
		orgID   *string
		timeout *int
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
				cursor:  createPointer("foo"),
				limit:   createPointer(1),
				search:  createPointer("foo"),
				orgID:   createPointer("foo"),
				timeout: createPointer(1),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cursor:  createPointer("foo"),
				limit:   createPointer(1),
				search:  createPointer("foo"),
				orgID:   createPointer("foo"),
				timeout: createPointer(1),
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
				got, err := c.ListProjects(tt.args.cursor, tt.args.limit, tt.args.search, tt.args.orgID, tt.args.timeout)
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

func Test_client_ListSharedProjects(t *testing.T) {
	deserializeResp := func(s string) ListSharedProjectsRespObj {
		var v ListSharedProjectsRespObj
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cursor  *string
		limit   *int
		search  *string
		timeout *int
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
				cursor:  createPointer("foo"),
				limit:   createPointer(1),
				search:  createPointer("foo"),
				timeout: createPointer(1),
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/shared"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cursor:  createPointer("foo"),
				limit:   createPointer(1),
				search:  createPointer("foo"),
				timeout: createPointer(1),
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
				got, err := c.ListSharedProjects(tt.args.cursor, tt.args.limit, tt.args.search, tt.args.timeout)
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

func Test_client_RemoveOrganizationMember(t *testing.T) {
	deserializeResp := func(s string) EmptyResponse {
		var v EmptyResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID    string
		memberID string
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EmptyResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID:    "foo",
				memberID: "foo",
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/members/{member_id}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID:    "foo",
				memberID: "foo",
			},
			apiKey:  "invalidApiKey",
			want:    EmptyResponse{},
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
				got, err := c.RemoveOrganizationMember(tt.args.orgID, tt.args.memberID)
				if (err != nil) != tt.wantErr {
					t.Errorf("RemoveOrganizationMember() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RemoveOrganizationMember() got = %v, want %v", got, tt.want)
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

func Test_client_RestartProjectEndpoint(t *testing.T) {
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
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/endpoints/{endpoint_id}/restart"]["POST"].Content),
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
				got, err := c.RestartProjectEndpoint(tt.args.projectID, tt.args.endpointID)
				if (err != nil) != tt.wantErr {
					t.Errorf("RestartProjectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RestartProjectEndpoint() got = %v, want %v", got, tt.want)
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

func Test_client_RevokeOrgApiKey(t *testing.T) {
	deserializeResp := func(s string) OrgApiKeyRevokeResponse {
		var v OrgApiKeyRevokeResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID string
		keyID int64
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    OrgApiKeyRevokeResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID: "foo",
				keyID: 1,
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/api_keys/{key_id}"]["DELETE"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID: "foo",
				keyID: 1,
			},
			apiKey:  "invalidApiKey",
			want:    OrgApiKeyRevokeResponse{},
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
				got, err := c.RevokeOrgApiKey(tt.args.orgID, tt.args.keyID)
				if (err != nil) != tt.wantErr {
					t.Errorf("RevokeOrgApiKey() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RevokeOrgApiKey() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_SetDefaultProjectBranch(t *testing.T) {
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
			want:    deserializeResp(endpointResponseExamples["/projects/{project_id}/branches/{branch_id}/set_as_default"]["POST"].Content),
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
				got, err := c.SetDefaultProjectBranch(tt.args.projectID, tt.args.branchID)
				if (err != nil) != tt.wantErr {
					t.Errorf("SetDefaultProjectBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("SetDefaultProjectBranch() got = %v, want %v", got, tt.want)
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

func Test_client_TransferNeonAuthProviderProject(t *testing.T) {
	deserializeResp := func(s string) NeonAuthTransferAuthProviderProjectResponse {
		var v NeonAuthTransferAuthProviderProjectResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cfg NeonAuthTransferAuthProviderProjectRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    NeonAuthTransferAuthProviderProjectResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: NeonAuthTransferAuthProviderProjectRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/projects/auth/transfer_ownership"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cfg: NeonAuthTransferAuthProviderProjectRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    NeonAuthTransferAuthProviderProjectResponse{},
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
				got, err := c.TransferNeonAuthProviderProject(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("TransferNeonAuthProviderProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TransferNeonAuthProviderProject() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_TransferProjectsFromOrgToOrg(t *testing.T) {
	deserializeResp := func(s string) EmptyResponse {
		var v EmptyResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		sourceOrgID string
		cfg         TransferProjectsToOrganizationRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EmptyResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				sourceOrgID: "foo",
				cfg:         TransferProjectsToOrganizationRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{source_org_id}/projects/transfer"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				sourceOrgID: "foo",
				cfg:         TransferProjectsToOrganizationRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    EmptyResponse{},
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
				got, err := c.TransferProjectsFromOrgToOrg(tt.args.sourceOrgID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("TransferProjectsFromOrgToOrg() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TransferProjectsFromOrgToOrg() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_TransferProjectsFromUserToOrg(t *testing.T) {
	deserializeResp := func(s string) EmptyResponse {
		var v EmptyResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		cfg TransferProjectsToOrganizationRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    EmptyResponse
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: TransferProjectsToOrganizationRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/users/me/projects/transfer"]["POST"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				cfg: TransferProjectsToOrganizationRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    EmptyResponse{},
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
				got, err := c.TransferProjectsFromUserToOrg(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("TransferProjectsFromUserToOrg() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TransferProjectsFromUserToOrg() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_UpdateOrganizationMember(t *testing.T) {
	deserializeResp := func(s string) Member {
		var v Member
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		orgID    string
		memberID string
		cfg      OrganizationMemberUpdateRequest
	}
	tests := []struct {
		name    string
		args    args
		apiKey  string
		want    Member
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				orgID:    "foo",
				memberID: "foo",
				cfg:      OrganizationMemberUpdateRequest{},
			},
			apiKey:  "foo",
			want:    deserializeResp(endpointResponseExamples["/organizations/{org_id}/members/{member_id}"]["PATCH"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				orgID:    "foo",
				memberID: "foo",
				cfg:      OrganizationMemberUpdateRequest{},
			},
			apiKey:  "invalidApiKey",
			want:    Member{},
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
				got, err := c.UpdateOrganizationMember(tt.args.orgID, tt.args.memberID, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateOrganizationMember() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UpdateOrganizationMember() got = %v, want %v", got, tt.want)
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
