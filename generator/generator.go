package generator

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	//go:embed templates
	templates   embed.FS
	templateGen map[string]*template.Template
)

// read and parse templates
func init() {
	tmplFiles, err := fs.ReadDir(templates, "templates")
	if err != nil {
		panic(err)
	}

	if templateGen == nil {
		templateGen = make(map[string]*template.Template, len(tmplFiles))
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(templates, "templates/"+tmpl.Name())
		if err != nil {
			panic(err)
		}

		templateGen[tmpl.Name()] = pt
	}
}

// Config generator configurations.
type Config struct {
	// OpenAPIReader defines the OpenAPI specs input.
	OpenAPIReader io.Reader

	// PathOutput defines the path to store generated files.
	PathOutput string
}

// Run executes code generation using the OpenAPI spec.
func Run(cfg Config) error {
	specBytes, err := io.ReadAll(cfg.OpenAPIReader)
	if err != nil {
		return errors.New("cannot read OpenAPI spec: " + err.Error())
	}

	var spec openAPISpec
	if err := spec.UnmarshalJSON(specBytes); err != nil {
		return errors.New("cannot parse OpenAPI spec: " + err.Error())
	}

	tempInput := extractSpecs(spec)

	var f io.WriteCloser

	for fName, temp := range templateGen {
		fName = strings.Replace(fName, ".templ", "", -1)
		if f, err = os.Create(cfg.PathOutput + "/" + fName); err != nil {
			return err
		}

		defer func() { _ = f.Close() }()

		var buf bytes.Buffer
		if err := temp.Execute(&buf, &tempInput); err != nil {
			return err
		}

		o := buf.Bytes()
		if strings.HasSuffix(fName, ".go") {
			o, err = format.Source(buf.Bytes())
			if err != nil {
				return err
			}
		}

		if _, err := f.Write(o); err != nil {
			return err
		}
	}
	return testGeneratedCode(cfg.PathOutput)
}

func extractSpecs(spec openAPISpec) templateInput {
	if len(spec.Servers) < 1 {
		panic("no server spec found")
	}

	endpoints := generateEndpointsImplementationMethods(spec)
	m := generateModels(spec)

	endpointsStr := make([]string, len(endpoints))
	endpointsTestStr := make([]string, len(endpoints))
	interfaceMethodsStr := make([]string, len(endpoints))
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

	for i, s := range endpoints {
		endpointsStr[i] = s.generateMethodImplementation()
		interfaceMethodsStr[i] = s.generateMethodDefinition()
		endpointsTestStr[i] = s.generateMethodImplementationTest()

		if _, ok := mockResponses[s.Route]; !ok {
			mockResponses[s.Route] = map[string]mockResponse{}
		}
		if _, ok := mockResponses[s.Route][s.Method]; !ok {
			mockResponses[s.Route][s.Method] = s.generateMockResponse()
		}

		filterModels(m, models, s.ResponseStruct)
		filterModels(m, models, s.RequestBodyStruct)
	}

	return templateInput{
		Info:                        spec.Info.Description,
		ServerURL:                   spec.Servers[0].URL,
		EndpointsInterfaceMethods:   interfaceMethodsStr,
		EndpointsImplementation:     endpointsStr,
		EndpointsImplementationTest: endpointsTestStr,
		EndpointsResponseExample:    mockResponses,
		Types:                       models.generateCode(),
	}
}

func filterModels(modelsSource models, output models, name string) {
	name = strings.NewReplacer("[", "", "]", "").Replace(name)

	v, ok := modelsSource[name]
	if !ok {
		return
	}

	output[v.name] = v

	for child := range v.children {
		filterModels(modelsSource, output, child)
	}

	for _, field := range v.fields {
		filterModels(modelsSource, output, field.v)
	}
}

func testGeneratedCode(p string) error {
	cmd := exec.Command("go", "test", ".")
	cmd.Dir = p
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		return errors.New("failed test")
	}
	return nil
}

type openAPISpec struct {
	openapi3.T
}

type templateInput struct {
	Info                        string
	ServerURL                   string
	EndpointsInterfaceMethods   []string
	EndpointsImplementation     []string
	EndpointsImplementationTest []string
	Types                       []string
	EndpointsResponseExample    map[string]map[string]mockResponse
}

