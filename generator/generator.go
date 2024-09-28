package generator

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed templates
var templatesFS embed.FS

var (
	templateNameSDK    = []string{"sdk.go.templ", "sdk_test.go.templ"}
	templateNameMock   = []string{"mockhttp.go.templ", "mockhttp_test.go.templ"}
	templateNameStatic = []string{"go.mod.templ", "doc.go.templ", "error.go.templ"}
)

// Config generator configurations.
type Config struct {
	// OpenAPIReader defines the OpenAPI specs input.
	OpenAPIReader io.Reader

	// PathOutput defines the path to store generated files.
	PathOutput string
}

// Run executes code generation using the OpenAPI spec.
func Run(cfg Config) error {
	templates := template.Must(template.ParseFS(templatesFS, "templates/*"))

	specBytes, err := io.ReadAll(cfg.OpenAPIReader)
	if err != nil {
		return errors.New("cannot read OpenAPI spec: " + err.Error())
	}

	var spec openAPISpec
	if err := spec.UnmarshalJSON(specBytes); err != nil {
		return errors.New("cannot parse OpenAPI spec: " + err.Error())
	}

	orderedEndpointRoutes, err := extractOrderedEndpointRoutes(specBytes)
	if err != nil {
		return errors.New("cannot extract ordered list of endpoints from the OpenAPI spec: " + err.Error())
	}
	tempInputSDK, tempInputMock := extractSpecs(spec, orderedEndpointRoutes)

	if err := generateFiles(templates, templateNameSDK, tempInputSDK, cfg.PathOutput); err != nil {
		return fmt.Errorf("could not generate sdk files: %w", err)
	}

	if err := generateFiles(templates, templateNameMock, tempInputMock, cfg.PathOutput); err != nil {
		return fmt.Errorf("could not generate mock files: %w", err)
	}

	if err := generateFiles(templates, templateNameStatic, nil, cfg.PathOutput); err != nil {
		return fmt.Errorf("could not generate static files: %w", err)
	}

	var e error
	if v, _ := strconv.ParseBool(os.Getenv("SKIP_TEST")); !v {
		e = testGeneratedCode(cfg.PathOutput)
	}

	if v, _ := strconv.ParseBool(os.Getenv("SKIP_FORMATTING")); !v {
		e = errors.Join(e, formatGeneratedCode(cfg.PathOutput))
	}

	return e
}

func formatGeneratedCode(p string) error {
	cmd := exec.Command("go", "fmt", ".")
	cmd.Dir = p
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to run go fmt: %w", err)
	}
	return nil
}

func generateFiles(t *template.Template, templateNames []string, data any, p string) error {
	for _, templateName := range templateNames {
		outputFileName := strings.TrimSuffix(templateName, ".templ")
		filePath := path.Join(p, outputFileName)
		f, err := os.Create(filePath)
		defer func() { _ = f.Close() }()
		if err != nil {
			return fmt.Errorf("could not open file: %s. %w", filePath, err)
		}

		if err := t.ExecuteTemplate(f, templateName, data); err != nil {
			return fmt.Errorf("could not generate a file %s. %w", filePath, err)
		}
	}
	return nil
}

