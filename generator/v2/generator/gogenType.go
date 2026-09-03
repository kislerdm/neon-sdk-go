package generator

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

type typeDefinitionInput struct {
	schema OpenAPISchema
	name   string
}

type TypesRepo struct {
	typesDefinition []string

	inputQueue []typeDefinitionInput
}

func (v *TypesRepo) AddTypeDefinitionInput(schema OpenAPISchema, name string) {
	if v.inputQueue == nil {
		v.inputQueue = make([]typeDefinitionInput, 0, 100)
	}
	v.inputQueue = append(v.inputQueue, typeDefinitionInput{
		schema: schema,
		name:   name,
	})
}

// Dequeue FIFO dequeue.
func (v *TypesRepo) dequeue() (OpenAPISchema, string) {
	var schema OpenAPISchema
	var name string

	if len(v.inputQueue) > 0 {
		schema, name = v.inputQueue[0].schema, v.inputQueue[0].name
	}

	if len(v.inputQueue) > 1 {
		v.inputQueue = v.inputQueue[1:]
	} else {
		v.inputQueue = nil
	}

	return schema, name
}

func (v *TypesRepo) EmptyInputQueue() bool {
	return len(v.inputQueue) == 0
}

func (v *TypesRepo) AddTypeDefinition(s string) {
	if v.typesDefinition == nil {
		v.typesDefinition = make([]string, 0, 100)
	}
	v.typesDefinition = append(v.typesDefinition, s)
}

func (v *TypesRepo) TypesDefinition() string {
	return strings.Join(v.typesDefinition, "\n")
}

func typesDefinitionInputFromComponents(typesRepo *TypesRepo, components Components) {
	for _, v := range components.Responses {
		// skip if the response type's name is identical to the schema's name
		if (v.Ref != nil && filepath.Base(*v.Ref) == v.xRefName) ||
			(v.Schema.Ref != nil && filepath.Base(*v.Schema.Ref) == v.xRefName) {
			continue
		}
		v.Schema.Description = v.Description
		typesRepo.AddTypeDefinitionInput(v.Schema, v.xRefName)
	}

	for _, v := range components.Schemas {
		typesRepo.AddTypeDefinitionInput(v, v.xRefName)
	}
}

func newGoTypesDefinition(repo *TypesRepo) error {
	for !repo.EmptyInputQueue() {
		schema, typeName := repo.dequeue()

		goType, _, err := newGoTypeDefinition(schema, typeName, repo, true)
		if err != nil {
			return err
		}

		var buf = new(bytes.Buffer)
		buf.WriteString("type ")
		buf.WriteString(typeName)
		buf.WriteString(" ")
		buf.WriteString(goType)
		s := buf.String()
		if schema.Description != "" {
			s = newGoDocString(typeName, schema.Description) + s
		}
		repo.AddTypeDefinition(s)
	}
	return nil
}

