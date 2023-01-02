package generator

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"go/format"
	"io"
	"io/fs"
	"os"
	"sort"
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

func testGeneratedCode(p string) error {
	if p, err := os.StartProcess("go test .", nil, &os.ProcAttr{Dir: p}); nil == err {
		if s, err := p.Wait(); nil == err {
			if s.ExitCode() != 0 {
				return errors.New("generated code failed validation")
			}
		}
	}
	return nil
}

type openAPISpec struct {
	openapi3.T
}

type templateInput struct {
	Info                      string
	ServerURL                 string
	EndpointsInterfaceMethods []string
	EndpointsImplementation   []string
	Types                     []string
	EndpointsResponseExample  map[string]map[string]mockResponse
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

func extractSpecs(spec openAPISpec) templateInput {
	if len(spec.Servers) < 1 {
		panic("no server spec found")
	}

	endpoints := generateEndpointsImplementationMethods(spec)
	m := generateModels(spec)

	endpointsStr := make([]string, len(endpoints))
	interfaceMethodsStr := make([]string, len(endpoints))
	models := m
	mockResponses := map[string]map[string]mockResponse{}
	for i, s := range endpoints {
		endpointsStr[i] = s.generateMethodImplementation()
		interfaceMethodsStr[i] = s.generateMethodDefinition()

		if _, ok := mockResponses[s.Route]; !ok {
			mockResponses[s.Route] = map[string]mockResponse{}
		}
		mockResponses[s.Route][s.Method] = s.generateMockResponse()
	}

	return templateInput{
		Info:                      spec.Info.Description,
		ServerURL:                 spec.Servers[0].URL,
		EndpointsInterfaceMethods: interfaceMethodsStr,
		EndpointsImplementation:   endpointsStr,
		EndpointsResponseExample:  mockResponses,
		Types:                     models.generateCode(),
	}
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