func extractSpecs(spec openAPISpec, orderedEndpointRoutes []string) (templateInputSDK, templateInputMock) {
	if len(spec.Servers) < 1 {
		panic("no server spec found")
	}

	m := generateModels(spec)
	endpoints := generateEndpointsImplementationMethods(spec, orderedEndpointRoutes)

	var endpointNames = make([]string, 0, len(endpoints))
	for k, _ := range endpoints {
		endpointNames = append(endpointNames, k)
	}
	slices.Sort(endpointNames)

	endpointsStr := make([]string, len(endpoints))
	endpointsTestStr := make([]string, 0, len(endpoints))
	models := models{}

	mockResponses := map[string]map[string]mockResponse{
		// hardcode based on the api spec because of complexity
		"/projects": {
			"POST": {
				Code: "201",
				Content: `{
		 "project": {
		   "maintenance_starts_at": "2023-01-02T20:03:02.273Z",
		   "id": "string",
		   "platform_id": "string",
		   "region_id": "string",
		   "name": "string",
		   "provisioner": "k8s-pod",
		   "default_endpoint_settings": {
		     "pg_settings": {
		       "additionalProp1": "string",
		       "additionalProp2": "string",
		       "additionalProp3": "string"
		     }
		   },
		   "pg_version": 0,
		   "created_at": "2023-01-02T20:03:02.273Z",
		   "updated_at": "2023-01-02T20:03:02.273Z",
		   "proxy_host": "string"
		 },
		 "connection_uris": [
		   {
		     "connection_uri": "string"
		   }
		 ],
		 "roles": [
		   {
		     "branch_id": "string",
		     "name": "string",
		     "password": "string",
		     "protected": true,
		     "created_at": "2023-01-02T20:03:02.273Z",
		     "updated_at": "2023-01-02T20:03:02.273Z"
		   }
		 ],
		 "databases": [
		   {
		     "id": 0,
		     "branch_id": "string",
		     "name": "string",
		     "owner_name": "string",
		     "created_at": "2023-01-02T20:03:02.273Z",
		     "updated_at": "2023-01-02T20:03:02.273Z"
		   }
		 ],
		 "operations": [
		     {
		       "id": "a07f8772-1877-4da9-a939-3a3ae62d1d8d",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "create_branch",
		       "status": "running",
		       "failures_count": 0,
		       "created_at": "2022-11-08T23:33:16Z",
		       "updated_at": "2022-11-08T23:33:20Z"
		     },
		     {
		       "id": "d8ac46eb-a757-42b1-9907-f78322ee394e",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "start_compute",
		       "status": "finished",
		       "failures_count": 0,
		       "created_at": "2022-11-15T20:02:00Z",
		       "updated_at": "2022-11-15T20:02:02Z"
		     }
		 ],
		 "branch": {
		   "id": "br-wispy-meadow-118737",
		   "project_id": "spring-example-302709",
		   "parent_id": "br-aged-salad-637688",
		   "parent_lsn": "0/1DE2850",
		   "name": "dev2",
		   "current_state": "ready",
		   "created_at": "2022-11-30T19:09:48Z",
		   "updated_at": "2022-12-01T19:53:05Z"
		 },
		 "endpoints": [
		   {
		     "host": "string",
		     "id": "string",
		     "project_id": "string",
		     "branch_id": "string",
		     "autoscaling_limit_min_cu": 0,
		     "autoscaling_limit_max_cu": 0,
		     "region_id": "string",
		     "type": "read_only",
		     "current_state": "init",
		     "pending_state": "init",
		     "settings": {
		       "pg_settings": {
		         "additionalProp1": "string",
		         "additionalProp2": "string",
		         "additionalProp3": "string"
		       }
		     },
		     "pooler_enabled": true,
		     "pooler_mode": "transaction",
		     "disabled": true,
		     "passwordless_access": true,
		     "last_active": "2023-01-02T20:03:02.273Z",
		     "created_at": "2023-01-02T20:03:02.273Z",
		     "updated_at": "2023-01-02T20:03:02.273Z",
		     "proxy_host": "string"
		   }
		 ]
		}`,
			},
		},
		"/projects/{project_id}/branches": {
			"POST": {
				Code: "201",
				Content: `{
		 "branch": {
		   "id": "br-wispy-meadow-118737",
		   "project_id": "spring-example-302709",
		   "parent_id": "br-aged-salad-637688",
		   "parent_lsn": "0/1DE2850",
		   "name": "dev2",
		   "current_state": "ready",
		   "created_at": "2022-11-30T19:09:48Z",
		   "updated_at": "2022-12-01T19:53:05Z"
		 },
		 "endpoints": [
		   {
		     "host": "string",
		     "id": "string",
		     "project_id": "string",
		     "branch_id": "string",
		     "autoscaling_limit_min_cu": 0,
		     "autoscaling_limit_max_cu": 0,
		     "region_id": "string",
		     "type": "read_only",
		     "current_state": "init",
		     "pending_state": "init",
		     "settings": {
		       "pg_settings": {
		         "additionalProp1": "string",
		         "additionalProp2": "string",
		         "additionalProp3": "string"
		       }
		     },
		     "pooler_enabled": true,
		     "pooler_mode": "transaction",
		     "disabled": true,
		     "passwordless_access": true,
		     "last_active": "2023-01-02T20:09:50.004Z",
		     "created_at": "2023-01-02T20:09:50.004Z",
		     "updated_at": "2023-01-02T20:09:50.004Z",
		     "proxy_host": "string"
		   }
		 ],
		 "operations": [
		     {
		       "id": "a07f8772-1877-4da9-a939-3a3ae62d1d8d",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "create_branch",
		       "status": "running",
		       "failures_count": 0,
		       "created_at": "2022-11-08T23:33:16Z",
		       "updated_at": "2022-11-08T23:33:20Z"
		     },
		     {
		       "id": "d8ac46eb-a757-42b1-9907-f78322ee394e",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "start_compute",
		       "status": "finished",
		       "failures_count": 0,
		       "created_at": "2022-11-15T20:02:00Z",
		       "updated_at": "2022-11-15T20:02:02Z"
		     }
		 ]
		}`,
			},
		},
		"/projects/{project_id}/operations": {
			"GET": {
				Code: "200",
				Content: `{
		 "operations": [
		     {
		       "id": "a07f8772-1877-4da9-a939-3a3ae62d1d8d",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "create_branch",
		       "status": "running",
		       "failures_count": 0,
		       "created_at": "2022-11-08T23:33:16Z",
		       "updated_at": "2022-11-08T23:33:20Z"
		     },
		     {
		       "id": "d8ac46eb-a757-42b1-9907-f78322ee394e",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "start_compute",
		       "status": "finished",
		       "failures_count": 0,
		       "created_at": "2022-11-15T20:02:00Z",
		       "updated_at": "2022-11-15T20:02:02Z"
		     }
		 ],
		 "pagination": {
		   "cursor": "string"
		 }
		}`,
			},
		},
		"/projects/{project_id}/branches/{branch_id}/databases": {
			"GET": {
				Code: "200",
				Content: `{
		"databases": [
			{
				"id": 834686,
				"branch_id": "br-aged-salad-637688",
				"name": "main",
				"owner_name": "casey",
				"created_at": "2022-11-30T18:25:15Z",
				"updated_at": "2022-11-30T18:25:15Z"
			},
			{
				"id": 834686,
				"branch_id": "br-aged-salad-637688",
				"name": "mydb",
				"owner_name": "casey",
				"created_at": "2022-10-30T17:14:13Z",
				"updated_at": "2022-10-30T17:14:13Z"
			}
		]}`,
			},
		},
		"/projects/{project_id}/endpoints": {
			"POST": {
				Code: "201",
				Content: `{
		 "endpoint": {
		   "autoscaling_limit_max_cu": 1,
		   "autoscaling_limit_min_cu": 1,
		   "branch_id": "br-proud-paper-090813",
		   "created_at": "2022-12-03T15:37:07Z",
		   "current_state": "init",
		   "disabled": false,
		   "host": "ep-shrill-thunder-454069.us-east-2.aws.neon.tech",
		   "id": "ep-shrill-thunder-454069",
		   "passwordless_access": true,
		   "pending_state": "active",
		   "pooler_enabled": false,
		   "pooler_mode": "transaction",
		   "project_id": "bitter-meadow-966132",
		   "proxy_host": "us-east-2.aws.neon.tech",
		   "region_id": "aws-us-east-2",
		   "settings": {
		     "pg_settings": {}
		   },
		   "type": "read_write",
		   "updated_at": "2022-12-03T15:37:07Z"
		 },
		 "operations": [{
		   "action": "start_compute",
		   "branch_id": "br-proud-paper-090813",
		   "created_at": "2022-12-03T15:37:07Z",
		   "endpoint_id": "ep-shrill-thunder-454069",
		   "failures_count": 0,
		   "id": "874f8bfe-f51d-4c61-85af-a29bea73e0e2",
		   "project_id": "bitter-meadow-966132",
		   "status": "running",
		   "updated_at": "2022-12-03T15:37:07Z"
		 }]
		}`,
			},
		},
		"/projects/{project_id}/endpoints/{endpoint_id}": {
			"DELETE": {
				Code:    "200",
				Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"fd11748e-3c68-458f-b9e3-66d409e3eef0","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			},
			"PATCH": {
				Code:    "200",
				Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"fd11748e-3c68-458f-b9e3-66d409e3eef0","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			},
		},
		"/projects/{project_id}/endpoints/{endpoint_id}/start": {
			"POST": {
				Code:    "200",
				Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"start_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"e061087e-3c99-4856-b9c8-6b7751a253af","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			},
		},
		"/projects/{project_id}/endpoints/{endpoint_id}/suspend": {
			"POST": {
				Code:    "200",
				Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"e061087e-3c99-4856-b9c8-6b7751a253af","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			},
		},
		"/projects/{project_id}/branches/{branch_id}/endpoints": {
			"GET": {
				Code:    "200",
				Content: `{"endpoints":[{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","current_state":"idle","disabled":false,"host":"ep-little-smoke-851426.us-east-2.aws.neon.tech","id":"ep-little-smoke-851426","last_active":"2022-11-23T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-11-30T18:25:21Z"}]}`,
			},
		},
		"/projects/{project_id}/branches/{branch_id}/roles/{role_name}/reset_password": {
			"POST": {
				Code:    "200",
				Content: `{"operations":[{"action":"apply_config","branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T12:58:18Z","endpoint_id":"ep-small-pine-767857","failures_count":0,"id":"6bef07a0-ebca-40cd-9100-7324036cfff2","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-03T12:58:18Z"},{"action":"suspend_compute","branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T12:58:18Z","endpoint_id":"ep-small-pine-767857","failures_count":0,"id":"16b5bfca-4697-4194-a338-d2cdc9aca2af","project_id":"shiny-wind-028834","status":"scheduling","updated_at":"2022-12-03T12:58:18Z"}],"role":{"branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T12:39:39Z","name":"sally","password":"ClfD0aVuK3eK","protected":false,"updated_at":"2022-12-03T12:58:18Z"}}`,
			},
		},
		"/projects/{project_id}/branches/{branch_id}/roles": {
			"POST": {
				Code:    "201",
				Content: `{"operations":[{"action":"apply_config","branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T11:58:29Z","endpoint_id":"ep-small-pine-767857","failures_count":0,"id":"2c2be371-d5ac-4db5-8b68-79f05e8bc287","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-03T11:58:29Z"}],"role":{"branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T11:58:29Z","name":"sally","password":"Onf1AjayKwe0","protected":false,"updated_at":"2022-12-03T11:58:29Z"}}`,
			},
		},
	}

	for i, name := range endpointNames {
		s := endpoints[name]
		endpointsStr[i] = s.generateMethodImplementation()
		if !skipTest(s.Route) {
			endpointsTestStr = append(endpointsTestStr, s.generateMethodImplementationTest())

			if _, ok := mockResponses[s.Route]; !ok {
				mockResponses[s.Route] = map[string]mockResponse{}
			}
			if _, ok := mockResponses[s.Route][s.Method]; !ok {
				mockResponses[s.Route][s.Method] = s.generateMockResponse()
			}
		}

		if s.ResponseStruct != nil {
			if s.ResponseStruct.generated {
				m[s.ResponseStruct.name] = *s.ResponseStruct
			}
			filterModels(m, models, *s.ResponseStruct)
		}
		if s.RequestBodyStruct != nil {
			if s.RequestBodyStruct.generated {
				m[s.RequestBodyStruct.name] = *s.RequestBodyStruct
			}
			filterModels(m, models, *s.RequestBodyStruct)
		}

		// filter models based on the parameters
		for _, param := range slices.Concat(s.RequestParametersPath, s.RequestParametersQuery) {
			mm, ok := m[param.v]
			if ok {
				filterModels(m, models, mm)
			}
		}
	}

	return templateInputSDK{
			ServerURL:                   spec.Servers[0].URL,
			EndpointsImplementation:     endpointsStr,
			Types:                       models.generateCode(),
			EndpointsImplementationTest: endpointsTestStr,
		}, templateInputMock{
			EndpointsResponseExample: mockResponses,
		}
}

func skipTest(route string) bool {
	skipTest := map[string]struct{}{
		"/projects/{project_id}/permissions":                 {},
		"/projects/{project_id}/permissions/{permission_id}": {},
	}
	_, found := skipTest[route]
	return found
}

func filterModels(modelsSource models, output models, m model) {
	name := strings.NewReplacer("[", "", "]", "").Replace(m.name)

	v, ok := modelsSource[name]
	if !ok {
		return
	}

	output[v.name] = v

	for child := range v.children {
		filterModels(modelsSource, output, model{name: child})
	}

	for _, field := range v.fields {
		filterModels(modelsSource, output, model{name: field.v})
	}
}

func testGeneratedCode(p string) error {
	cmd := exec.Command("go", "test", ".")
	cmd.Dir = p
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed test: %w", err)
	}
	return nil
}

type openAPISpec struct {
	openapi3.T
}

type templateInputSDK struct {
	ServerURL                   string
	EndpointsImplementation     []string
	Types                       []string
	EndpointsImplementationTest []string
}

type templateInputMock struct {
	EndpointsResponseExample map[string]map[string]mockResponse
}

type endpointImplementation struct {
	Name                           string
	Method                         string
	Route                          string
	Description                    string
	RequestBodyRequires            bool
	RequestBodyStruct              *model
	RequestBodyStructExample       interface{}
	ResponseStruct                 *model
	RequestParametersPath          []field
	RequestParametersQuery         []field
	ResponsePositivePathExample    interface{}
	ResponsePositivePathStatusCode string
}

func (e endpointImplementation) functionDescription() string {
	if e.Description == "" {
		return ""
	}

	o := "// " + e.Name
	for i, s := range strings.Split(e.Description, "\n") {
		if ss := strings.TrimSpace(s); ss != "" {
			if i > 0 {
				o += "\n// " + ss
			} else {
				o += " " + ss
			}
		}
	}
	return o
}

func (e endpointImplementation) generateMethodImplementation() string {
	var o string
	if e.Description != "" {
		o += e.functionDescription() + "\n"
	}
	o += "func (c Client) " + e.generateMethodHeader() + " {\n"

	reqObj := "nil"
	if e.RequestBodyStruct != nil {
		reqObj = "cfg"
	}

	var query string
	if len(e.RequestParametersQuery) > 0 {
		query = " + query"
		o += e.generateQueryBuilder()
	}

	if e.ResponseStruct == nil {
		return o + `return c.requestHandler(c.baseURL+` + e.route() + query + `, "` + e.Method + `", ` + reqObj + `, nil)
}`
	}

	returnStatementUnhappyPath := e.ResponseStruct.name + "{}"
	if e.ResponseStruct.name == "" || e.ResponseStruct.name[0] == '[' {
		returnStatementUnhappyPath = "nil"
	}

	return o + "	var v " + e.ResponseStruct.name + `
	if err := c.requestHandler(c.baseURL+` + e.route() + query + `, "` + e.Method + `", ` + reqObj + `, &v); err != nil {
		return ` + returnStatementUnhappyPath + `, err
	}
	return v, nil
}`
}

func (e endpointImplementation) generateMethodDefinition() string {
	o := ""
	if e.Description != "" {
		o += e.functionDescription() + "\n"
	}
	return o + e.generateMethodHeader()
}

func (e endpointImplementation) route() string {
	if !strings.Contains(e.Route, "{") {
		return `"` + e.Route + `"`
	}

	o := `"/`
	els := strings.Split(e.Route, "/")[1:]
	for i, el := range els {
		if el[0] != '{' {
			o += el
			if i < len(els)-1 {
				o += "/"
			} else {
				o += `"`
			}
			continue
		}

		for _, p := range e.RequestParametersPath {
			if p.k == el[1:len(el)-1] {
				prefix := `+"/`
				if o[len(o)-1] == '/' {
					prefix = ""
				}
				o += prefix + `"+` + p.routeElement()
				if i < len(els)-1 {
					o += `+"/`
				}
				break
			}
		}
	}
	return o
}

func (e endpointImplementation) inputArgStr() string {
	o := ""
	parameters := e.RequestParametersPath
	parameters = append(parameters, e.RequestParametersQuery...)

	for i, v := range parameters {
		o += v.canonicalName() + " " + v.argType(!v.required)
		if i < len(parameters)-1 {
			o += ", "
		}
	}
	return o
}

func (e endpointImplementation) generateMethodHeader() string {
	args := e.inputArgStr()
	reqPointer := ""

	if e.RequestBodyStruct != nil {
		if args != "" {
			args += ", "
		}
		if !e.RequestBodyRequires {
			reqPointer = "*"
		}
		args += "cfg " + reqPointer + e.RequestBodyStruct.name
	}

	resp := "error"
	if e.ResponseStruct != nil {
		resp = "(" + e.ResponseStruct.name + ", error)"
	}

	return e.Name + "(" + args + ") " + resp
}

type mockResponse struct {
	Code    string
	Content string
}

func (e endpointImplementation) generateMockResponse() mockResponse {
	o, err := json.Marshal(e.ResponsePositivePathExample)
	if err != nil {
		panic(err)
	}
	return mockResponse{
		Code:    e.ResponsePositivePathStatusCode,
		Content: string(o),
	}
}

func (e endpointImplementation) generateMethodImplementationTest() string {
	if e.ResponseStruct == nil {
		return e.generateMethodImplementationTestNoResponseExpected()
	}

	o := `func Test_client_` + e.Name + `(t *testing.T) {
	deserializeResp := func(s string) ` + e.ResponseStruct.name + ` {
		var v ` + e.ResponseStruct.name + `
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
`

	var (
		argsInpt, testInpt, fnInputArgs, pointerTypeCfg string
	)
	inputParameters := e.RequestParametersPath
	inputParameters = append(inputParameters, e.RequestParametersQuery...)
	if len(inputParameters) > 0 || e.RequestBodyStruct != nil {
		argsInpt = "\n\t\targs args"
		testInpt = "\n\t\t\targs: args{\n"
		o += "\ttype args struct {\n"
		for i, v := range inputParameters {
			prf := "\t\t" + v.canonicalName()
			o += fmt.Sprintf("%s %v\n", prf, v.argType(!v.required))
			dummyDate := v.generateDummy()
			if !v.required && !v.isArray {
				dummyDate = wrapIntoPointerGenFn(dummyDate)
			}
			testInpt += fmt.Sprintf("\t\t%s: %v,\n", prf, dummyDate)

			if i > 0 {
				fnInputArgs += ", "
			}
			fnInputArgs += "tt.args." + v.canonicalName()
		}

		if e.RequestBodyStruct != nil {
			cfgInpt := e.RequestBodyStruct.name + "{}"
			if !e.RequestBodyRequires {
				pointerTypeCfg = "*"
				cfgInpt = "nil"
			}
			o += "\t\tcfg " + pointerTypeCfg + e.RequestBodyStruct.name + "\n"
			testInpt += "\t\t\t\tcfg: " + cfgInpt + ","

			if fnInputArgs != "" {
				fnInputArgs += ", "
			}
			fnInputArgs += "tt.args.cfg"
		}

		testInpt += "\n\t\t\t},"
		o += "\t}\n"
	}

	wantUnhappyPath := e.ResponseStruct.name + "{}"
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("slice range out of bound: %#v", e.ResponseStruct)
		}
	}()
	if e.ResponseStruct.name == "" || e.ResponseStruct.name[:2] == "[]" {
		wantUnhappyPath = "nil"
	}

	o += `	tests := []struct {
		name string` + argsInpt + `
		apiKey string
		want ` + e.ResponseStruct.name + `
		wantErr bool
	}{
		{
			name: "happy path",` + testInpt + `
			apiKey: "foo",
			want: deserializeResp(endpointResponseExamples["` + e.Route + `"]["` + e.Method + `"].Content),
			wantErr: false,
		},
		{
			name: "unhappy path",` + testInpt + `
			apiKey: "invalidApiKey",
			want: ` + wantUnhappyPath + `,
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
				got, err := c.` + e.Name + "(" + fnInputArgs + `)
				if (err != nil) != tt.wantErr {
					t.Errorf("` + e.Name + `() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("` + e.Name + `() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}`

	return o
}

