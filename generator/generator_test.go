package generator

import (
	_ "embed"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed fixtures/openapi.json
var openAPIFixture string

func TestRun(t *testing.T) {
	createTempDir := func() string {
		s := "/tmp/" + time.Now().UTC().Format("20060102150405.000")
		s = strings.ReplaceAll(s, ".", "")
		if err := os.Mkdir(s, 0774); err != nil {
			panic(err)
		}
		return s
	}

	type args struct {
		cfg Config
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		files   map[string]struct{}
	}{
		{
			name: "happy path",
			args: args{
				cfg: Config{
					OpenAPIReader: strings.NewReader(openAPIFixture),
					PathOutput:    createTempDir(),
				},
			},
			wantErr: false,
			files: map[string]struct{}{
				"go.mod":         {},
				"go.sum":         {},
				"doc.go":         {},
				"client.go":      {},
				"client_test.go": {},
				"endpoints.go":   {},
				"models.go":      {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// WHEN
				// run generator
				if err := Run(tt.args.cfg); (err != nil) != tt.wantErr {
					t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				}

				// WANT
				// all generated files are present in the output dir
				if err := fs.WalkDir(
					os.DirFS(tt.args.cfg.PathOutput), ".", func(path string, d fs.DirEntry, err error) error {
						if path == "." {
							return nil
						}

						if _, ok := tt.files[d.Name()]; !ok {
							t.Errorf(d.Name() + " is not expected to be generated")
						} else {
							delete(tt.files, d.Name())
						}
						return nil
					},
				); err != nil {
					panic(err)
				}

				if len(tt.files) > 0 {
					t.Errorf("not all expected files were generated")
				}
			},
		)
		//t.Cleanup(
		//	func() {
		//		if err := os.RemoveAll(tt.args.cfg.PathOutput); err != nil {
		//			panic(err)
		//		}
		//	},
		//)
	}
}

func Test_endpointImplementation_generateFunctionCode(t *testing.T) {
	type fields struct {
		Name                  string
		Method                string
		Route                 string
		Description           string
		RequestBodyStruct     string
		ResponseStruct        string
		RequestParametersPath []field
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "list objects",
			fields: fields{
				Name:                  "ListProjects",
				Method:                "GET",
				Route:                 "/projects",
				Description:           "Retrieves a list of projects for the Neon account",
				RequestBodyStruct:     "",
				ResponseStruct:        "ProjectsResponse",
				RequestParametersPath: nil,
			},
			want: `// ListProjects Retrieves a list of projects for the Neon account
func (c *Client) ListProjects() (ProjectsResponse, error) {
	var v ProjectsResponse
	if err := c.requestHandler(c.baseURL+"/projects", "GET", nil, &v); err != nil {
		return ProjectsResponse{}, err
	}
	return v, nil
}`,
		},
		{
			name: "get project details",
			fields: fields{
				Name:                  "GetProject",
				Method:                "GET",
				Route:                 "/projects/{project_id}",
				Description:           "Retrieves information about the specified project",
				RequestBodyStruct:     "",
				ResponseStruct:        "ProjectsResponse",
				RequestParametersPath: []field{{"project_id", "string", "", true, true, false}},
			},
			want: `// GetProject Retrieves information about the specified project
func (c *Client) GetProject(projectID string) (ProjectsResponse, error) {
	var v ProjectsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "GET", nil, &v); err != nil {
		return ProjectsResponse{}, err
	}
	return v, nil
}`,
		},
		{
			name: "get project details",
			fields: fields{
				Name:              "ListProjectBranchDatabases",
				Method:            "GET",
				Route:             "/projects/{project_id}/branches/{branch_id}/databases",
				Description:       "Retrieves a list of databases for the specified branch",
				RequestBodyStruct: "",
				ResponseStruct:    "DatabasesResponse",
				RequestParametersPath: []field{
					{"project_id", "string", "", true, true, false},
					{"branch_id", "string", "", true, true, false},
				},
			},
			want: `// ListProjectBranchDatabases Retrieves a list of databases for the specified branch
func (c *Client) ListProjectBranchDatabases(projectID string, branchID string) (DatabasesResponse, error) {
	var v DatabasesResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID+"/branches/"+branchID+"/databases", "GET", nil, &v); err != nil {
		return DatabasesResponse{}, err
	}
	return v, nil
}`,
		},
		{
			name: "revoke api key",
			fields: fields{
				Name:                  "RevokeApiKey",
				Method:                "DELETE",
				Route:                 "/api_keys/{key_id}",
				Description:           "Revokes the specified API key",
				RequestBodyStruct:     "",
				ResponseStruct:        "ApiKeyRevokeResponse",
				RequestParametersPath: []field{{"key_id", "integer", "int64", true, true, false}},
			},
			want: `// RevokeApiKey Revokes the specified API key
func (c *Client) RevokeApiKey(keyID int64) (ApiKeyRevokeResponse, error) {
	var v ApiKeyRevokeResponse
	if err := c.requestHandler(c.baseURL+"/api_keys/"+strconv.FormatInt(keyID, 10), "DELETE", nil, &v); err != nil {
		return ApiKeyRevokeResponse{}, err
	}
	return v, nil
}`,
		},
		{
			name: "create a project",
			fields: fields{
				Name:                  "CreateProject",
				Method:                "POST",
				Route:                 "/projects",
				Description:           "Creates a Neon project",
				RequestBodyStruct:     "ProjectCreateRequest",
				ResponseStruct:        "CreatedProject",
				RequestParametersPath: nil,
			},
			want: `// CreateProject Creates a Neon project
func (c *Client) CreateProject(cfg *ProjectCreateRequest) (CreatedProject, error) {
	var v CreatedProject
	if err := c.requestHandler(c.baseURL+"/projects", "POST", cfg, &v); err != nil {
		return CreatedProject{}, err
	}
	return v, nil
}`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := endpointImplementation{
					Name:                  tt.fields.Name,
					Method:                tt.fields.Method,
					Route:                 tt.fields.Route,
					Description:           tt.fields.Description,
					RequestBodyStruct:     tt.fields.RequestBodyStruct,
					ResponseStruct:        tt.fields.ResponseStruct,
					RequestParametersPath: tt.fields.RequestParametersPath,
				}
				if got := e.generateFunctionCode(); got != tt.want {
					t.Errorf("generateFunctionCode() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_generateEndpointsImplementationMethods(t *testing.T) {
	type args struct {
		o openAPISpec
	}
	tests := []struct {
		name          string
		args          args
		wantEndpoints []endpointImplementation
	}{
		{
			name: "happy path",
			args: args{
				o: openAPISpec{
					T: openapi3.T{
						OpenAPI:    "3.0.3",
						Components: openapi3.Components{},
						Info: &openapi3.Info{
							Title:       "foo",
							Description: "bar",
							Version:     "v2",
						},
						Paths: openapi3.Paths{
							"/foo/{bar}/{qux_id}": {
								Summary:     "/foo endpoint",
								Description: "/foo endpoint",
								Connect:     nil,
								Delete:      nil,
								Get: &openapi3.Operation{
									Summary:     "get /foo",
									Description: "get /foo",
									OperationID: "fooEndpoint",
									Responses: openapi3.Responses{
										"200": &openapi3.ResponseRef{
											Ref: "",
											Value: &openapi3.Response{
												Content: openapi3.Content{
													"application/json": &openapi3.MediaType{
														Schema: &openapi3.SchemaRef{
															Ref: "",
															Value: &openapi3.Schema{
																Type: "array",
																Example: []map[string]interface{}{
																	{
																		"foo_id": "bar",
																		"bar":    1,
																	},
																	{
																		"foo_id": "aff",
																		"bar":    2,
																	},
																},
																Items: &openapi3.SchemaRef{
																	Ref: "#/components/schemas/FooResponse",
																},
															},
														},
													},
												},
											},
										},
									},
									Deprecated: false,
								},
								Parameters: openapi3.Parameters{
									{
										Ref: "",
										Value: &openapi3.Parameter{
											Name:            "bar",
											In:              "path",
											Description:     "bar parameter",
											AllowEmptyValue: false,
											Required:        true,
											Schema: &openapi3.SchemaRef{
												Ref: "",
												Value: &openapi3.Schema{
													Type: "string",
												},
											},
										},
									},
									{
										Ref: "",
										Value: &openapi3.Parameter{
											Name:            "qux_id",
											In:              "path",
											Description:     "qux parameter",
											AllowEmptyValue: false,
											Required:        true,
											Schema: &openapi3.SchemaRef{
												Ref: "",
												Value: &openapi3.Schema{
													Type:   "integer",
													Format: "int64",
												},
											},
										},
									},
								},
							},
							"/foo/bar/{qux_id}/{date_submit}": {
								Summary:     "/foo/bar endpoint",
								Description: "/foo/bar endpoint",
								Connect:     nil,
								Delete:      nil,
								Get: &openapi3.Operation{
									Summary:     "get /foo/bar",
									Description: "get /foo/bar",
									OperationID: "fooBarEndpoint",
									Responses: openapi3.Responses{
										"200": &openapi3.ResponseRef{
											Ref: "",
											Value: &openapi3.Response{
												Content: openapi3.Content{
													"application/json": &openapi3.MediaType{
														Schema: &openapi3.SchemaRef{
															Ref: "#/components/schemas/FooBarResponse",
														},
													},
												},
											},
										},
									},
									Deprecated: false,
								},
								Parameters: openapi3.Parameters{
									{
										Value: &openapi3.Parameter{
											Name:            "qux_id",
											In:              openapi3.ParameterInPath,
											Description:     "qux parameter",
											AllowEmptyValue: false,
											Required:        true,
											Schema: &openapi3.SchemaRef{
												Ref: "",
												Value: &openapi3.Schema{
													Type:   "integer",
													Format: "int64",
												},
											},
										},
									},
									{
										Value: &openapi3.Parameter{
											Name:            "date_submit",
											In:              openapi3.ParameterInPath,
											Description:     "date parameter",
											AllowEmptyValue: false,
											Required:        true,
											Schema: &openapi3.SchemaRef{
												Ref: "",
												Value: &openapi3.Schema{
													Type:   "string",
													Format: "date-time",
												},
											},
										},
									},
									{
										Value: &openapi3.Parameter{
											Name:            "limit",
											In:              openapi3.ParameterInQuery,
											Description:     "query limit",
											AllowEmptyValue: false,
											Required:        false,
											Schema: &openapi3.SchemaRef{
												Ref: "",
												Value: &openapi3.Schema{
													Type: "integer",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantEndpoints: []endpointImplementation{
				{
					Name:              "FooEndpoint",
					Method:            "GET",
					Route:             "/foo/{bar}/{qux_id}",
					Description:       "get /foo",
					RequestBodyStruct: "",
					ResponseStruct:    "[]FooResponse",
					RequestParametersPath: []field{
						{
							k:        "bar",
							v:        "string",
							format:   "",
							required: true,
							isInPath: true,
						},
						{
							k:        "qux_id",
							v:        "integer",
							format:   "int64",
							required: true,
							isInPath: true,
						},
					},
				},
				{
					Name:              "FooBarEndpoint",
					Method:            "GET",
					Route:             "/foo/bar/{qux_id}/{date_submit}",
					Description:       "get /foo/bar",
					RequestBodyStruct: "",
					ResponseStruct:    "FooBarResponse",
					RequestParametersPath: []field{
						{
							k:        "qux_id",
							v:        "integer",
							format:   "int64",
							required: true,
							isInPath: true,
						},
						{
							k:        "date_submit",
							v:        "string",
							format:   "date-time",
							required: true,
							isInPath: true,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if gotEndpoints := generateEndpointsImplementationMethods(tt.args.o, nil); !reflect.DeepEqual(
					gotEndpoints, tt.wantEndpoints,
				) {
					t.Errorf("generateEndpointsImplementationMethods() = %v, want %v", gotEndpoints, tt.wantEndpoints)
				}
			},
		)
	}
}

func Test_parameterPath_canonicalName(t *testing.T) {
	type fields struct {
		k string
		v string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "project_id",
			fields: fields{"project_id", "string"},
			want:   "projectID",
		},
		{
			name:   "api_key_id",
			fields: fields{"api_key_id", "string"},
			want:   "apiKeyID",
		},
		{
			name:   "database_name",
			fields: fields{"database_name", "string"},
			want:   "databaseName",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := field{
					k: tt.fields.k,
					v: tt.fields.v,
				}
				if got := v.canonicalName(); got != tt.want {
					t.Errorf("canonicalName() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_parameterPath_routeElement(t *testing.T) {
	type fields struct {
		k      string
		v      string
		format string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "int64",
			fields: fields{
				k:      "qux_id",
				v:      "integer",
				format: "int64",
			},
			want: "strconv.FormatInt(quxID, 10)",
		},
		{
			name: "int32",
			fields: fields{
				k:      "qux_id",
				v:      "integer",
				format: "int32",
			},
			want: "strconv.FormatInt(int64(quxID), 10)",
		},
		{
			name: "uuid",
			fields: fields{
				k:      "qux_id",
				v:      "string",
				format: "uuid",
			},
			want: "quxID.String()",
		},
		{
			name: "double",
			fields: fields{
				k:      "qux_id",
				v:      "number",
				format: "double",
			},
			want: "strconv.FormatFloat(quxID, 'f', -1, 64)",
		},
		{
			name: "float",
			fields: fields{
				k:      "qux_id",
				v:      "number",
				format: "float",
			},
			want: "strconv.FormatFloat(quxID, 'f', -1, 32)",
		},
		{
			name: "date-time",
			fields: fields{
				k:      "qux_id",
				v:      "string",
				format: "date-time",
			},
			want: "quxID.Format(time.RFC3339)",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := field{
					k:      tt.fields.k,
					v:      tt.fields.v,
					format: tt.fields.format,
				}
				if got := v.routeElement(); got != tt.want {
					t.Errorf("routeElement() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
