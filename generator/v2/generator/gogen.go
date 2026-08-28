package generator

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
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
	o.Types, err = newGoTypes(spec.Components)
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

func newMethodsDefinition(spec OpenAPIDefinition) (methods []string, types []string, err error) {
	panic("todo")
}