func (e endpointImplementation) generateMethodImplementationTestNoResponseExpected() string {
	o := `func Test_client_` + e.Name + `(t *testing.T) {
`

	var (
		argsInpt, testInpt, fnInputArgs, pointerTypeCfg string
	)
	inputParameters := e.RequestParametersPath
	inputParameters = append(inputParameters, e.RequestParametersQuery...)
	if len(inputParameters) > 0 || e.RequestBodyStruct != nil {
		argsInpt = "\n\t\targs args"
		testInpt = "\n\t\t\targs: args{\n"
		o += "\ttype args struct {\n"
		for i, v := range inputParameters {
			prf := "\t\t" + v.canonicalName()
			o += fmt.Sprintf("%s %v\n", prf, v.argType(!v.required))
			dummyDate := v.generateDummy()
			if !v.required && !v.isArray {
				dummyDate = wrapIntoPointerGenFn(dummyDate)
			}
			testInpt += fmt.Sprintf("\t\t%s: %v,\n", prf, dummyDate)

			if i > 0 {
				fnInputArgs += ", "
			}
			fnInputArgs += "tt.args." + v.canonicalName()
		}

		if e.RequestBodyStruct != nil {
			cfgInpt := e.RequestBodyStruct.name + "{}"
			if !e.RequestBodyRequires {
				pointerTypeCfg = "*"
				cfgInpt = "nil"
			}
			o += "\t\tcfg " + pointerTypeCfg + e.RequestBodyStruct.name + "\n"
			testInpt += "\t\t\t\tcfg: " + cfgInpt + ","

			if fnInputArgs != "" {
				fnInputArgs += ", "
			}
			fnInputArgs += "tt.args.cfg"
		}

		testInpt += "\n\t\t\t},"
		o += "\t}\n"
	}

	o += `	tests := []struct {
		name string` + argsInpt + `
		apiKey string
		wantErr bool
	}{
		{
			name: "happy path",` + testInpt + `
			apiKey: "foo",
			wantErr: false,
		},
		{
			name: "unhappy path",` + testInpt + `
			apiKey: "invalidApiKey",
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
				err = c.` + e.Name + "(" + fnInputArgs + `)
				if (err != nil) != tt.wantErr {
					t.Errorf("` + e.Name + `() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}`

	return o
}

