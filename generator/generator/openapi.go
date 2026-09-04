package generator

import (
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type OpenAPIResponse struct {
	// xRefName defines the name of the response if it was defined in the `components.responses` section.
	xRefName    string
	Description string
	Schema      OpenAPISchema
	Ref         *string
}

func (v *OpenAPIResponse) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Description string `json:"description"`
		Content     struct {
			ApplicationJson struct {
				Schema OpenAPISchema `json:"schema"`
			} `json:"application/json"`
		} `json:"content"`
		Ref *string `json:"$ref,omitempty"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	v.Description = tmp.Description
	v.Schema = tmp.Content.ApplicationJson.Schema
	v.Ref = tmp.Ref
	return nil
}

type OpenAPISchema struct {
	// xRefName defines the name of the schema if it was defined in the `components.schemas` section
	//  or from the properties of an `object`.
	xRefName string

	Type        string
	Format      *string
	Enum        []string
	Minimum     *float64
	Maximum     *float64
	Ref         *string
	Description string
	Required    []string
	Properties  []OpenAPISchema
	Items       *OpenAPISchema
	AllOf       []OpenAPISchema
}

func (v *OpenAPISchema) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Type        string                   `json:"type,omitempty"`
		Format      *string                  `json:"format,omitempty"`
		Enum        []string                 `json:"enum,omitempty"`
		Minimum     *float64                 `json:"minimum,omitempty"`
		Maximum     *float64                 `json:"maximum,omitempty"`
		Ref         *string                  `json:"$ref,omitempty"`
		Description string                   `json:"description,omitempty"`
		Required    []string                 `json:"required,omitempty"`
		Properties  map[string]OpenAPISchema `json:"properties,omitempty"`
		Items       *OpenAPISchema           `json:"items,omitempty"`
		AllOf       []OpenAPISchema          `json:"allOf,omitempty"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	v.Type = tmp.Type
	v.Format = tmp.Format
	v.Enum = tmp.Enum
	v.Minimum = tmp.Minimum
	v.Maximum = tmp.Maximum
	v.Ref = tmp.Ref
	v.Description = tmp.Description
	v.Required = tmp.Required
	v.Items = tmp.Items
	v.AllOf = tmp.AllOf

	if len(tmp.Properties) > 0 {
		v.Properties = make([]OpenAPISchema, 0, len(tmp.Properties))
		for _, k := range sortMapKeys(tmp.Properties) {
			vv := tmp.Properties[k]
			vv.xRefName = k
			v.Properties = append(v.Properties, vv)
		}
	}

	return nil
}

type OpenAPIParameter struct {
	// xRefName defines the name of the parameter if it was defined in the `components.parameters` section.
	xRefName    string
	Name        string        `json:"name"`
	In          string        `json:"in"`
	Required    bool          `json:"required"`
	Description string        `json:"description"`
	Schema      OpenAPISchema `json:"schema"`
	Ref         *string       `json:"$ref,omitempty"`
}

type Components struct {
	Responses  []OpenAPIResponse
	Schemas    []OpenAPISchema
	Parameters []OpenAPIParameter
	// parameters duplicate the parameters for lookup when generating SDK operation-methods.
	parameters map[string]OpenAPIParameter
}

