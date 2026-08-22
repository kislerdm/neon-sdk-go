package generator

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates
var templatesFS embed.FS

// CreatorFn function to generate the output artifacts.
type CreatorFn func(fileName string) (io.WriteCloser, error)

func Run(openAPISpec []byte, w CreatorFn) error {
	var spec OpenAPIDefinition
	if err := json.Unmarshal(openAPISpec, &spec); err != nil {
		return fmt.Errorf("could not deserialize API spec: %w", err)
	}

	goGenConfig, err := newGenConfig(spec)
	if err != nil {
		return err
	}

	templates := template.Must(template.ParseFS(templatesFS, "templates/*"))

	// 	generates static artifacts
	for _, fileName := range []string{"go.mod.templ", "doc.go.templ", "error.go.templ"} {
		if err := generateFile(templates, fileName, nil, w); err != nil {
			return err
		}
	}

	// generate SDK methods
	if err := generateFile(templates, "sdk.go.templ", goGenConfig, w); err != nil {
		return err
	}

	return nil
}

func generateFile(t *template.Template, fileName string, data any, w CreatorFn) error {
	fileName = strings.TrimSuffix(fileName, ".templ")
	f, err := w(fileName)
	defer func() { _ = f.Close() }()
	if err != nil {
		return fmt.Errorf("could not open file: %s. %w", fileName, err)
	}
	if err := t.ExecuteTemplate(f, fileName, data); err != nil {
		return fmt.Errorf("could not generate a file %s. %w", fileName, err)
	}
	return nil
}

type genConfig struct {
	// Methods defines the go implementation of the SDK client's methods.
	Methods []string

	// Types defines the go types definition.
	Types []string
}

func newGenConfig(spec OpenAPIDefinition) (genConfig, error) {
	var (
		o   genConfig
		err error
	)
	o.Types, err = newTypesDefinition(spec.Components)
	if err != nil {
		return genConfig{}, err
	}

	var inlineTypes []string
	o.Methods, inlineTypes, err = newMethodsDefinition(spec)
	if err != nil {
		return genConfig{}, err
	}

	o.Types = append(o.Types, inlineTypes...)

	return o, nil
}

func newTypesDefinition(components Components) ([]string, error) {
	var o = make([]string, 0, len(components.Responses)+len(components.Schemas))

	for _, v := range components.Responses {
		// skip if the response type's name is identical to the schema's name
		if v.Ref != nil && filepath.Base(*v.Ref) == v.xRefName {
			continue
		}
		if v.Description != "" && v.Schema.Description == "" {
			v.Schema.Description = v.Description
		}
		schema, err := newGoTypeDefinition(v.xRefName, v.Schema)
		if err != nil {
			return nil, err
		}
		o = append(o, schema)
	}

	for _, v := range components.Schemas {
		schema, err := newGoTypeDefinition(v.xRefName, v)
		if err != nil {
			return nil, err
		}
		o = append(o, schema)
	}

	return o, nil
}

func newGoTypeDefinition(typeName string, schema OpenAPISchema) (string, error) {
	var buf = new(strings.Builder)
	buf.WriteString("type ")
	buf.WriteString(typeName)
	buf.WriteString(" ")

	goType, isStruct, err := newGoType(schema)
	if err != nil {
		return "", err
	}
	if !isStruct {
		buf.WriteString(goType)

	} else {
		buf.WriteString(newGoStructDefinition(schema))
	}

	o := buf.String()

	if schema.Description != "" {
		o = "// " + typeName + " " + schema.Description + "\n" + o
	}

	return o, nil
}

func newMethodsDefinition(spec OpenAPIDefinition) (methods []string, types []string, err error) {
	panic("todo")
}

func newGoType(schema OpenAPISchema) (t string, isStruct bool, err error) {
	switch {
	case schema.Type == "" && (schema.Ref != nil || schema.AllOf != nil):
		return "", false, nil

	case schema.Type == "" && schema.AllOf == nil && schema.Ref == nil:
		return "", false, errors.New("unknown type")

	case schema.AllOf != nil || schema.Ref != nil:
		return "struct", true, nil

	case schema.Type == "object":
		if len(schema.Properties) > 0 {
			return "struct", true, nil
		}
		return "map[string]any", false, nil

	case schema.Type == "string":
		if schema.Format != nil && *schema.Format == "date-time" {
			return "time.Time", false, nil
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
			return "", true, errors.New("array type must have items")
		}
		itemType, structItem, err := newGoType(*schema.Items)
		if err != nil {
			return "", false, err
		}
		if structItem {
			item := *schema.Items
			if item.Ref == nil {
				itemType = newGoStructDefinition(item)
			} else {
				itemType = filepath.Base(*item.Ref)
			}
		}
		return "[]" + itemType, false, nil

	default:
		return "any", false, err
	}
}

func newGoStructDefinition(schema OpenAPISchema) string {
	panic("todo")
}