// wrapIntoPointerGenFn wraps a dummy value into the generic function
//
//	type dummyType interface {
//		int | int64 | int32 | bool | string | float64 | float32
//	}
//
//	func createPointer[V dummyType](v V) *V {
//		return &v
//	}
func wrapIntoPointerGenFn(v interface{}) string {
	return fmt.Sprintf("createPointer(%v)", v)
}

// generateQueryBuilder generates the function to define the request's query
func (e endpointImplementation) generateQueryBuilder() string {
	if len(e.RequestParametersQuery) == 0 {
		return ""
	}

	o := "\tvar (\n"
	o += "\t\tqueryElements []string\n"
	o += "\t\tquery string\n"
	o += "\t)\n"

	for _, p := range filterRequiredParameters(e.RequestParametersQuery) {
		queryElement := `"` + p.name() + "=\"+" + p.routeElement()
		o += "\tqueryElements = append(queryElements, " + queryElement + ")\n"
	}

	for _, p := range filterOptionalParameters(e.RequestParametersQuery) {
		var (
			ifStatement  string
			queryElement string
			tmpArrayDef  string
		)

		switch p.isArray {
		case false:
			optionalCondition := p.canonicalName() + " != nil "
			ifStatement = "\tif " + optionalCondition + "{\n"
			queryElement = `"` + p.name() + "=\"+" + p.routeElement(true)

		case true:
			optionalCondition := "len(" + p.canonicalName() + ") > 0 "
			ifStatement = "\tif " + optionalCondition + "{\n"

			switch p.v {
			case "string":
				queryElement = `"` + p.name() + "=\"+strings.Join(" + p.canonicalName() + `, ",")`

			default:
				tmpArrName := p.canonicalName() + "Tmp"
				tmpArrayDef = "\t\tvar " + tmpArrName + " = make([]string, len(" + p.canonicalName() + "))\n"
				tmpArrayDef += "\t\tfor i, el := range " + p.canonicalName() + " {\n"
				tmpArrayDef += "\t\t\t" + tmpArrName + "[i] = fmt.Sprintf(\"%v\", el)\n"
				tmpArrayDef += "\t\t}\n"
				queryElement = `"` + p.name() + "=\"+strings.Join(" + tmpArrName + `, ",")`
			}
		}

		o += ifStatement + tmpArrayDef + "\t\tqueryElements = append(queryElements, " + queryElement + ")\n"
		o += "\t}\n"
	}

	o += "\tif len(queryElements) > 0 {\n"
	o += "\t\tquery = \"?\" + strings.Join(queryElements, \"&\")\n"
	o += "\t}\n"
	return o
}

