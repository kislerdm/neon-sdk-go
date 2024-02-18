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
				"go.mod":           {},
				"doc.go":           {},
				"sdk.go":           {},
				"sdk_test.go":      {},
				"error.go":         {},
				"mockhttp.go":      {},
				"mockhttp_test.go": {},
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
		t.Cleanup(
			func() {
				if err := os.RemoveAll(tt.args.cfg.PathOutput); err != nil {
					panic(err)
				}
			},
		)
	}
}

func Test_endpointImplementation_generateMethodImplementation(t *testing.T) {
	type fields struct {
		Name                   string
		Method                 string
		Route                  string
		Description            string
		RequestBodyStruct      *model
		ResponseStruct         *model
		RequestParametersPath  []field
		RequestParametersQuery []field
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
				RequestBodyStruct:     nil,
				ResponseStruct:        &model{name: "ListProjectsResponse"},
				RequestParametersPath: nil,
				RequestParametersQuery: []field{
					{
						k:           "cursor",
						v:           "string",
						description: "Specify the cursor value from the previous response to get the next batch of projects.",
						required:    false,
						isInPath:    false,
						isInQuery:   true,
					},
					{
						k:           "limit",
						v:           "integer",
						description: "Specify a value from 1 to 100 to limit number of projects in the response",
						required:    false,
						isInPath:    false,
						isInQuery:   true,
					},
				},
			},
			want: `// ListProjects Retrieves a list of projects for the Neon account
func (c Client) ListProjects(cursor *string, limit *int) (ListProjectsResponse, error) {
	var (
		queryElements []string
		query string
	)
	if cursor != nil {
		queryElements = append(queryElements, "cursor=" + *cursor)
	}
	if limit != nil {
		queryElements = append(queryElements, "limit=" + strconv.FormatInt(int64(*limit), 10))
	}
	if len(queryElements) > 0 {
		query = "?" + strings.Join(queryElements, "&")
	}
	var v ListProjectsResponse
	if err := c.requestHandler(c.baseURL+"/projects" + query, "GET", nil, &v); err != nil {
		return ListProjectsResponse{}, err
	}
	return v, nil
}`,
		},
		{
			name: "get project details",
			fields: fields{
				Name:   "GetProject",
				Method: "GET",
				Route:  "/projects/{project_id}",
				Description: `Retrieves information about the specified project.
foo bar
qux`,
				RequestBodyStruct:     nil,
				ResponseStruct:        &model{name: "ProjectsResponse"},
				RequestParametersPath: []field{{"project_id", "string", "", "", true, true, false}},
			},
			want: `// GetProject Retrieves information about the specified project.
// foo bar
// qux
func (c Client) GetProject(projectID string) (ProjectsResponse, error) {
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
				RequestBodyStruct: nil,
				ResponseStruct:    &model{name: "DatabasesResponse"},
				RequestParametersPath: []field{
					{"project_id", "string", "", "", true, true, false},
					{"branch_id", "string", "", "", true, true, false},
				},
			},
			want: `// ListProjectBranchDatabases Retrieves a list of databases for the specified branch
func (c Client) ListProjectBranchDatabases(projectID string, branchID string) (DatabasesResponse, error) {
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
				RequestBodyStruct:     nil,
				ResponseStruct:        &model{name: "ApiKeyRevokeResponse"},
				RequestParametersPath: []field{{"key_id", "integer", "int64", "", true, true, false}},
			},
			want: `// RevokeApiKey Revokes the specified API key
func (c Client) RevokeApiKey(keyID int64) (ApiKeyRevokeResponse, error) {
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
				RequestBodyStruct:     &model{name: "ProjectCreateRequest"},
				ResponseStruct:        &model{name: "CreatedProject"},
				RequestParametersPath: nil,
			},
			want: `// CreateProject Creates a Neon project
func (c Client) CreateProject(cfg *ProjectCreateRequest) (CreatedProject, error) {
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
						Name:                           tt.fields.Name,
						Method:                         tt.fields.Method,
						Route:                          tt.fields.Route,
						Description:                    tt.fields.Description,
						RequestBodyStruct:              tt.fields.RequestBodyStruct,
						ResponseStruct:                 tt.fields.ResponseStruct,
						RequestParametersPath:          tt.fields.RequestParametersPath,
						RequestParametersQuery:         tt.fields.RequestParametersQuery,
						ResponsePositivePathStatusCode: "",
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
								AllowEmptyValue: true,
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
		o                openAPISpec
		orderedEndpoints []string
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
				orderedEndpoints: []string{
					"/foo/{bar}/{qux_id}",
					"/foo/bar/{qux_id}/{date_submit}",
				},
			},
			wantEndpoints: []endpointImplementation{
				{
					Name:              "FooEndpoint",
					Method:            "GET",
					Route:             "/foo/{bar}/{qux_id}",
					Description:       "get /foo",
					RequestBodyStruct: nil,
					ResponseStruct:    &model{name: "[]Foo"},
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
					ResponsePositivePathStatusCode: "200",
					RequestParametersPath: []field{
						{
							k:           "bar",
							v:           "string",
							description: "bar parameter",
							format:      "",
							required:    true,
							isInPath:    true,
						},
						{
							k:           "qux_id",
							v:           "integer",
							description: "qux parameter",
							format:      "int64",
							required:    true,
							isInPath:    true,
						},
					},
					RequestParametersQuery: []field{
						{
							k:           "limit",
							v:           "integer",
							description: "query limit",
							isInQuery:   true,
						},
					},
				},
				{
					Name:              "FooBarEndpoint",
					Method:            "GET",
					Route:             "/foo/bar/{qux_id}/{date_submit}",
					Description:       "get /foo/bar",
					RequestBodyStruct: nil,
					ResponseStruct:    &model{name: "FooBarResponse"},
					ResponsePositivePathExample: map[string]interface{}{
						"foo": map[string]interface{}{
							"foo_id": "bar",
							"bar":    1,
						},
						"bar": "init",
					},
					ResponsePositivePathStatusCode: "200",
					RequestParametersPath: []field{
						{
							k:           "qux_id",
							description: "qux parameter",
							v:           "integer",
							format:      "int64",
							required:    true,
							isInPath:    true,
						},
						{
							k:           "date_submit",
							description: "date parameter",
							v:           "string",
							format:      "date-time",
							required:    true,
							isInPath:    true,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equal(
					t, tt.wantEndpoints, generateEndpointsImplementationMethods(tt.args.o, tt.args.orderedEndpoints),
				)
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
					name:     "FooBarResponse",
					children: map[string]struct{}{"FooResponse": {}, "BarResponse": {}},
				},
			},
			want: []string{
				`type FooBarResponse struct {
BarResponse
FooResponse
}`,
			},
		},
		{
			name: "one type, one field: ref type",
			v: models{
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
			},
			want: []string{
				"type FooResponse struct {\nFoo Foo `json:\"foo\"`\n}",
			},
		},
		{
			name: "primitive type",
			v: models{
				"EndpointPoolerMode": model{
					name: "EndpointPoolerMode",
					primitive: fieldType{
						name: "string",
					},
				},
			},
			want: []string{"type EndpointPoolerMode string"},
		},
		{
			name: "primitive type with docstring",
			v: models{
				"Foo": model{
					name:        "Foo",
					description: "foo\nbar\nqux\n",
					primitive: fieldType{
						name: "string",
					},
				},
			},
			want: []string{
				`// Foo foo
// bar
// qux
type Foo string`,
			},
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
		{
			name: "project_id",
			args: args{"project_id"},
			want: "ProjectID",
		},
		{
			name: "connection_uri",
			args: args{"connection_uri"},
			want: "ConnectionURI",
		},
		{
			name: "connection_uris",
			args: args{"connection_uris"},
			want: "ConnectionUris",
		},
		{
			name: "to",
			args: args{"to"},
			want: "To",
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
		RequestBodyStruct           *model
		RequestBodyStructExample    interface{}
		ResponseStruct              *model
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
				RequestBodyStruct:        &model{name: "BranchCreateRequest"},
				RequestBodyStructExample: nil,
				ResponseStruct:           &model{name: "CreatedBranch"},
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
		want *model
	}{
		{
			name: "reference case",
			args: args{
				schema: &openapi3.SchemaRef{Ref: "#/components/schemas/ExplainData"},
			},
			want: &model{name: "ExplainData"},
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
			want: &model{name: "[][]string"},
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
			want: &model{name: "[][][]string"},
		},
		{
			name: "array of arrays of arrays",
			args: args{
				schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						AllOf: openapi3.SchemaRefs{
							{Ref: "#/components/schemas/Foo"},
							{Ref: "#/components/schemas/Bar"},
						},
					},
				},
			},
			want: &model{name: "", children: map[string]struct{}{"Foo": {}, "Bar": {}}, generated: true},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t, tt.want, extractStructFromSchemaRef(tt.args.schema), "extractStructFromSchemaRef(%v)",
					tt.args.schema,
				)
			},
		)
	}
}

