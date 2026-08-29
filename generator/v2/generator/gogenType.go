package generator

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"text/template"
)

func newGoTypes(components Components) ([]string, error) {
	repo := new(TypesRepo)

	for _, v := range components.Responses {
		// skip if the response type's name is identical to the schema's name
		if v.Ref != nil && filepath.Base(*v.Ref) == v.xRefName ||
			v.Schema.Ref != nil && filepath.Base(*v.Schema.Ref) == v.xRefName {
			continue
		}
		repo.AddTypeDefinitionInput(v.Schema, v.xRefName)
	}

	for _, v := range components.Schemas {
		repo.AddTypeDefinitionInput(v, v.xRefName)
	}

	return newGoTypesDefinition(repo)
}

func newGoTypesDefinition(repo *TypesRepo) ([]string, error) {
	for !repo.EmptyInputQueue() {
		schema, typeName := repo.dequeue()

		goType, _, err := newGoTypeDefinition(schema, typeName, repo, true)
		if err != nil {
			return nil, err
		}

		var buf = new(strings.Builder)
		buf.WriteString("type ")
		buf.WriteString(typeName)
		buf.WriteString(" ")
		buf.WriteString(goType)
		s := buf.String()
		if schema.Description != "" {
			s = "// " + typeName + " " + schema.Description + "\n" + s
		}
		repo.AddTypeDefinition(s)
	}
	return repo.TypesDefinition(), nil
}

func newGoTypeDefinition(schema OpenAPISchema, typeName string, repo *TypesRepo, topLevel bool) (
	t string, isNillable bool, err error) {
	if typeName == "" {
		typeName = newGoNameFromJsonAttribute(filepath.Base(schema.xRefName))
	}

	switch {
	case schema.Type == "" && schema.AllOf == nil && schema.Ref == nil:
		return "", false, errors.New("unknown type")

	case schema.Ref != nil:
		return filepath.Base(*schema.Ref), false, nil

	case len(schema.Enum) > 0:
		var definition string
		if topLevel {
			definition, err = newGoEnumDefinition(schema, typeName)
		} else {
			repo.AddTypeDefinitionInput(schema, typeName)
			definition = typeName
		}
		return definition, false, err

	case schema.Type == "object":
		if len(schema.Properties) > 0 || len(schema.AllOf) > 0 {
			var definition string
			if topLevel {
				definition, err = newGoStructDefinition(schema, typeName, repo)
			} else {
				repo.AddTypeDefinitionInput(schema, typeName)
				definition = typeName
			}
			return definition, false, err
		}
		return "map[string]any", true, nil

	case schema.Type == "string":
		if schema.Format != nil {
			switch *schema.Format {
			case "date-time":
				return "time.Time", false, nil
			case "uri":
				return "url.URL", false, nil
			case "uuid":
				return "uuid.UUID", false, nil
			}
		}
		return "string", false, nil

	case schema.Type == "integer":
		switch {
		case schema.Format != nil && *schema.Format == "int64":
			return "int64", false, nil

		case schema.Format != nil && *schema.Format == "int32":
			return "int32", false, nil

		case schema.Minimum != nil && *schema.Minimum >= 0:
			switch {
			case schema.Maximum != nil && *schema.Maximum <= 255:
				return "uint8", false, nil
			case schema.Maximum != nil && *schema.Maximum <= 65535:
				return "uint16", false, nil
			case schema.Maximum != nil && *schema.Maximum <= 4294967295:
				return "uint32", false, nil
			default:
				return "uint", false, nil
			}

		default:
			return "int", false, nil
		}

	case schema.Type == "number":
		switch {
		case schema.Format != nil && *schema.Format == "double":
			return "float64", false, nil
		case schema.Maximum != nil && *schema.Maximum <= math.MaxFloat32:
			return "float32", false, nil
		default:
			return "float", false, nil
		}

	case schema.Type == "boolean":
		return "bool", false, nil

	case schema.Type == "array":
		if schema.Items == nil {
			return "", false, errors.New("array type must have items")
		}

		itemType := typeName + "Item"
		itemType, _, err = newGoTypeDefinition(*schema.Items, itemType, repo, false)

		return "[]" + itemType, true, nil

	default:
		return "any", true, nil
	}
}

func newGoStructDefinition(schema OpenAPISchema, typeName string, repo *TypesRepo) (string, error) {
	var buf = new(strings.Builder)
	buf.WriteString("struct {")

	switch {
	case len(schema.Properties) > 0:
		buf.WriteString("\n")

		var requiredProps = make(map[string]struct{}, len(schema.Required))
		for _, propName := range schema.Required {
			requiredProps[propName] = struct{}{}
		}

		for _, prop := range schema.Properties {
			jsonAttrName := prop.xRefName
			_, isRequired := requiredProps[jsonAttrName]

			structAttrName := newGoNameFromJsonAttribute(jsonAttrName)

			if prop.Description != "" {
				buf.WriteString("// ")
				buf.WriteString(structAttrName)
				buf.WriteString(" ")
				buf.WriteString(prop.Description)
				buf.WriteString("\n")
			}

			buf.WriteString(structAttrName)
			buf.WriteString(" ")

			propType, isNillable, err := newGoTypeDefinition(prop, typeName+structAttrName, repo, false)
			if err != nil {
				return "", err
			}

			if !isRequired && !isNillable {
				buf.WriteString("*")
			}
			buf.WriteString(propType)

			buf.WriteString(" `json:\"")
			buf.WriteString(jsonAttrName)
			if !isRequired {
				buf.WriteString(",omitempty")
			}
			buf.WriteString("\"`")

			buf.WriteString("\n")
		}

	case len(schema.AllOf) > 0:
		panic("todo")
	}

	buf.WriteString("}")

	return buf.String(), nil
}

func newGoEnumDefinition(schema OpenAPISchema, typeName string) (string, error) {
	if schema.Type != "string" {
		return "", fmt.Errorf("enum type %s is not supported", schema.Type)
	}

	templ := template.Must(template.New("").Parse(`struct {
	v string
}

func (v {{.TypeName}}) String() string {
	return v.v
}

func (v *{{.TypeName}}) UnmarshalJSON(data []byte) error {
	o, err := New{{.TypeName}}(string(data))
	if err != nil {
		return err
	}
	*v = o
	return nil
}

func (v {{.TypeName}}) MarshalJSON() ([]byte, error) {
	return []byte(v.v), nil
}

var (
	{{- range .EnumValues }}
	{{$.TypeName}}{{.}} = {{$.TypeName}}{"{{.}}"}
	{{- end }}
)

func New{{.TypeName}}(s string) ({{.TypeName}}, error) {
	m := map[string]{{.TypeName}}{
	{{- range .EnumValues }}
		"{{.}}": {{$.TypeName}}{{.}},
	{{- end }}
	}
	v, ok := m[s]
	if !ok {
		return {{.TypeName}}{}, fmt.Errorf("unknown value: %v", s)
	}
	return v, nil
}
`))

	var buf = new(bytes.Buffer)
	err := templ.Execute(buf, struct {
		TypeName   string
		EnumValues []string
	}{
		TypeName:   typeName,
		EnumValues: schema.Enum,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func newGoNameFromJsonAttribute(s string) string {
	var o = new(strings.Builder)
	for _, el := range strings.Split(strings.ToLower(s), "_") {
		switch el {
		case "uuid", "api", "url", "uri", "id":
			el = strings.ToUpper(el)
		default:
			el = strings.ToUpper(el[:1]) + el[1:]
		}
		o.WriteString(el)
	}
	return o.String()
}