func filterRequiredParameters(v []field) []field {
	var o []field
	for _, p := range v {
		if p.required {
			o = append(o, p)
		}
	}
	return o
}

func filterOptionalParameters(v []field) []field {
	var o []field
	for _, p := range v {
		if !p.required {
			o = append(o, p)
		}
	}
	return o
}

func extractStructFromSchemaRef(schema *openapi3.SchemaRef) *model {
	if schema.Value != nil {
		if schema.Value.Type == "array" {
			t := modelNameFromRef(schema.Value.Items.Ref)
			if t == "" {
				t = field{
					k:      "",
					v:      schema.Value.Items.Value.Type,
					format: schema.Value.Items.Value.Format,
				}.argType()

				switch t {
				case "":
					t = "struct {\n"
					for _, c := range schema.Value.Items.Value.AllOf {
						t += modelNameFromRef(c.Ref) + "\n"
					}
					t += "}"
				case openapi3.TypeArray:
					t = extractStructFromSchemaRef(schema.Value.Items).name
				}
			}
			return &model{name: "[]" + t}
		}
		if len(schema.Value.AllOf) > 0 {
			o := model{
				children: map[string]struct{}{},
			}
			for _, c := range schema.Value.AllOf {
				o.children[modelNameFromRef(c.Ref)] = struct{}{}
			}
			o.generated = true
			return &o
		}
	}
	return &model{name: modelNameFromRef(schema.Ref)}
}

