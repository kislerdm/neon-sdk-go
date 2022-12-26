package generator

import (
	"bytes"
	"embed"
	"errors"
	"go/format"
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

type field struct {
	k, v, format string
	required     bool
	isInPath     bool
	isInQuery    bool
}

func (v *field) setRequired(b bool) {
	v.required = b
}

func (v field) canonicalName() string {
	return objNameGoConvention(v.k)
}

func objNameGoConvention(s string) string {
	o := ""
	for i, el := range strings.Split(s, "_") {
		if i > 0 {
			el = strings.ToUpper(el[:1]) + el[1:]
		}
		o += el
	}

	switch o[len(o)-2:] {
	case "id", "Id", "iD":
		return o[:len(o)-2] + "ID"
	default:
		return o
	}
}

func objNameGoConventionExport(s string) string {
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
	case "uuid":
		return r + ".String()"
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
	case "uuid":
		return "uuid.UUID"
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

func (v field) hasUUIDArg() bool {
	return v.format == "uuid"
}

func (v field) hasTimeArg() bool {
	return v.format == "date" || v.format == "date-time"
}

type endpointImplementation struct {
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
	reqPointer := ""

	if e.RequestBodyStruct != "" {
		if args != "" {
			args += ", "
		}
		if !e.RequestBodyRequires {
			reqPointer = "*"
		}
		args += "cfg " + reqPointer + e.RequestBodyStruct
		reqObj = "cfg"
	}

	o += "(" + args + ") "
	if e.ResponseStruct == "" {
		o += `error {
	return c.requestHandler(c.baseURL+` + e.route() + `, "` + e.Method + `", ` + reqObj + `, nil)
}`
		return o
	}

	returnStatementUnhappyPath := e.ResponseStruct + "{}"
	if e.ResponseStruct[0] == '[' {
		returnStatementUnhappyPath = "nil"
	}

	o += "(" + e.ResponseStruct + `, error) {
	var v ` + e.ResponseStruct + `
	if err := c.requestHandler(c.baseURL+` + e.route() + `, "` + e.Method + `", ` + reqObj + `, &v); err != nil {
		return ` + returnStatementUnhappyPath + `, err
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
		t := modelNameFromRef(schema.Value.Items.Ref)
		if t == "" {
			t = field{
				k:      "",
				v:      schema.Value.Items.Value.Type,
				format: schema.Value.Items.Value.Format,
			}.argType()
		}
		return "[]" + t
	}
	return modelNameFromRef(schema.Ref)
}

type model struct {
	fields   map[string]*field
	children map[string]struct{}
}

type templateInput struct {
	Info                string
	ServerURL           string
	Endpoints           []string
	EndpointsImportsStr string
	Types               []string
	TypesImportsStr     string
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

func generateEndpointsImplementationMethods(o openAPISpec, dependencies *imports) (endpoints []endpointImplementation) {
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
				if dependencies != nil {
					if p.hasTimeArg() {
						dependencies.set("time")
					}
					if p.hasUUIDArg() {
						dependencies.set("github.com/google/uuid")
					}
				}
			}

			for _, httpCode := range httpCodes {
				if v, ok := ops.Responses[httpCode]; ok {
					if v.Value == nil {
						e.ResponseStruct = modelNameFromRef(v.Ref)
					} else {
						if vv, ok := v.Value.Content["application/json"]; ok {
							e.ResponseStruct = extractStructFromSchemaRef(vv.Schema)
							e.ResponsePositivePathExample = vv.Example
						}
					}
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

type imports map[string]struct{}

func (v imports) set(s string) {
	v[s] = struct{}{}
}

func (v imports) generateImportsStr() string {
	if len(v) == 0 {
		return ""
	}

	o := "import (\n"
	for k := range v {
		o += `"` + k + `"` + "\n"
	}
	o += ")"

	return o
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

	return nil
}

func extractSpecs(spec openAPISpec) templateInput {
	if len(spec.Servers) < 1 {
		panic("no server spec found")
	}

	i := imports{}
	modelsImports := imports{}

	endpoints := generateEndpointsImplementationMethods(spec, &i)
	m := generateModels(spec, &modelsImports)

	endpointsStr := make([]string, len(endpoints))
	models := models{}

	for i, s := range endpoints {
		endpointsStr[i] = s.generateFunctionCode()

		if _, ok := m[s.ResponseStruct]; !ok {
			continue
		}
		models[s.ResponseStruct] = m[s.ResponseStruct]
		for c := range m[s.ResponseStruct].children {
			models.add(c, m[c])
		}
	}

	return templateInput{
		Info:                spec.Info.Description,
		ServerURL:           spec.Servers[0].URL,
		Endpoints:           endpointsStr,
		EndpointsImportsStr: i.generateImportsStr(),
		//Types:               models.generateImportsStr(),
		TypesImportsStr: modelsImports.generateImportsStr(),
	}
}

type models map[string]model

func (v models) add(k string, m model) {
	if _, ok := v[k]; !ok {
		v[k] = m
	}
}

func (v models) addChild(m string, child string) {
	if v[m].children == nil {
		v[m] = model{
			fields:   v[m].fields,
			children: map[string]struct{}{},
		}
	}
	v[m].children[modelNameFromRef(child)] = struct{}{}
}

func (v models) addField(m string, f field) {
	if v[m].fields == nil {
		v[m] = model{
			fields:   map[string]*field{},
			children: v[m].children,
		}
	}
	v[m].fields[f.k] = &f
}

func (v models) generateImportsStr() []string {
	panic("todo")
}

func generateModels(spec openAPISpec, i *imports) models {
	m := models{}

	for k, v := range spec.Components.Responses {
		m.add(k, model{})
		modelsFromSchema(m, k, v.Value.Content["application/json"].Schema)
	}

	for k, v := range spec.Components.Schemas {
		m.add(k, model{})
		modelsFromSchema(m, k, v)
	}

	return m
}

func modelsFromSchema(m models, k string, s *openapi3.SchemaRef) {
	if s.Ref != "" {
		m.addChild(k, s.Ref)
	}

	if v := s.Value; v != nil {
		addFromValue(m, k, v)
	}
}

func addFromValue(m models, k string, v *openapi3.Schema) {
	for _, c := range v.AllOf {
		m.addChild(k, c.Ref)
	}

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
				m.add(field.v, model{})
				addFromValue(m, field.v, property.Value)
			case openapi3.TypeArray:
				m.addChild(k, extractStructFromSchemaRef(property))
			default:
				field.v = property.Value.Type
				field.format = property.Value.Format
			}
		} else {
			if property.Value != nil {
				field.format = property.Value.Format
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
}
