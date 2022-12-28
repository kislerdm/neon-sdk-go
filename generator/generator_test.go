package generator

import (
	"bytes"
	_ "embed"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

//go:embed fixtures/openapi.json
var openAPIFixture []byte

var (
	openAPIFixtureSpec openAPISpec
)

func init() {
	if err := openAPIFixtureSpec.UnmarshalJSON(openAPIFixture); err != nil {
		panic(err)
	}
}

func helperExtractSchemaNames(spec openAPISpec) (o []string) {
	for k := range spec.Components.Schemas {
		o = append(o, k)
	}
	for k := range spec.Components.Responses {
		o = append(o, k)
	}
	return
}

func TestRun(t *testing.T) {
	createTempDir := func() string {
		dir := "/tmp"
		if s := os.Getenv("TEMP_DIR"); s != "" {
			dir = s
		}
		s := dir + "/" +
			strings.ReplaceAll(
				time.Now().UTC().Format("20060102150405.000"),
				".", "",
			)
		if err := os.MkdirAll(s, 0774); err != nil {
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
					OpenAPIReader: bytes.NewReader(openAPIFixture),
					PathOutput:    createTempDir(),
				},
			},
			wantErr: false,
			files: map[string]struct{}{
				"go.mod":         {},
				"doc.go":         {},
				"sdk.go":         {},
				"client_test.go": {},
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

func Test_endpointImplementation_generateMethodImplementation(t *testing.T) {
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
			want: `func (c *client) ListProjects() (ProjectsResponse, error) {
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
			want: `func (c *client) GetProject(projectID string) (ProjectsResponse, error) {
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
			want: `func (c *client) ListProjectBranchDatabases(projectID string, branchID string) (DatabasesResponse, error) {
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
			want: `func (c *client) RevokeApiKey(keyID int64) (ApiKeyRevokeResponse, error) {
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
			want: `func (c *client) CreateProject(cfg *ProjectCreateRequest) (CreatedProject, error) {
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
				assert.Equal(
					t, tt.want,
					endpointImplementation{
						Name:                  tt.fields.Name,
						Method:                tt.fields.Method,
						Route:                 tt.fields.Route,
						Description:           tt.fields.Description,
						RequestBodyStruct:     tt.fields.RequestBodyStruct,
						ResponseStruct:        tt.fields.ResponseStruct,
						RequestParametersPath: tt.fields.RequestParametersPath,
					}.generateMethodImplementation(),
				)
			},
		)
	}
}

var inputSpec = openAPISpec{
	T: openapi3.T{
		OpenAPI: "3.0.3",
		Components: openapi3.Components{
			Responses: openapi3.Responses{
				"FooBarResponse": {
					Value: &openapi3.Response{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										AllOf: openapi3.SchemaRefs{
											{Ref: "#/components/responses/FooResponse"},
											{Ref: "#/components/responses/BarResponse"},
										},
									},
								},
							},
						},
					},
				},
				"FooResponse": {
					Value: &openapi3.Response{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     openapi3.TypeObject,
										Required: []string{"foo"},
										Properties: openapi3.Schemas{
											"foo": &openapi3.SchemaRef{
												Ref: "#/components/schemas/Foo",
											},
										},
									},
								},
							},
						},
					},
				},
				"BarResponse": {
					Value: &openapi3.Response{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:     openapi3.TypeObject,
										Required: []string{"bar"},
										Properties: openapi3.Schemas{
											"bar": &openapi3.SchemaRef{
												Ref: "#/components/schemas/Bar",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Schemas: openapi3.Schemas{
				"Foo": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:     openapi3.TypeObject,
						Required: []string{"foo_id", "bar"},
						Properties: openapi3.Schemas{
							"foo_id": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:      openapi3.TypeString,
									MinLength: 2,
								},
							},
							"bar": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:   openapi3.TypeInteger,
									Format: "int64",
								},
							},
						},
					},
				},
				"Bar": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:     openapi3.TypeObject,
						Required: []string{"type"},
						Properties: openapi3.Schemas{
							"type": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: openapi3.TypeString,
									Enum: []interface{}{"init", "ready"},
								},
							},
						},
					},
				},
				"Qux": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:     openapi3.TypeObject,
						Required: []string{"foo"},
						Properties: openapi3.Schemas{
							"foo": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:     openapi3.TypeObject,
									Required: []string{"foo", "bar", "qux"},
									Properties: openapi3.Schemas{
										"foo": {
											Value: &openapi3.Schema{
												Type:   openapi3.TypeInteger,
												Format: "int32",
											},
										},
										"bar": {
											Value: &openapi3.Schema{
												Type:     openapi3.TypeObject,
												Required: []string{"foo"},
												Properties: openapi3.Schemas{
													"foo": {
														Value: &openapi3.Schema{
															Type: openapi3.TypeArray,
															Items: &openapi3.SchemaRef{
																Value: &openapi3.Schema{
																	Type:   openapi3.TypeString,
																	Format: "date-time",
																},
															},
														},
													},
													"bar": {
														Ref: "#/components/schemas/Bar",
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
			},
		},
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
					Parameters: openapi3.Parameters{
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
												Items: &openapi3.SchemaRef{
													Ref: "#/components/schemas/Foo",
												},
											},
										},
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
					Parameters: openapi3.Parameters{
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
					},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Content: openapi3.Content{
									"application/json": &openapi3.MediaType{
										Schema: &openapi3.SchemaRef{
											Ref: "#/components/responses/FooBarResponse",
										},
										Example: map[string]interface{}{
											"foo": map[string]interface{}{
												"foo_id": "bar",
												"bar":    1,
											},
											"bar": "init",
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
				},
			},
		},
	},
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
				o: inputSpec,
			},
			wantEndpoints: []endpointImplementation{
				{
					Name:              "FooEndpoint",
					Method:            "GET",
					Route:             "/foo/{bar}/{qux_id}",
					Description:       "get /foo",
					RequestBodyStruct: "",
					ResponseStruct:    "[]Foo",
					ResponsePositivePathExample: []map[string]interface{}{
						{
							"foo_id": "bar",
							"bar":    1,
						},
						{
							"foo_id": "aff",
							"bar":    2,
						},
					},
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
					ResponsePositivePathExample: map[string]interface{}{
						"foo": map[string]interface{}{
							"foo_id": "bar",
							"bar":    1,
						},
						"bar": "init",
					},
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
				assert.Equal(t, tt.wantEndpoints, generateEndpointsImplementationMethods(tt.args.o))
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
				assert.Equal(
					t, tt.want, field{
						k: tt.fields.k,
						v: tt.fields.v,
					}.canonicalName(),
				)
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
				assert.Equal(
					t, tt.want, field{
						k:      tt.fields.k,
						v:      tt.fields.v,
						format: tt.fields.format,
					}.routeElement(),
				)
			},
		)
	}
}

func Test_generateModels(t *testing.T) {
	type args struct {
		spec openAPISpec
	}
	tests := []struct {
		name                 string
		args                 args
		want                 models
		checkElementWiseOnly bool
	}{
		{
			name: "five models: three responses, two schemas",
			args: args{
				spec: inputSpec,
			},
			want: models{
				"FooBarResponse": model{
					name:     "FooBarResponse",
					children: map[string]struct{}{"FooResponse": {}, "BarResponse": {}},
				},
				"FooResponse": model{
					name: "FooResponse",
					fields: map[string]*field{
						"foo": {
							k:        "foo",
							v:        "Foo",
							format:   "",
							required: true,
						},
					},
					children: map[string]struct{}{"Foo": {}},
				},
				"BarResponse": model{
					name: "BarResponse",
					fields: map[string]*field{
						"bar": {
							k:        "bar",
							v:        "Bar",
							format:   "",
							required: true,
						},
					},
					children: map[string]struct{}{"Bar": {}},
				},
				"Qux": model{
					name: "Qux",
					fields: map[string]*field{
						"foo": {
							k:        "foo",
							v:        "QuxFoo",
							required: true,
						},
					},
					children: map[string]struct{}{"QuxFoo": {}},
				},
				"QuxFoo": model{
					name: "QuxFoo",
					fields: map[string]*field{
						"foo": {
							k:        "foo",
							v:        openapi3.TypeInteger,
							format:   "int32",
							required: true,
						},
						"bar": {
							k:        "bar",
							v:        "QuxFooBar",
							required: true,
						},
					},
					children: map[string]struct{}{"QuxFooBar": {}},
				},
				"QuxFooBar": model{
					name: "QuxFooBar",
					fields: map[string]*field{
						"foo": {
							k:        "foo",
							v:        "[]time.Time",
							format:   "",
							required: true,
						},
						"bar": {
							k: "bar",
							v: "Bar",
						},
					},
					children: map[string]struct{}{"Bar": {}},
				},
				"Foo": model{
					name: "Foo",
					fields: map[string]*field{
						"foo_id": {
							k:        "foo_id",
							v:        openapi3.TypeString,
							format:   "",
							required: true,
						},
						"bar": {
							k:        "bar",
							v:        openapi3.TypeInteger,
							format:   "int64",
							required: true,
						},
					},
				},
				"Bar": model{
					name: "Bar",
					fields: map[string]*field{
						"type": {
							k:        "type",
							v:        "string",
							required: true,
						},
					},
				},
			},
		},
		{
			name: "fixture",
			args: args{
				spec: openAPIFixtureSpec,
			},
			want:                 nil,
			checkElementWiseOnly: true,
		},
		{
			name: "primitive type",
			args: args{
				spec: openAPISpec{
					T: openapi3.T{
						OpenAPI: "3.0.3",
						Components: openapi3.Components{
							Schemas: openapi3.Schemas{
								"PgVersion": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Description: "Major version of the Postgres",
										Type:        openapi3.TypeString,
									},
								},
							},
						},
					},
				},
			},
			want: map[string]model{
				"PgVersion": {
					name:        "PgVersion",
					description: "Major version of the Postgres",
					primitive:   fieldType{name: openapi3.TypeString},
				},
			},
		},
		{
			name: "array of references",
			args: args{
				spec: openAPISpec{
					T: openapi3.T{
						OpenAPI: "3.0.3",
						Components: openapi3.Components{
							Schemas: openapi3.Schemas{
								"VercelIntegration": &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Description: "Vercel integration is bound to a Neon branch.\nUser specifies endpoint to expose to each Vercel project.\n",
										Type:        openapi3.TypeObject,
										Properties: openapi3.Schemas{
											"details": {
												Value: &openapi3.Schema{
													Type: openapi3.TypeArray,
													Items: &openapi3.SchemaRef{
														Value: &openapi3.Schema{
															AllOf: openapi3.SchemaRefs{
																{Ref: "#/components/schemas/VercelIntegrationDetailsResponse"},
																{Ref: "#/components/schemas/ProjectResponse"},
																{Ref: "#/components/schemas/BranchResponse"},
																{Ref: "#/components/schemas/EndpointResponse"},
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
					},
				},
			},
			want: map[string]model{
				"VercelIntegration": {
					name:        "VercelIntegration",
					description: "Vercel integration is bound to a Neon branch.\nUser specifies endpoint to expose to each Vercel project.\n",
					fields: map[string]*field{
						"details": {
							k: "details",
							v: `[]struct {
VercelIntegrationDetailsResponse
ProjectResponse
BranchResponse
EndpointResponse
}`,
							format: "",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := generateModels(tt.args.spec)
				if tt.checkElementWiseOnly {
					assert.Subset(t, got, helperExtractSchemaNames(tt.args.spec))
					return
				}
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func Test_models_generateCode(t *testing.T) {
	tests := []struct {
		name string
		v    models
		want []string
	}{
		{
			name: "one type, one field: ref to schemas only",
			v: models{
				"FooBarResponse": model{
					children: map[string]struct{}{"FooResponse": {}, "BarResponse": {}},
				},
			},
			want: []string{
				`type FooBarResponse struct {
FooResponse
BarResponse
}`,
			},
		},
		{
			name: "one type, one field: ref type",
			v: models{
				"FooResponse": model{
					fields: map[string]*field{
						"foo": {
							k:        "foo",
							v:        "Foo",
							format:   "",
							required: true,
						},
					},
					children: map[string]struct{}{"Foo": {}},
				},
			},
			want: []string{
				"type FooResponse struct {\nFoo Foo `json:\"foo\"`\n}",
			},
		},
		{
			name: "one type, two fields: type required import and ref type",
			v: models{
				"QuxFooBar": model{
					fields: map[string]*field{
						"foo": {
							k:        "foo",
							v:        "[]time.Time",
							format:   "",
							required: true,
						},
						"bar": {
							k: "bar",
							v: "Bar",
						},
					},
					children: map[string]struct{}{"Bar": {}},
				},
			},
			want: []string{
				"type QuxFooBar struct {\nFoo []time.Time `json:\"foo\"`\nBar Bar `json:\"bar,omitempty\"`\n}",
			},
		},
		{
			name: "primitive type",
			v: models{
				"EndpointPoolerMode": model{
					primitive: fieldType{
						name: "string",
					},
				},
			},
			want: []string{"type EndpointPoolerMode string"},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := tt.v.generateCode()
				for _, el := range tt.want {
					assert.Contains(t, got, el, "generateCode()")
				}
			},
		)
	}
}

func Test_objNameGoConventionExport(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "autoscaling_limit_max_cu",
			args: args{"autoscaling_limit_max_cu"},
			want: "AutoscalingLimitMaxCu",
		},
		{
			name: "QUERY PLAN",
			args: args{"QUERY PLAN"},
			want: "QueryPlan",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t, tt.want, objNameGoConventionExport(tt.args.s), "objNameGoConventionExport(%v)", tt.args.s,
				)
			},
		)
	}
}