type fieldType struct {
	name, format string
}

func (v fieldType) argType() string {
	return field{v: v.name, format: v.format}.argType()
}

type field struct {
	k, v, format string
	description  string
	isArray      bool
	required     bool
	isInPath     bool
	isInQuery    bool
}

func (v *field) setRequired(b bool) {
	v.required = b
}

func (v *field) setDescription(s string) {
	v.description = s
}

func (v field) name() string {
	return v.k
}

func (v field) canonicalName() string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("objNameGoConvention(%s) panic:\n%v\n", v.k, r)
		}
	}()
	return objNameGoConvention(v.k)
}

func (v field) docString() string {
	if v.description == "" {
		return ""
	}
	return docString(objNameGoConventionExport(v.k), v.description)
}

func objNameGoConvention(s string) string {
	o := ""

	s = replaceSpecialChars(s)
	s = strings.ToLower(s)

	for i, el := range strings.Split(s, "_") {
		el = correctSpecialTagWords(el)
		if i > 0 {
			el = strings.ToUpper(el[:1]) + el[1:]
		}
		o += el
	}

	return o
}

func correctSpecialTagWords(s string) string {
	switch sUp := strings.ToUpper(s); sUp {
	case "ID", "URI", "URL":
		return sUp
	case "IDS", "URIS", "URLS":
		return sUp[:len(sUp)-1] + strings.ToLower(sUp[len(sUp)-1:])
	default:
		return s
	}
}

func replaceSpecialChars(o string) string {
	return strings.NewReplacer(
		"-", "_",
		".", "_",
		" ", "_",
	).Replace(o)
}

func objNameGoConventionExport(s string) string {
	s = objNameGoConvention(s)
	return strings.ToUpper(s[:1]) + s[1:]
}

func (v field) routeElement(withPointer ...bool) string {
	base := func() string {
		r := v.canonicalName()

		if len(withPointer) > 0 && withPointer[0] && v.format != "date-time" && v.format != "date" {
			r = "*" + r
		}

		switch v.format {
		case "date-time", "date":
			return r + ".Format(time.RFC3339)"
		case "int64":
			return "strconv.FormatInt(" + r + ", 10)"
		case "int32":
			return "strconv.FormatInt(int64(" + r + "), 10)"
		case "double":
			return "strconv.FormatFloat(" + r + ", 'f', -1, 64)"
		case "float":
			return "strconv.FormatFloat(" + r + ", 'f', -1, 32)"
		default:
			switch {
			case v.v == "integer":
				return "strconv.FormatInt(int64(" + r + "), 10)"
			case v.v == "boolean":
				varName := v.canonicalName()
				return "func (" + varName + ` bool) string { if ` + varName + ` { return "true" }; return "false" } (` + r + ")"
			// enum case
			case startsWithNumber(v.v) || startsWithCapitalLetter(v.v):
				return "string(" + r + ")"
			}

			return r
		}
	}

	return base()
}

// see the ascii table https://www.asciitable.com/
func startsWithCapitalLetter(v string) bool {
	return v[0] > 64 && v[0] < 91
}

// see the ascii table https://www.asciitable.com/
func startsWithNumber(v string) bool {
	return v[0] > 47 && v[0] < 58
}

func (v field) argType(withPointer ...bool) string {
	switch {
	// DECISION: do not use pointers for 'optional' slices
	//  rationale: slice is a pointer to an array already
	case v.isArray:
		return "[]" + v.argItemType()
	case len(withPointer) > 0 && withPointer[0]:
		return "*" + v.argItemType()
	default:
		return v.argItemType()
	}
}