type endpointImplementation struct {
	Name                           string
	Method                         string
	Route                          string
	Description                    string
	RequestBodyRequires            bool
	RequestBodyStruct              string
	RequestBodyStructExample       interface{}
	ResponseStruct                 string
	RequestParametersPath          []field
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
	o := "func (c *client) " + e.generateMethodHeader() + " {\n"

	reqObj := "nil"
	if e.RequestBodyStruct != "" {
		reqObj = "cfg"
	}

	if e.ResponseStruct == "" {
		return o + `return c.requestHandler(c.baseURL+` + e.route() + `, "` + e.Method + `", ` + reqObj + `, nil)
}`
	}

	returnStatementUnhappyPath := e.ResponseStruct + "{}"
	if e.ResponseStruct[0] == '[' {
		returnStatementUnhappyPath = "nil"
	}

	return o + "	var v " + e.ResponseStruct + `
	if err := c.requestHandler(c.baseURL+` + e.route() + `, "` + e.Method + `", ` + reqObj + `, &v); err != nil {
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
	for i, v := range e.RequestParametersPath {
		o += v.canonicalName() + " " + v.argType()
		if i < len(e.RequestParametersPath)-1 {
			o += ", "
		}
	}
	return o
}

func (e endpointImplementation) generateMethodHeader() string {
	args := e.inputArgStr()
	reqPointer := ""

	if e.RequestBodyStruct != "" {
		if args != "" {
			args += ", "
		}
		if !e.RequestBodyRequires {
			reqPointer = "*"
		}
		args += "cfg " + reqPointer + e.RequestBodyStruct
	}

	resp := "error"
	if e.ResponseStruct != "" {
		resp = "(" + e.ResponseStruct + ", error)"
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
	o := `func Test_client_` + e.Name + `(t *testing.T) {
	deserializeResp := func(s string) ` + e.ResponseStruct + ` {
		var v ` + e.ResponseStruct + `
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			panic(err)
		}
		return v
	}
`

	var (
		argsInpt, testInpt, fnInputArgs, pointerCfg, pointerTypeCfg string
	)
	if len(e.RequestParametersPath) > 0 || e.RequestBodyStruct != "" {
		argsInpt = "\n\t\targs args"
		testInpt = "\n\t\t\targs: args{\n"
		o += "\ttype args struct {\n"
		for i, v := range e.RequestParametersPath {
			prf := "\t\t" + v.canonicalName()
			o += fmt.Sprintf("%s %v\n", prf, v.argType())
			testInpt += fmt.Sprintf("\t\t%s: %v,\n", prf, v.generateDummy())

			if i > 0 {
				fnInputArgs += ", "
			}
			fnInputArgs += "tt.args." + v.canonicalName()
		}

		if e.RequestBodyStruct != "" {
			if !e.RequestBodyRequires {
				pointerCfg = "&"
				pointerTypeCfg = "*"
			}
			o += "\t\tcfg " + pointerTypeCfg + e.RequestBodyStruct + "\n"
			testInpt += "\t\t\t\tcfg: " + pointerCfg + e.RequestBodyStruct + `{},`

			if fnInputArgs != "" {
				fnInputArgs += ", "
			}
			fnInputArgs += "tt.args.cfg"
		}

		testInpt += "\n\t\t\t},"
		o += "\t}\n"
	}

	wantUnhappyPath := e.ResponseStruct + "{}"
	if e.ResponseStruct[:2] == "[]" {
		wantUnhappyPath = "nil"
	}

	o += `	tests := []struct {
		name string` + argsInpt + `
		apiKey string
		want ` + e.ResponseStruct + `
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
				c, err := NewClient(WithAPIKey(tt.apiKey), WithHTTPClient(NewMockHTTPClient()))
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

func extractStructFromSchemaRef(schema *openapi3.SchemaRef) string {
	if schema.Value != nil && schema.Value.Type == "array" {
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
				t = extractStructFromSchemaRef(schema.Value.Items)
			}
		}
		return "[]" + t
	}
	return modelNameFromRef(schema.Ref)
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

func (v field) canonicalName() string {
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
		if i > 0 {
			el = strings.ToUpper(el[:1]) + el[1:]
		}
		o += el
	}

	switch o[len(o)-2:] {
	case "id", "Id":
		return o[:len(o)-2] + "ID"
	}

	switch o[len(o)-3:] {
	case "url", "Url":
		return o[:len(o)-3] + "URL"
	case "uri", "Uri":
		return o[:len(o)-3] + "URI"
	default:
		return o
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

func (v field) routeElement() string {
	r := v.canonicalName()

	switch v.format {
	case "int64":
		return "strconv.FormatInt(" + r + ", 10)"
	case "int32":
		return "strconv.FormatInt(int64(" + r + "), 10)"
	case "double":
		return "strconv.FormatFloat(" + r + ", 'f', -1, 64)"
	case "float":
		return "strconv.FormatFloat(" + r + ", 'f', -1, 32)"
	case "date-time", "date":
		return r + ".Format(time.RFC3339)"
	default:
		switch v.v {
		case "integer":
			return "strconv.FormatInt(int64(" + r + "), 10)"
		case "boolean":
			return "func (" + r + ` bool) string { if r { return "true" }; return "false" } (` + r + ")"
		}
		return r
	}
}

func (v field) argType() string {
	switch v.format {
	case "date-time", "date":
		return "time.Time"
	case "int64", "int32":
		return v.format
	case "double":
		return "float64"
	case "float":
		return "float32"
	default:
		switch v.v {
		case "integer":
			return "int"
		case "boolean":
			return "bool"
		}
		return v.v
	}
}

func (v field) generateDummy() interface{} {
	if v.v[:2] == "[]" {
		return []interface{}{field{v: v.v[2:], format: v.format}.generateDummy()}
	}

	switch v.format {
	case "date-time", "date":
		return time.Time{}
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
}

func (m *model) setPrimitiveType(t fieldType) {
	m.primitive = t
}

func (m *model) setDescription(s string) {
	m.description = s
}

func (m model) generateCode() string {
	k := m.name
	if m.primitive.name != "" {
		return m.docString() + "type " + k + " " + m.primitive.argType()
	}

	tmp := m.docString() + "type " + k + " struct {"

	if len(m.fields) > 0 || len(m.children) > 0 {
		tmp += "\n"
	}

	for fieldName, field := range m.fields {
		tmp += field.docString()

		omitEmpty := ""
		if !field.required {
			omitEmpty = ",omitempty"
		}
		tmp += objNameGoConventionExport(fieldName) + " " + field.argType() + " `json:\"" + field.k + omitEmpty + "\"`\n"
	}

	if len(m.fields) == 0 {
		for k := range m.children {
			tmp += k + "\n"
		}
	}

	return tmp + "}"
}

func (m *model) docString() string {
	if m.description == "" {
		return ""
	}
	return docString(m.name, m.description)
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

func generateEndpointsImplementationMethods(o openAPISpec) (endpoints []endpointImplementation) {
	httpCodes := []string{"200", "201"}
	for route, p := range o.Paths {
		for httpMethod, ops := range p.Operations() {
			e := endpointImplementation{
				Name:              implementationNameFromID(ops.OperationID),
				Method:            httpMethod,
				Route:             route,
				Description:       ops.Description,
				RequestBodyStruct: "",
				ResponseStruct:    "",
			}

			pp := p.Parameters
			pp = append(pp, ops.Parameters...)
			for _, p := range extractParameters(pp) {
				if !p.isInPath {
					continue
				}
				e.RequestParametersPath = append(e.RequestParametersPath, p)
			}

			for _, httpCode := range httpCodes {
				if v, ok := ops.Responses[httpCode]; ok {
					if v.Value == nil {
						e.ResponseStruct = modelNameFromRef(v.Ref)
					} else {
						if vv, ok := v.Value.Content["application/json"]; ok {
							e.ResponseStruct = extractStructFromSchemaRef(vv.Schema)
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
				e.RequestBodyStruct = modelNameFromRef(v.Ref)
				if v.Value != nil {
					if vv, ok := v.Value.Content["application/json"]; ok {
						e.RequestBodyStruct = extractStructFromSchemaRef(vv.Schema)
						e.RequestBodyRequires = v.Value.Required
					}
				}
			}

			endpoints = append(endpoints, e)
		}
	}
	return
}

func extractParameters(params openapi3.Parameters) []field {
	o := make([]field, len(params))
	for i, p := range params {
		o[i] = field{
			k:         p.Value.Name,
			v:         p.Value.Schema.Value.Type,
			format:    p.Value.Schema.Value.Format,
			required:  p.Value.Required,
			isInPath:  p.Value.In == openapi3.ParameterInPath,
			isInQuery: p.Value.In == openapi3.ParameterInQuery,
		}
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
				v:        extractStructFromSchemaRef(property),
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
					m.addChild(k, extractStructFromSchemaRef(property))
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
		m[k] = tmp
	}
}