func newGoTypeDefinition(schema OpenAPISchema, typeName string, repo *TypesRepo, returnComplexTypeDefinition bool) (
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
		if returnComplexTypeDefinition {
			definition, err = newGoEnumDefinition(schema, typeName)
		} else {
			if repo != nil {
				repo.AddTypeDefinitionInput(schema, typeName)
			}
			definition = typeName
		}
		return definition, false, err

	case schema.Type == "object" || len(schema.AllOf) > 0:
		if len(schema.Properties) > 0 || len(schema.AllOf) > 0 {
			var definition string
			if returnComplexTypeDefinition {
				definition, err = newGoStructDefinition(schema, typeName, repo)
			} else {
				if repo != nil {
					repo.AddTypeDefinitionInput(schema, typeName)
				}
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
		return "float64", false, nil

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
		err := writeGoStructProperties(schema, typeName, repo, buf)
		if err != nil {
			return "", err
		}

	case len(schema.AllOf) > 0:
		// presort sub-schemas so the referenced types come first
		type tmp struct {
			schema OpenAPISchema
			isRef  bool
		}
		var subSchemas = make([]tmp, 0, len(schema.AllOf))
		for _, subSchema := range schema.AllOf {
			if subSchema.Ref == nil && subSchema.Type != "object" {
				return "", fmt.Errorf("%s is unsupported type for allOf clause", subSchema.Type)
			}
			subSchemas = append(subSchemas, tmp{
				schema: subSchema,
				isRef:  subSchema.Ref != nil,
			})
		}
		slices.SortStableFunc(subSchemas, func(a tmp, b tmp) int {
			if !b.isRef && a.isRef {
				return -1
			}
			return 0
		})

		for i, t := range subSchemas {
			subSchema := t.schema
			if subSchema.Ref != nil {
				buf.WriteString("\n")
				buf.WriteString(filepath.Base(*subSchema.Ref))
				if i == len(schema.AllOf)-1 {
					buf.WriteString("\n")
				}

			} else {
				err := writeGoStructProperties(subSchema, typeName, repo, buf)
				if err != nil {
					return "", err
				}
			}
		}
	}

	buf.WriteString("}")

	return buf.String(), nil
}

func writeGoStructProperties(schema OpenAPISchema, typeName string, repo *TypesRepo, buf *strings.Builder) error {
	if len(schema.Properties) > 0 {
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
				buf.WriteString(newGoDocString(structAttrName, prop.Description))
			}

			buf.WriteString(structAttrName)
			buf.WriteString(" ")

			propType, isNillable, err := newGoTypeDefinition(prop, typeName+structAttrName, repo, false)
			if err != nil {
				return err
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
	}
	return nil
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
	{{$.TypeName}}{{.Go}} = {{$.TypeName}}{"{{.Original}}"}
	{{- end }}
)

func New{{.TypeName}}(s string) ({{.TypeName}}, error) {
	m := map[string]{{.TypeName}}{
	{{- range .EnumValues }}
		"{{.Original}}": {{$.TypeName}}{{.Go}},
	{{- end }}
	}
	v, ok := m[s]
	if !ok {
		return {{.TypeName}}{}, fmt.Errorf("unknown value: %v", s)
	}
	return v, nil
}
`))

	type enumTemplateImpute struct {
		Original string
		Go       string
	}
	var enumVals = make([]enumTemplateImpute, 0, len(schema.Enum))
	for _, enumVal := range schema.Enum {
		enumVals = append(enumVals, enumTemplateImpute{
			Original: enumVal,
			Go:       newGoNameFromJsonAttribute(enumVal),
		})
	}

	var buf = new(bytes.Buffer)
	err := templ.Execute(buf, struct {
		TypeName   string
		EnumValues []enumTemplateImpute
	}{
		TypeName:   typeName,
		EnumValues: enumVals,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func newGoNameFromJsonAttribute(s string) string {
	var isReservedWord = func(s string) bool {
		m := map[string]struct{}{
			"uuid": {}, "api": {}, "url": {}, "uri": {}, "id": {},
		}
		_, yes := m[s]
		return yes
	}

	var o = new(strings.Builder)
	s = strings.ToLower(s)
	s = strings.Join(strings.Split(s, "_"), ".")
	s = strings.Join(strings.Split(s, "."), "-")
	s = strings.Join(strings.Split(s, "-"), ":")
	for _, el := range strings.Split(s, ":") {
		switch {
		case isReservedWord(el), len(el) == 1:
			el = strings.ToUpper(el)
		default:
			el = strings.ToUpper(el[:1]) + el[1:]
		}
		o.WriteString(el)
	}
	return o.String()
}

// newGoDocString generates the docstring for a type given its name and description.
func newGoDocString(name string, description string) string {
	var buf = new(strings.Builder)
	if len(description) > 0 {
		description = strings.TrimRight(description, " \n")
		description = strings.TrimLeft(description, " \n")
		for i, el := range strings.Split(description, "\n") {
			el = strings.TrimRight(strings.TrimLeft(el, " "), " ")

			buf.WriteString("//")
			if i == 0 {
				buf.WriteString(" ")
				buf.WriteString(name)
			}

			if len(el) > 0 {
				buf.WriteString(" ")
				buf.WriteString(el)
			}
			buf.WriteString("\n")
		}
	}
	return buf.String()
}