func Test_endpointImplementation_generateMethodDefinition(t *testing.T) {
	type fields struct {
		Name                        string
		Method                      string
		Route                       string
		Description                 string
		RequestBodyRequires         bool
		RequestBodyStruct           string
		RequestBodyStructExample    interface{}
		ResponseStruct              string
		RequestParametersPath       []field
		ResponsePositivePathExample interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "CreateProjectBranch",
			fields: fields{
				Name:                     "CreateProjectBranch",
				Method:                   "POST",
				Route:                    "/projects/{project_id}/branches",
				Description:              "Creates a branch in the specified project",
				RequestBodyRequires:      false,
				RequestBodyStruct:        "BranchCreateRequest",
				RequestBodyStructExample: nil,
				ResponseStruct:           "CreatedBranch",
				RequestParametersPath: []field{
					{
						k:        "project_id",
						v:        "string",
						required: true,
						isInPath: true,
					},
				},
			},
			want: `// CreateProjectBranch Creates a branch in the specified project
CreateProjectBranch(projectID string, cfg *BranchCreateRequest) (CreatedBranch, error)`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := endpointImplementation{
					Name:                        tt.fields.Name,
					Method:                      tt.fields.Method,
					Route:                       tt.fields.Route,
					Description:                 tt.fields.Description,
					RequestBodyRequires:         tt.fields.RequestBodyRequires,
					RequestBodyStruct:           tt.fields.RequestBodyStruct,
					RequestBodyStructExample:    tt.fields.RequestBodyStructExample,
					ResponseStruct:              tt.fields.ResponseStruct,
					RequestParametersPath:       tt.fields.RequestParametersPath,
					ResponsePositivePathExample: tt.fields.ResponsePositivePathExample,
				}
				assert.Equalf(t, tt.want, e.generateMethodDefinition(), "generateMethodDefinition()")
			},
		)
	}
}

func Test_extractStructFromSchemaRef(t *testing.T) {
	type args struct {
		schema *openapi3.SchemaRef
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "reference case",
			args: args{
				schema: &openapi3.SchemaRef{Ref: "#/components/schemas/ExplainData"},
			},
			want: "ExplainData",
		},
		{
			name: "array of arrays",
			args: args{
				schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: openapi3.TypeArray,
						Items: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: openapi3.TypeArray,
								Items: &openapi3.SchemaRef{
									Value: &openapi3.Schema{Type: openapi3.TypeString},
								},
							},
						},
					},
				},
			},
			want: "[][]string",
		},
		{
			name: "array of arrays of arrays",
			args: args{
				schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: openapi3.TypeArray,
						Items: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: openapi3.TypeArray,
								Items: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: openapi3.TypeArray,
										Items: &openapi3.SchemaRef{
											Value: &openapi3.Schema{
												Type: openapi3.TypeString,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: "[][][]string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, extractStructFromSchemaRef(tt.args.schema), "extractStructFromSchemaRef(%v)", tt.args.schema)
		})
	}
}