func (v field) argItemType() string {
	switch v.format {
	case "date-time", "date":
		return "time.Time"
	case "int64", "int32":
		return v.format
	case "double", "number":
		return "float64"
	case "float":
		return "float32"
	default:
		switch v.v {
		case "integer":
			return "int"
		case "boolean":
			return "bool"
		case "number":
			return "float64"
		}
		return v.v
	}
}

func (v field) generateDummy() interface{} {
	switch {
	case v.isArray:
		el := field{v: v.v, format: v.format}
		return "[]" + el.argItemType() + "{" + fmt.Sprintf("%v", el.generateDummyElement()) + "}"
	default:
		return v.generateDummyElement()
	}
}

func (v field) generateDummyElement() interface{} {
	switch v.format {
	case "date-time", "date":
		return "time.Time{}"
	case "int64", "int32", "double", "float":
		return 1
	default:
		switch v.v {
		case "integer":
			return 1
		case "boolean":
			return true
		}
		return "\"foo\""
	}
}

type model struct {
	fields            map[string]*field
	children          map[string]struct{}
	primitive         fieldType
	name, description string
	generated         bool
	isEnum            bool
}

func (m *model) setPrimitiveType(t fieldType) {
	m.primitive = t
}

func (m *model) setDescription(s string) {
	m.description = s
}

func (m model) generateCode() string {
	if m.isEnum {
		return m.generateCodeEnum()
	}

	k := m.name
	if m.primitive.name != "" {
		return m.docString() + "type " + k + " " + m.primitive.argType()
	}

	tmp := m.docString() + "type " + k

	if len(m.fields) == 0 && len(m.children) == 0 {
		return tmp + " map[string]interface{}"
	}

	if len(m.fields) > 0 || len(m.children) > 0 {
		tmp += " struct {\n"
	}

	for _, fieldName := range m.orderedFieldNames() {
		field := m.fields[fieldName]

		tmp += field.docString()

		omitEmpty := ""
		var pointerFlag bool
		if !field.required {
			omitEmpty = ",omitempty"
			pointerFlag = true
		}
		tmp += objNameGoConventionExport(fieldName) + " " + field.argType(pointerFlag) + " `json:\"" + field.k + omitEmpty + "\"`\n"
	}

	if len(m.fields) == 0 {
		var childrenTypes []string
		for k := range m.children {
			childrenTypes = append(childrenTypes, k)
		}
		sort.Strings(childrenTypes)
		tmp += strings.Join(childrenTypes, "\n") + "\n"
	}

	return tmp + "}"
}

func (m *model) docString() string {
	if m.description == "" {
		return ""
	}
	return docString(m.name, m.description)
}

func (m model) orderedFieldNames() []string {
	if len(m.fields) == 0 {
		return nil
	}

	var o = make([]string, len(m.fields))
	var i uint16
	for k := range m.fields {
		o[i] = k
		i++
	}
	sort.Strings(o)
	return o
}

func (m model) generateCodeEnum() string {
	tmp := m.docString() + "type " + m.name + " string\n\n"
	tmp += "const (\n"

	var children = make([]string, len(m.children))
	var i int
	for c := range m.children {
		children[i] = c
		i++
	}
	slices.Sort(children)

	for _, child := range children {
		enumOption := strings.ToUpper(child[:1]) + child[1:]

		enumOption = removeSpecialCharAndMakeCamelCase(enumOption, "-")
		enumOption = removeSpecialCharAndMakeCamelCase(enumOption, "_")

		tmp += m.name + enumOption + " " + m.name + " = \"" + child + "\"\n"
	}

	tmp += ")"
	return tmp
}

func removeSpecialCharAndMakeCamelCase(s string, specialChar string) string {
	els := strings.Split(s, specialChar)
	s = els[0]
	if len(els) > 1 {
		for _, el := range els[1:] {
			s += strings.ToUpper(el[:1]) + el[1:]
		}
	}
	return s
}

func docString(name string, description string) string {
	o := ""
	for i, s := range strings.Split(strings.TrimRight(description, "\n"), "\n") {
		o += "// "
		if i == 0 && name != strings.Split(s, " ")[0] {
			o += name + " "
		}
		o += s + "\n"
	}
	return o
}

func modelNameFromRef(s string) string {
	o := strings.Split(s, "/")
	return o[len(o)-1]
}

func implementationNameFromID(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func generateEndpointsImplementationMethods(
	o openAPISpec, orderedEndpoints []string,
) (endpoints map[string]endpointImplementation) {
	const suffixResponseObject = "RespObj"
	const suffixRequestObject = "ReqObj"

	httpCodes := []string{"200", "201"}
	httpMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPatch,
		http.MethodPut,
		http.MethodDelete,
	}

	endpoints = make(map[string]endpointImplementation)
	for _, route := range orderedEndpoints {
		p := o.Paths.Find(route)
		if p == nil {
			continue
		}

		operations := p.Operations()

		for httpMethod, ops := range operations {
			if !slices.Contains(httpMethods, httpMethod) {
				continue
			}

			e := endpointImplementation{
				Name:        implementationNameFromID(ops.OperationID),
				Method:      httpMethod,
				Route:       route,
				Description: ops.Description,
			}
			// read common parameters for all methods
			pp := p.Parameters
			pp = append(pp, ops.Parameters...)

			for _, p := range extractParameters(pp) {
				if p.isInPath {
					e.RequestParametersPath = append(e.RequestParametersPath, p)
				}
				if p.isInQuery {
					e.RequestParametersQuery = append(e.RequestParametersQuery, p)
				}
			}

			for _, httpCode := range httpCodes {
				if v, ok := ops.Responses[httpCode]; ok {
					if v.Value == nil {
						e.ResponseStruct = &model{name: modelNameFromRef(v.Ref)}
					} else {
						if vv, ok := v.Value.Content["application/json"]; ok {
							e.ResponseStruct = extractStructFromSchemaRef(vv.Schema)
							if e.ResponseStruct.name == "" {
								e.ResponseStruct.name = e.Name + suffixResponseObject
							}

							e.ResponsePositivePathExample = vv.Example
							if e.ResponsePositivePathExample == nil && vv.Schema.Value != nil {
								e.ResponsePositivePathExample = vv.Schema.Value.Example
							}
						}
					}
					e.ResponsePositivePathStatusCode = httpCode
					break
				}
			}

			if v := ops.RequestBody; v != nil {
				e.RequestBodyStruct = &model{name: modelNameFromRef(v.Ref)}
				if v.Value != nil {
					if vv, ok := v.Value.Content["application/json"]; ok {
						e.RequestBodyStruct = extractStructFromSchemaRef(vv.Schema)
						e.RequestBodyRequires = v.Value.Required
						if e.RequestBodyStruct.name == "" {
							e.RequestBodyStruct.name = e.Name + suffixRequestObject
						}
					}
				}
			}

			endpoints[e.Name] = e
		}
	}
	return
}

