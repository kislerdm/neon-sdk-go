package generator

import (
	"bytes"
	"fmt"
	"maps"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type typeDescriptor struct {
	// FnArgumentDefinition defines the generated method argument's definition.
	// Example: FnArgumentDefinition = "limit *int" for the optional query argument `limit`.
	FnArgumentDefinition string
	// TransformationToStrFnDefinition defines the generated function to transform the function argument's value
	// to the string type.
	// Example: TransformationToStrFnDefinition = "strconv.FormatInt(int64(*limit), 10)"
	// for the optional query argument `limit`.
	TransformationToStrFnDefinition string
	InQuery                         bool
	Nillable                        bool
	// Name the name as in the openAPI spec.
	Name     string
	Required bool
	GoName   string
}

// newGoMethodsDefinition generates the Go methods definition based on the provided paths and the global parameters.
func newGoMethodsDefinition(paths []OpenAPIPath, globalParameters []OpenAPIParameter, typesRepo *TypesRepo) (string,
	error) {
	// the map of referenced types defined by the openAPI spec. `components.parameters` attribute
	var globalParametersMap = make(map[string]typeDescriptor, len(globalParameters))
	for _, v := range globalParameters {
		if v.Removed() {
			continue
		}
		td, err := newTypeDescriptor(v, v.xRefName, typesRepo)
		if err != nil {
			return "", err
		}
		globalParametersMap[v.xRefName] = td
	}

	var o = new(bytes.Buffer)
	for _, p := range paths {
		var pathParameters = make(map[string]typeDescriptor, len(p.Parameters))
		var pathParamKeys []string
		for _, v := range p.Parameters {
			if v.Removed() {
				continue
			}
			if v.xRefName == "" && v.Ref == nil {
				td, err := newTypeDescriptor(v, p.URLPath+newGoNameFromJsonAttribute(v.Name), typesRepo)
				if err != nil {
					return "", err
				}
				pathParameters[v.Name] = td
				pathParamKeys = append(pathParamKeys, v.Name)
			} else {
				if v.Ref != nil {
					v.xRefName = filepath.Base(*v.Ref)
				}
				pathParameters[v.xRefName] = globalParametersMap[v.xRefName]
				pathParamKeys = append(pathParamKeys, v.xRefName)
			}
		}

		if p.Post != nil && !excludeEndpoint(p.Post) {
			err := processEndpoint(p.URLPath, p.Post, http.MethodPost, globalParametersMap, pathParameters,
				pathParamKeys, typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Get != nil && !excludeEndpoint(p.Get) {
			err := processEndpoint(p.URLPath, p.Get, http.MethodGet, globalParametersMap, pathParameters, pathParamKeys,
				typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Put != nil && !excludeEndpoint(p.Put) {
			err := processEndpoint(p.URLPath, p.Put, http.MethodPut, globalParametersMap, pathParameters, pathParamKeys,
				typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Patch != nil && !excludeEndpoint(p.Patch) {
			err := processEndpoint(p.URLPath, p.Patch, http.MethodPatch, globalParametersMap, pathParameters,
				pathParamKeys, typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Delete != nil && !excludeEndpoint(p.Delete) {
			err := processEndpoint(p.URLPath, p.Delete, http.MethodDelete, globalParametersMap, pathParameters,
				pathParamKeys, typesRepo, o)
			if err != nil {
				return "", err
			}
		}
	}

	return o.String(), nil
}

func excludeEndpoint(op *OpenAPIPathMethod) bool {
	return op.Deprecated && op.Sunset != nil && time.Now().UTC().After(time.Time(*op.Sunset))
}

func newTypeDescriptor(v OpenAPIParameter, typeName string, typesRepo *TypesRepo) (typeDescriptor, error) {
	inQuery := v.In == "query"
	if v.Schema.Type == "" && v.Schema.Ref == nil {
		return typeDescriptor{}, nil
	}

	t, nillable, err := newGoTypeDefinition(v.Schema, typeName, typesRepo, false)
	if err != nil {
		return typeDescriptor{}, err
	}

	argName := methodArgName(v.Name)

	argDefinition := argName + " "
	if !nillable && !v.Required {
		argDefinition += "*"
	}
	argDefinition += t

	return typeDescriptor{
		FnArgumentDefinition:            argDefinition,
		TransformationToStrFnDefinition: newArgTransformationToStrFnDefinition(argName, t, nillable, v.Required),
		InQuery:                         inQuery,
		Nillable:                        nillable,
		Name:                            v.Name,
		GoName:                          argName,
		Required:                        v.Required,
	}, nil
}

func newArgTransformationToStrFnDefinition(name string, goType string, nillable bool, required bool) string {
	if !required && !nillable && goType != "time.Time" {
		name = "*" + name
	}

	switch goType {
	case "bool":
		return "func(v bool) string {if v {return \"true\"}; return \"false\"}(" + name + ")"
	case "int", "int16", "int32":
		return "strconv.FormatInt(int64(" + name + "), 10)"
	case "int64":
		return "strconv.FormatInt(" + name + ", 10)"
	case "float32":
		return "strconv.FormatFloat(float64(" + name + "), 'f', -1, 32)"
	case "float64":
		return "strconv.FormatFloat(" + name + ", 'f', -1, 64)"
	case "uint", "uint8", "uint16", "uint32":
		return "strconv.FormatUint(uint64(" + name + "), 10)"
	case "uint64":
		return "strconv.FormatUint(" + name + ", 10)"
	case "time.Time":
		return name + ".Format(time.RFC3339)"
	case "url.URL":
		return name + ".String()"
	case "string":
		return name
	case "[]string":
		return "strings.Join(" + name + ", \", \")"
	default:
		return "fmt.Sprintf(\"%v\", " + name + ")"
	}
}

func methodArgName(s string) string {
	s = newGoNameFromJsonAttribute(s)
	switch {
	case len(s) == 0:
	case len(s) > 1:
		s = strings.ToLower(s[:1]) + s[1:]
	default:
		s = strings.ToLower(s)
	}
	return s
}

func processEndpoint(urlPath string, op *OpenAPIPathMethod, httpMethod string,
	globalParametersMap map[string]typeDescriptor, pathParameters map[string]typeDescriptor, pathParameterKeys []string,
	typesRepo *TypesRepo, o *bytes.Buffer) error {
	methodName := newMethodName(op.OperationID)
	if methodName == "" {
		return fmt.Errorf("method name cannot be defined for the endpoint: %s %s", httpMethod, urlPath)
	}

	endpointParameters := maps.Clone(pathParameters)
	var endpointParamKeys = slices.Clone(pathParameterKeys)
	for _, v := range op.Parameters {
		if v.Removed() || (v.Deprecated && v.In == "query") {
			continue
		}
		if v.xRefName == "" && v.Ref == nil {
			td, err := newTypeDescriptor(v, methodName+newGoNameFromJsonAttribute(v.Name), typesRepo)
			if err != nil {
				return err
			}
			endpointParameters[v.Name] = td
			endpointParamKeys = append(endpointParamKeys, v.Name)
		} else {
			if v.Ref != nil {
				v.xRefName = filepath.Base(*v.Ref)
			}
			endpointParameters[v.xRefName] = globalParametersMap[v.xRefName]
			endpointParamKeys = append(endpointParamKeys, v.xRefName)
		}
	}

	if op.Description != "" {
		o.WriteString(newGoDocString(methodName, op.Description))
	}

	var requestBodyType string
	var requestBodyNillable bool
	hasRequestBody := httpMethod != http.MethodGet && schemaContentDefined(op.RequestBody.Schema)
	if hasRequestBody {
		var err error
		requestBodyType, requestBodyNillable, err = newGoTypeDefinition(
			op.RequestBody.Schema, methodName+"Cfg", typesRepo, false,
		)
		if err != nil {
			return err
		}
	}

	// define the method's signature: start
	o.WriteString("func (c Client) ")
	o.WriteString(methodName)
	o.WriteString("(")

	var queryParamKeys []string
	for i, key := range endpointParamKeys {
		param := endpointParameters[key]
		o.WriteString(param.FnArgumentDefinition)
		if i < len(endpointParamKeys)-1 {
			o.WriteString(", ")
		}
		if param.InQuery {
			queryParamKeys = append(queryParamKeys, key)
		}
	}

	if hasRequestBody {
		if len(endpointParamKeys) > 0 {
			o.WriteString(", ")
		}
		o.WriteString("cfg ")
		if !op.RequestBody.Required && !requestBodyNillable {
			o.WriteString("*")
		}
		o.WriteString(requestBodyType)
	}

	o.WriteString(") ")

	respType := methodName + "RespObj"
	hasResponseBody := op.Responses.Code204 == nil
	if hasResponseBody {
		resp := op.Responses.Code200
		if resp == nil {
			resp = op.Responses.Code201
		}
		if resp == nil {
			resp = op.Responses.Code202
		}
		if resp == nil {
			return fmt.Errorf("no success response found for operation: %s", op.OperationID)
		}
		hasResponseBody = resp.Ref != nil || schemaContentDefined(resp.Schema)

		if hasResponseBody {
			switch {
			case resp.Ref != nil:
				respType = filepath.Base(*resp.Ref)
			case resp.Schema.Ref != nil:
				respType = filepath.Base(*resp.Schema.Ref)
			default:
				if resp.Schema.Description == "" {
					resp.Schema.Description = resp.Description
				}
				typesRepo.AddTypeDefinitionInput(resp.Schema, respType)
			}
		}

		if hasResponseBody {
			o.WriteString("(")
			o.WriteString(respType)
			o.WriteString(", error) {")
		} else {
			o.WriteString("error {")
		}
	} else {
		o.WriteString("error {")
	}
	o.WriteString("\n")
	// define the method's signature: end

	// define the query string: start
	if len(queryParamKeys) > 0 {
		o.WriteString(`var (
queryElements []string
query string
)
`)

		for _, key := range queryParamKeys {
			param := endpointParameters[key]
			if !param.Required && !param.Nillable {
				o.WriteString("if ")
				o.WriteString(param.GoName)
				o.WriteString(" != nil {\n")
			}
			o.WriteString(`queryElements = append(queryElements, "`)
			o.WriteString(param.Name)
			o.WriteString(`="+`)
			o.WriteString(param.TransformationToStrFnDefinition)
			o.WriteString(")\n")
			if !param.Required && !param.Nillable {
				o.WriteString("}\n")
			}
		}

		o.WriteString(`if len(queryElements) > 0 {
query = "?" + strings.Join(queryElements, "&")
}
`)
	}
	// define the query string: end

	if hasResponseBody {
		o.WriteString("var v ")
		o.WriteString(respType)
		o.WriteString("\n")
	}

	pathCode, err := newPathCode(urlPath, endpointParameters)
	if err != nil {
		return err
	}
	if hasResponseBody {
		o.WriteString("if err := c.requestHandler(c.baseURL")
	} else {
		o.WriteString("return c.requestHandler(c.baseURL")
	}

	if pathCode != "" {
		o.WriteString("+")
		o.WriteString(pathCode)
	}

	if len(queryParamKeys) > 0 {
		o.WriteString("+query")
	}

	o.WriteString(", \"")
	o.WriteString(httpMethod)
	o.WriteString("\", ")

	reqPayload := "nil"
	if hasRequestBody {
		reqPayload = "cfg"
	}
	o.WriteString(reqPayload)

	if hasResponseBody {
		o.WriteString(`, &v); err != nil {
return `)
		o.WriteString(respType)
		o.WriteString(`{}, err
}
return v, nil`)
	} else {
		o.WriteString(", nil)")
	}

	o.WriteString("\n}\n")
	return nil
}

func schemaContentDefined(schema OpenAPISchema) bool {
	return schema.Type != "" || schema.Ref != nil || schema.AllOf != nil
}

func newMethodName(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func newPathCode(path string, parameters map[string]typeDescriptor) (string, error) {
	if path == "" {
		return "", nil
	}

	var o = new(strings.Builder)
	var i int
	newPart := true
	for i < len(path) {
		el := path[i]
		if el == '{' && i < len(path)-1 {
			var param = new(strings.Builder)
			for _, pEl := range path[i+1:] {
				i++
				if pEl == '}' {
					p, ok := parameters[param.String()]
					if !ok {
						return "", fmt.Errorf("parameter %s not found to qualify path", param.String())
					}
					o.WriteString("\"+")
					o.WriteString(p.TransformationToStrFnDefinition)
					newPart = true
					break
				}
				param.WriteRune(pEl)
			}
		} else {
			if newPart {
				if i > 0 {
					o.WriteString("+")
				}
				o.WriteString("\"")
				newPart = false
			}
			o.WriteRune(rune(el))
		}
		i++
	}

	s := o.String()
	if !newPart && !strings.HasSuffix(s, "\"") {
		s += "\""
	}
	return s, nil
}
