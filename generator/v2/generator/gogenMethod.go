package generator

import (
	"bytes"
	"fmt"
	"maps"
	"net/http"
	"path/filepath"
	"strings"
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
}

// newGoMethodsDefinition generates the Go methods definition based on the provided paths and the global parameters.
func newGoMethodsDefinition(paths []OpenAPIPath, globalParameters []OpenAPIParameter, typesRepo *TypesRepo) (string,
	error) {
	// the map of referenced types defined by the openAPI spec. `components.parameters` attribute
	var globalParametersMap = make(map[string]typeDescriptor, len(globalParameters))
	for _, v := range globalParameters {
		err := processParams(globalParametersMap, v, v.xRefName)
		if err != nil {
			return "", err
		}
	}

	var o = new(bytes.Buffer)
	for _, p := range paths {
		var pathParameters = make(map[string]typeDescriptor, len(p.Parameters))
		var pathParamKeys []string
		for _, v := range p.Parameters {
			if v.xRefName == "" {
				err := processParams(pathParameters, v, v.Name)
				if err != nil {
					return "", err
				}
				pathParamKeys = append(pathParamKeys, v.Name)
			} else {
				pathParameters[v.xRefName] = globalParametersMap[v.xRefName]
				pathParamKeys = append(pathParamKeys, v.xRefName)
			}
		}

		if p.Post != nil {
			err := processEndpoint(p.URLPath, p.Post, http.MethodPost, globalParametersMap, pathParameters, typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Get != nil {
			err := processEndpoint(p.URLPath, p.Get, http.MethodGet, globalParametersMap, pathParameters, typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Put != nil {
			err := processEndpoint(p.URLPath, p.Put, http.MethodPut, globalParametersMap, pathParameters, typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Patch != nil {
			err := processEndpoint(p.URLPath, p.Patch, http.MethodPatch, globalParametersMap, pathParameters,
				typesRepo, o)
			if err != nil {
				return "", err
			}
		}

		if p.Delete != nil {
			err := processEndpoint(p.URLPath, p.Delete, http.MethodDelete, globalParametersMap, pathParameters,
				typesRepo, o)
			if err != nil {
				return "", err
			}
		}
	}

	return o.String(), nil
}

func processParams(m map[string]typeDescriptor, v OpenAPIParameter, key string) error {
	inQuery := v.In == "query"
	t, nillable, err := newGoTypeDefinition(v.Schema, filepath.Base(key), nil,
		false)
	if err != nil {
		return err
	}

	argName := methodArgName(v.Name)

	argDefinition := argName + " "
	if !nillable && !v.Required {
		argDefinition += "*"
	}
	argDefinition += t

	m[key] = typeDescriptor{
		FnArgumentDefinition:            argDefinition,
		TransformationToStrFnDefinition: newArgTransformationToStrFnDefinition(argName, t, nillable, v.Required),
		InQuery:                         inQuery,
		Nillable:                        nillable,
		Name:                            v.Name,
		Required:                        v.Required,
	}
	return nil
}

func newArgTransformationToStrFnDefinition(name string, goType string, nillable bool, required bool) string {
	if !required && !nillable && goType != "time.Time" {
		name = "*" + name
	}

	switch goType {
	case "bool":
		return "func(v bool) string {if v return \"true\"; return \"false\"}(" + name + ")"
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
	default:
		return name
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

func processEndpoint(urlPath string, operation *OpenAPIPathMethod, httpMethod string,
	globalParametersMap map[string]typeDescriptor, pathParameters map[string]typeDescriptor, typesRepo *TypesRepo,
	o *bytes.Buffer) error {
	methodName := newGoNameFromJsonAttribute(operation.OperationID)
	if methodName == "" {
		return fmt.Errorf("method name cannot be defined for the endpoint: %s %s", httpMethod, urlPath)
	}

	endpointParameters := maps.Clone(pathParameters)
	var endpointParamKeys []string
	for _, v := range operation.Parameters {
		if v.xRefName == "" {
			err := processParams(endpointParameters, v, v.Name)
			if err != nil {
				return err
			}
			endpointParamKeys = append(endpointParamKeys, v.Name)
		} else {
			endpointParameters[v.xRefName] = globalParametersMap[v.xRefName]
			endpointParamKeys = append(endpointParamKeys, v.xRefName)
		}
	}

	if operation.Description != "" {
		o.WriteString(newGoDocString(methodName, operation.Description))
	}

	// define the method's signature: start
	o.WriteString("func (c Client) ")
	o.WriteString(methodName)
	o.WriteString("(")

	var orationQueryParamKeys []string
	for i, key := range endpointParamKeys {
		param := endpointParameters[key]
		o.WriteString(param.FnArgumentDefinition)
		if i < len(endpointParamKeys)-1 {
			o.WriteString(", ")
		}
		if param.InQuery {
			orationQueryParamKeys = append(orationQueryParamKeys, key)
		}
	}

	if httpMethod != http.MethodGet {
		// TODO: add request's body type as an argument to extend the logic for the request types beyond GET
	}

	o.WriteString(") ")

	respType := methodName + "RespObj"
	if operation.Responses.Code204 == nil {
		resp := operation.Responses.Code200
		if resp == nil {
			resp = operation.Responses.Code201
		}
		if resp == nil {
			return fmt.Errorf("no success response found for operation: %s", operation.OperationID)
		}

		if resp.Ref != nil {
			respType = filepath.Base(*resp.Ref)
		} else {
			if resp.Schema.Description == "" {
				resp.Schema.Description = resp.Description
			}
			typesRepo.AddTypeDefinitionInput(resp.Schema, respType)
		}

		o.WriteString("(")
		o.WriteString(respType)
		o.WriteString(", error) {")
	} else {
		o.WriteString("error {")
	}
	o.WriteString("\n")
	// define the method's signature: end

	// define the query string: start
	if len(orationQueryParamKeys) > 0 {
		o.WriteString(`var (
		queryElements []string
		query string
	)
`)

		for _, key := range orationQueryParamKeys {
			param := endpointParameters[key]
			if !param.Required && !param.Nillable {
				o.WriteString("if ")
				o.WriteString(param.Name)
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

	if operation.Responses.Code204 == nil {
		o.WriteString("var v ")
		o.WriteString(respType)
		o.WriteString("\n")
	}

	pathCode, err := newPathCode(urlPath, endpointParameters)
	if err != nil {
		return err
	}
	if operation.Responses.Code204 == nil {
		o.WriteString("if err := c.requestHandler(c.baseURL")
	} else {
		o.WriteString("return c.requestHandler(c.baseURL")
	}

	if pathCode != "" {
		o.WriteString("+")
		o.WriteString(pathCode)
	}
	o.WriteString(", ")
	o.WriteString(httpMethod)

	reqPayload := "nil"
	if httpMethod != http.MethodGet {
		reqPayload = "cfg"
	}
	o.WriteString(reqPayload)

	if operation.Responses.Code204 == nil {
		o.WriteString(`, &v); err != nil {
return `)
		o.WriteString(respType)
		o.WriteString(`{}, err
}
return v, nil`)
	} else {
		o.WriteString(", nil)")
	}

	o.WriteString("\n}\n\n")
	return nil
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