func Test_endpointImplementation_generateMethodImplementationTest(t *testing.T) {
	type fields struct {
		Name                           string
		Method                         string
		Route                          string
		Description                    string
		RequestBodyRequires            bool
		RequestBodyStruct              *model
		RequestBodyStructExample       interface{}
		ResponseStruct                 *model
		RequestParametersPath          []field
		ResponsePositivePathExample    interface{}
		ResponsePositivePathStatusCode string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "listProjects",
			fields: fields{
				Name:           "ListProjects",
				Method:         "GET",
				Route:          "/projects",
				Description:    "Retrieves a list of projects for the Neon account",
				ResponseStruct: &model{name: "ProjectsResponse"},
				ResponsePositivePathExample: map[string]interface{}{
					"projects": []map[string]interface{}{
						{
							"id":          "shiny-wind-028834",
							"platform_id": "aws",
							"region_id":   "aws-us-east-2",
							"name":        "shiny-wind-028834",
							"pg_version":  15,
							"locked":      false,
							"created_at":  "2022-11-23T17:42:25Z",
							"updated_at":  "2022-11-23T17:42:25Z",
							"proxy_host":  "us-east-2.aws.neon.tech",
						},
						{
							"id":          "winter-boat-259881",
							"platform_id": "aws",
							"region_id":   "aws-us-east-2",
							"name":        "winter-boat-259881",
							"pg_version":  15,
							"locked":      false,
							"created_at":  "2022-11-23T17:52:25Z",
							"updated_at":  "2022-11-23T17:52:25Z",
							"proxy_host":  "us-east-2.aws.neon.tech",
						},
					},
				},
				ResponsePositivePathStatusCode: "200",
			},
			want: `func Test_client_ListProjects(t *testing.T) {
	deserializeResp := func(s string) ProjectsResponse {
		var v ProjectsResponse
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	tests := []struct {
		name string
		apiKey string
		want ProjectsResponse
		wantErr bool
	}{
		{
			name: "happy path",
			apiKey: "foo",
			want: deserializeResp(endpointResponseExamples["/projects"]["GET"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			apiKey: "invalidApiKey",
			want: ProjectsResponse{},
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
				got, err := c.ListProjects()
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
}`,
		},
		{
			name: "UpdateProject",
			fields: fields{
				Name:                "UpdateProject",
				Method:              "PATCH",
				Route:               "/projects/{project_id}",
				Description:         "Updates the specified project",
				RequestBodyRequires: true,
				RequestBodyStruct:   &model{name: "ProjectUpdateRequest"},
				RequestBodyStructExample: map[string]interface{}{
					"project": map[string]interface{}{"name": "foo"},
				},
				ResponseStruct: &model{name: "ProjectOperations"},
				RequestParametersPath: []field{
					{
						k:           "project_id",
						v:           "string",
						description: "The Neon project ID",
						required:    true,
					},
				},
				ResponsePositivePathExample: map[string]interface{}{
					"project": map[string]interface{}{
						"id":          "shiny-wind-028834",
						"platform_id": "aws",
						"region_id":   "aws-us-east-2",
						"name":        "myproject",
						"provisioner": "k8s-pod",
						"pg_version":  15,
						"locked":      false,
						"created_at":  "2022-11-23T17:42:25Z",
						"updated_at":  "2022-12-04T02:39:25Z",
						"proxy_host":  "us-east-2.aws.neon.tech",
					},
				},
				ResponsePositivePathStatusCode: "200",
			},
			want: `func Test_client_UpdateProject(t *testing.T) {
	deserializeResp := func(s string) ProjectOperations {
		var v ProjectOperations
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
	type args struct {
		projectID string
		cfg ProjectUpdateRequest
	}
	tests := []struct {
		name string
		args args
		apiKey string
		want ProjectOperations
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				projectID: "foo",
				cfg: ProjectUpdateRequest{},
			},
			apiKey: "foo",
			want: deserializeResp(endpointResponseExamples["/projects/{project_id}"]["PATCH"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				projectID: "foo",
				cfg: ProjectUpdateRequest{},
			},
			apiKey: "invalidApiKey",
			want: ProjectOperations{},
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
}`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := endpointImplementation{
					Name:                           tt.fields.Name,
					Method:                         tt.fields.Method,
					Route:                          tt.fields.Route,
					Description:                    tt.fields.Description,
					RequestBodyRequires:            tt.fields.RequestBodyRequires,
					RequestBodyStruct:              tt.fields.RequestBodyStruct,
					RequestBodyStructExample:       tt.fields.RequestBodyStructExample,
					ResponseStruct:                 tt.fields.ResponseStruct,
					RequestParametersPath:          tt.fields.RequestParametersPath,
					ResponsePositivePathExample:    tt.fields.ResponsePositivePathExample,
					ResponsePositivePathStatusCode: tt.fields.ResponsePositivePathStatusCode,
				}
				assert.Equalf(t, tt.want, e.generateMethodImplementationTest(), "generateMethodImplementationTest()")
			},
		)
	}
}
