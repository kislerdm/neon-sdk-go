package generator

import (
	_ "embed"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
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
		RequestParametersPath map[string]string
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
		return nil, err
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
				RequestParametersPath: map[string]string{"project_id": "string"},
			},
			want: `// GetProject Retrieves information about the specified project
func (c *Client) GetProject(projectID string) (ProjectsResponse, error) {
	var v ProjectsResponse
	if err := c.requestHandler(c.baseURL+"/projects/"+projectID, "GET", nil, &v); err != nil {
		return nil, err
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
