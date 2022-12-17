package generator

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"text/template"

	"github.com/swaggest/openapi-go/openapi3"
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

type templateInput struct {
	Info *string
}

// Config generator configurations.
type Config struct {
	// OpenAPIReader defines the OpenAPI specs input.
	OpenAPIReader io.Reader

	// PathOutput defines the path to store generated files.
	PathOutput string
}

type openAPISpec struct {
	openapi3.Spec
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
	return templateInput{
		Info: spec.Info.Description,
	}
}