func (v *Components) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Responses  map[string]OpenAPIResponse  `json:"responses,omitempty"`
		Schemas    map[string]OpenAPISchema    `json:"schemas,omitempty"`
		Parameters map[string]OpenAPIParameter `json:"parameters,omitempty"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	v.Responses = make([]OpenAPIResponse, 0, len(tmp.Responses))
	for _, k := range sortMapKeys(tmp.Responses) {
		vv := tmp.Responses[k]
		vv.xRefName = k
		v.Responses = append(v.Responses, vv)
	}

	v.Schemas = make([]OpenAPISchema, 0, len(tmp.Schemas))
	for _, k := range sortMapKeys(tmp.Schemas) {
		vv := tmp.Schemas[k]
		vv.xRefName = k
		v.Schemas = append(v.Schemas, vv)
	}

	v.Parameters = make([]OpenAPIParameter, 0, len(tmp.Parameters))
	for _, k := range sortMapKeys(tmp.Parameters) {
		vv := tmp.Parameters[k]
		vv.xRefName = k
		v.Parameters = append(v.Parameters, vv)
	}

	v.parameters = tmp.Parameters

	return nil
}

type OpenAPIPathMethodResponses struct {
	Code200 *OpenAPIResponse `json:"200,omitempty"`
	Code201 *OpenAPIResponse `json:"201,omitempty"`
	Code202 *OpenAPIResponse `json:"202,omitempty"`
	Code204 *OpenAPIResponse `json:"204,omitempty"`
	Default OpenAPIResponse  `json:"default"`
}

type OpenAPIPathMethodRequestBody struct {
	Required bool
	Schema   OpenAPISchema
}

func (v *OpenAPIPathMethodRequestBody) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Required bool `json:"required"`
		Content  struct {
			ApplicationJson struct {
				Schema OpenAPISchema `json:"schema"`
			} `json:"application/json"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.Required = tmp.Required
	v.Schema = tmp.Content.ApplicationJson.Schema
	return nil
}

type OpenAPIPathMethod struct {
	Description string                       `json:"description"`
	OperationID string                       `json:"operationId"`
	Parameters  []OpenAPIParameter           `json:"parameters,omitempty"`
	Responses   OpenAPIPathMethodResponses   `json:"responses"`
	RequestBody OpenAPIPathMethodRequestBody `json:"requestBody"`
	Deprecated  bool                         `json:"deprecated,omitempty"`
	Sunset      *Date                        `json:"x-sunset,omitempty"`
}

type OpenAPIPath struct {
	URLPath string

	Parameters []OpenAPIParameter `json:"parameters,omitempty"`
	Get        *OpenAPIPathMethod `json:"get,omitempty"`
	Post       *OpenAPIPathMethod `json:"post,omitempty"`
	Delete     *OpenAPIPathMethod `json:"delete,omitempty"`
	Put        *OpenAPIPathMethod `json:"put,omitempty"`
	Patch      *OpenAPIPathMethod `json:"path,omitempty"`
}

type OpenAPIDefinition struct {
	ServerURL  string
	Paths      []OpenAPIPath
	Components Components
}

type Date time.Time

func (v *Date) UnmarshalJSON(data []byte) error {
	t, err := time.Parse("2006-01-02", strings.Trim(string(data), "\""))
	if err != nil {
		return err
	}
	*v = Date(t)
	return nil
}

type tmpOpenAPIPath struct {
	OpenAPIPath
	Deprecated bool  `json:"deprecated,omitempty"`
	Sunset     *Date `json:"x-sunset,omitempty"`
}

func (v *OpenAPIDefinition) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Servers []struct {
			URL string `json:"url"`
		} `json:"servers"`
		Paths      map[string]tmpOpenAPIPath `json:"paths"`
		Components Components                `json:"components"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.Paths = make([]OpenAPIPath, 0, len(tmp.Paths))
	for _, k := range sortMapKeys(tmp.Paths) {
		vv := tmp.Paths[k]
		if vv.Deprecated && vv.Sunset != nil && time.Now().After(time.Time(*vv.Sunset)) {
			continue
		}
		vv.URLPath = k
		v.Paths = append(v.Paths, vv.OpenAPIPath)
	}
	v.Components = tmp.Components
	if len(tmp.Servers) > 0 {
		v.ServerURL = tmp.Servers[0].URL
	}
	return nil
}

func sortMapKeys[E ~map[string]V,
	V OpenAPISchema | OpenAPIParameter | OpenAPIResponse | OpenAPIPath | tmpOpenAPIPath](v E) []string {
	var o = make([]string, 0, len(v))
	for k := range v {
		o = append(o, k)
	}
	sort.Strings(o)
	return o
}