func extractParameters(params openapi3.Parameters) []field {
	o := make([]field, len(params))
	for i, p := range params {
		o[i] = field{
			k:           p.Value.Name,
			description: p.Value.Description,
			required:    p.Value.Required,
			isInPath:    p.Value.In == openapi3.ParameterInPath,
			isInQuery:   p.Value.In == openapi3.ParameterInQuery,
		}

		if p.Value.Schema.Value != nil {
			tmp := extractItemFromRef(p.Value.Schema)
			o[i].v = tmp.v
			o[i].format = tmp.format

			if p.Value.Schema.Value.Type == "array" {
				item := extractItemFromRef(p.Value.Schema.Value.Items)
				o[i].v = item.v
				o[i].format = item.format
				o[i].isArray = true
			}

		} else {
			o[i].v = modelNameFromRef(p.Value.Schema.Ref)
		}
	}
	return o
}

func extractItemFromRef(v *openapi3.SchemaRef) field {
	var o field
	switch {
	case v.Value != nil:
		o.v = v.Value.Type
		o.format = v.Value.Format
	default:
		o.v = modelNameFromRef(v.Ref)
	}
	return o
}

type models map[string]model

func (v models) add(k string) {
	if _, ok := v[k]; !ok {
		v[k] = model{name: k}
	}
}

func (v models) addChild(m string, child string) {
	if v[m].children == nil {
		tmp := v[m]
		tmp.children = map[string]struct{}{}
		v[m] = tmp
	}
	v[m].children[modelNameFromRef(child)] = struct{}{}
}

func (v models) addField(m string, f field) {
	if v[m].fields == nil {
		tmp := v[m]
		tmp.fields = map[string]*field{}
		v[m] = tmp
	}
	v[m].fields[f.k] = &f
}

func (v models) generateCode() []string {
	o := make([]string, len(v))
	for i, k := range v.orderedNames() {
		o[i] = v[k].generateCode()
	}
	return o
}

func (v models) orderedNames() []string {
	o := make([]string, len(v))
	var i uint8
	for k := range v {
		o[i] = k
		i++
	}
	sort.Strings(o)
	return o
}

func generateModels(spec openAPISpec) models {
	m := models{}

	for k, v := range spec.Components.Responses {
		m.add(k)
		modelsFromSchema(m, k, v.Value.Content["application/json"].Schema)
	}

	for k, v := range spec.Components.Schemas {
		m.add(k)
		modelsFromSchema(m, k, v)
	}

	return m
}

func modelsFromSchema(m models, k string, s *openapi3.SchemaRef) {
	if s.Ref != "" {
		m.addChild(k, s.Ref)
	}

	if v := s.Value; v != nil {
		tmp := m[k]
		tmp.setDescription(v.Description)
		m[k] = tmp

		addFromValue(m, k, v)
	}
}

func addFromValue(m models, k string, v *openapi3.Schema) {
	switch v.Type {
	case "":
		for _, c := range v.AllOf {
			m.addChild(k, c.Ref)
		}
	case openapi3.TypeObject:
		for propertyName, property := range v.Properties {
			field := field{
				k:        propertyName,
				v:        extractStructFromSchemaRef(property).name,
				format:   "",
				required: false,
			}

			if field.v == "" {
				switch property.Value.Type {
				case openapi3.TypeObject:
					field.v = k + objNameGoConventionExport(propertyName)
					m.addChild(k, field.v)
					m.add(field.v)
					modelsFromSchema(m, field.v, property)
				case openapi3.TypeArray:
					m.addChild(k, extractStructFromSchemaRef(property).name)
				default:
					field.v = property.Value.Type
					field.format = property.Value.Format
					field.description = property.Value.Description
				}
			} else {
				if property.Value != nil {
					field.format = property.Value.Format
					field.description = property.Value.Description
				} else {
					m.addChild(k, field.v)
				}
			}

			m.addField(k, field)
		}
		for _, s := range v.Required {
			// in case openAPI json does not define required fields
			// according to the props. definition
			if _, ok := m[k].fields[s]; ok {
				m[k].fields[s].setRequired(true)
			}
		}
	default:
		tmp := m[k]
		tmp.setPrimitiveType(
			fieldType{
				name:   v.Type,
				format: v.Format,
			},
		)
		tmp.setDescription(v.Description)

		if len(v.Enum) > 0 {
			tmp.isEnum = true

			tmp.children = make(map[string]struct{}, len(v.Enum))
			for _, el := range v.Enum {
				child := fmt.Sprintf("%v", el)
				tmp.children[child] = struct{}{}
			}
		}

		m[k] = tmp
	}
}
