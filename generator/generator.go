package generator

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"text/template"

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

type parameterPath struct {
	k, v, format string
}

func (v parameterPath) canonicalName() string {
	o := ""
	for i, s := range strings.Split(v.k, "_") {
		if i > 0 {
			s = strings.ToUpper(s[:1]) + s[1:]
		}
		o += s
	}

	switch o[len(o)-2:] {
	case "id", "Id", "iD":
		return o[:len(o)-2] + "ID"
	default:
		return o
	}
}

func (v parameterPath) routeElement() string {
	r := v.canonicalName()

	switch v.format {
	case "int64":
		return "strconv.FormatInt(" + r + ", 10)"
	case "int32":
		return "strconv.FormatInt(int64(" + r + "), 10)"
	case "double":
		return "strconv.FormatFloat(" + r + ", 'f', -1, 32)"
	case "float":
		return "strconv.FormatFloat(" + r + ", 'f', -1, 64)"
	case "date-time":
		return r + ".Format(time.RFC3339)"
	case "uuid":
		return r + ".String()"
	default:
		if v.v == "integer" {
			return "strconv.FormatInt(int64(" + r + "), 10)"
		}
		return r
	}
}

func (v parameterPath) argType() string {
	switch v.format {
	case "date-time":
		return "time.Time"
	case "uuid":
		return "uuid.UUID"
	default:
		return v.v
	}
}

type endpointImplementation struct {
	Name                  string
	Method                string
	Route                 string
	Description           string
	RequestBodyStruct     string
	ResponseStruct        string
	RequestParametersPath []parameterPath
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

func (e endpointImplementation) generateFunctionCode() string {
	o := "func (c *Client) " + e.Name
	if e.Description != "" {
		o = e.functionDescription() + "\n" + o
	}

	args := e.inputArgStr()

	reqObj := "nil"

	if e.RequestBodyStruct != "" {
		reqObj = "cfg"
		args += ", cfg " + e.RequestBodyStruct
	}

	o += "(" + args + ") "
	if e.ResponseStruct == "" {
		o += `error {
	return c.requestHandler(c.baseURL+` + e.route() + `, "` + e.Method + `", ` + reqObj + `, nil)
}`
		return o
	}

	o += "(" + e.ResponseStruct + `, error) {
	var v ` + e.ResponseStruct + `
	if err := c.requestHandler(c.baseURL+` + e.route() + `, "` + e.Method + `", ` + reqObj + `, &v); err != nil {
		return nil, err
	}
	return v, nil
}`

	return o
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

func extractStructFromSchemaRef(schema *openapi3.SchemaRef) string {
	if schema.Value != nil && schema.Value.Type == "array" {
		return "[]" + modelNameFromRef(schema.Value.Items.Ref)
	}
	return modelNameFromRef(schema.Ref)
}

type modelField struct {
	Name     string
	Type     string
	Required bool
}

type model struct {
	Name   string
	Fields []modelField
}

type templateInput struct {
	Info      string
	ServerURL string
	Endpoints []string
	Models    []model
}

// Config generator configurations.
type Config struct {
	// OpenAPIReader defines the OpenAPI specs input.
	OpenAPIReader io.Reader

	// PathOutput defines the path to store generated files.
	PathOutput string
}

type openAPISpec struct {
	openapi3.T
}

func modelNameFromRef(s string) string {
	o := strings.Split(s, "/")
	return o[len(o)-1]
}

func implementationNameFromID(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func generateEndpointsImplementationMethods(o openAPISpec) (endpoints []string) {
	for route, p := range o.Paths {
		for httpMethod, ops := range p.Operations() {
			e := endpointImplementation{
				Name:                  implementationNameFromID(ops.OperationID),
				Method:                httpMethod,
				Route:                 route,
				Description:           ops.Description,
				RequestBodyStruct:     "",
				ResponseStruct:        "",
				RequestParametersPath: extractParametersPath(p.Parameters),
			}

			if v, ok := ops.Responses["200"]; ok {
				if v.Value == nil {
					e.ResponseStruct = modelNameFromRef(v.Ref)
				} else {
					if vv, ok := v.Value.Content["application/json"]; ok {
						e.ResponseStruct = extractStructFromSchemaRef(vv.Schema)
					}
				}
			}

			if v := ops.RequestBody; v != nil {
				e.RequestBodyStruct = modelNameFromRef(v.Ref)
				if v.Value != nil {
					if vv, ok := v.Value.Content["application/json"]; ok {
						e.RequestBodyStruct = extractStructFromSchemaRef(vv.Schema)
					}
				}
			}

			endpoints = append(endpoints, e.generateFunctionCode())
		}
	}
	return
}

func extractParametersPath(params openapi3.Parameters) []parameterPath {
	o := make([]parameterPath, len(params))
	for i, p := range params {
		o[i] = parameterPath{
			k:      p.Value.Name,
			v:      p.Value.Schema.Value.Type,
			format: p.Value.Schema.Value.Format,
		}
	}
	return o
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
		if f, err = os.Create(cfg.PathOutput + "/" + strings.Replace(fName, ".templ", "", -1)); err != nil {
			return err
		}

		defer func() { _ = f.Close() }()

		if err := temp.Execute(f, &tempInput); err != nil {
			return err
		}
	}

	return nil
}

func extractSpecs(spec openAPISpec) templateInput {
	if len(spec.Servers) < 1 {
		panic("no server spec found")
	}

	return templateInput{
		Info:      spec.Info.Description,
		ServerURL: spec.Servers[0].URL,
		Endpoints: generateEndpointsImplementationMethods(spec),
	}
}
