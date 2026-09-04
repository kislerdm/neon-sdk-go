package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	goModFile = `module github.com/kislerdm/neon-sdk-go

go 1.24
`
	sdkFile = `// Package sdk to communicate to the Neon Postgres SaaS Platform.
// Find more about the service: https://neon.com/docs/reference/api/get-started
// Author: Dmitry Kisler <https://www.dkisler.com>

package sdk

import (
	"bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "reflect"
	"strconv"
	"strings"
	"time"
)

// Error API error.
type Error struct {
	HTTPCode int
	errorResp
}

func (e Error) Error() string {
	return "[HTTP Code: " + strconv.Itoa(e.HTTPCode) + "][Error Code: " + e.Code + "] " + e.Message
}

func (e Error) httpResp() *http.Response {
	o, _ := json.Marshal(e.errorResp)
	return &http.Response{
		Status:        e.Code,
		StatusCode:    e.HTTPCode,
		Body:          io.NopCloser(bytes.NewReader(o)),
		ContentLength: int64(len(o)),
	}
}

type errorResp struct {
	Code    string ` + "`json:\"code\"`" +
		`
	Message string ` + "`json:\"message\"`" +
		`
}

func convertErrorResponse(res *http.Response) error {
	var v errorResp
	buf, err := io.ReadAll(res.Body)
	defer func() { _ = res.Body.Close() }()
	if err != nil {
		return Error{
			HTTPCode: res.StatusCode,
			errorResp: errorResp{
				Message: "cannot read response bytes",
			},
		}
	}
	if err := json.Unmarshal(buf, &v); err != nil {
		return Error{
			HTTPCode: res.StatusCode,
			errorResp: errorResp{
				Message: err.Error(),
			},
		}
	}
	return Error{
		HTTPCode:  res.StatusCode,
		errorResp: v,
	}
}

// NewClient initialised the Client to communicate to the Neon Platform.
func NewClient(cfg Config) (*Client, error) {
    if cfg.Key == "" {
		return nil, errors.New(
			"authorization key must be provided: https://neon.com/docs/reference/api/get-started#get-an-api-key",
		)
	}

	c := &Client{
        baseURL: baseURL,
        cfg: cfg,
    }

    if c.cfg.HTTPClient == nil {
        c.cfg.HTTPClient = &http.Client{Timeout: defaultTimeout}
    }

	return c, nil
}

// Config defines the client's configuration.
type Config struct {
	// Key defines the access API key.
	Key string

	// HTTPClient HTTP client to communicate with the API.
	HTTPClient HTTPClient
}

const (
	baseURL        = "%s"
	defaultTimeout = 2 * time.Minute
)

// Client defines the Neon SDK client.
type Client struct {
	cfg Config

	baseURL string
}

// HTTPClient client to handle http requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func setHeaders(req *http.Request, token string) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
}

func (c Client) requestHandler(url string, t string, reqPayload any, responsePayload any) error {
	var body io.Reader
	var err error

	if reqPayload != nil {
        if v := reflect.ValueOf(reqPayload); v.Kind() == reflect.Struct || !v.IsNil() {
            b, err := json.Marshal(reqPayload)
            if err != nil {
                return err
            }
            body = bytes.NewReader(b)
        }
    }

	req, _ := http.NewRequest(t, url, body)
	setHeaders(req, c.cfg.Key)

	res, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
		return convertErrorResponse(res)
	}

	if responsePayload != nil {
		buf, err := io.ReadAll(res.Body)
	    defer func() { _ = res.Body.Close() }()
		if err != nil {
			return err
		}
		return json.Unmarshal(buf, responsePayload)
	}

	return nil
}

`
)

func Run(openAPISpec []byte, outputDir string) error {
	var spec OpenAPIDefinition
	if err := json.Unmarshal(openAPISpec, &spec); err != nil {
		return fmt.Errorf("could not deserialize API spec: %w", err)
	}

	_, err := os.ReadDir(outputDir)
	switch {
	case os.IsNotExist(err):
		err = os.MkdirAll(outputDir, 0774)
		if err != nil {
			return fmt.Errorf("cannot create output directory %s: %w", outputDir, err)
		}
	case err == nil:
	default:
		return fmt.Errorf("cannot read output directory %s: %w", outputDir, err)
	}

	goModFileOut, err := os.OpenFile(filepath.Join(outputDir, "go.mod"), os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("cannot create go.mod file: %w", err)
	}
	_, err = goModFileOut.WriteString(goModFile)
	if err != nil {
		return fmt.Errorf("cannot write go.mod file: %w", err)
	}
	_ = goModFileOut.Close()

	sdkFileOut, err := os.OpenFile(filepath.Join(outputDir, "sdk.go"), os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("cannot create sdk.go file: %w", err)
	}
	defer func() { _ = sdkFileOut.Close() }()

	sdkFileContent := fmt.Sprintf(sdkFile, spec.ServerURL)
	_, err = sdkFileOut.WriteString(sdkFileContent)
	if err != nil {
		return fmt.Errorf("cannot write sdk.go file: %w", err)
	}

	typeRepo := new(TypesRepo)

	typesDefinitionInputFromComponents(typeRepo, spec.Components)

	methodsDef, err := newGoMethodsDefinition(spec.Paths, spec.Components.Parameters, typeRepo)
	if err != nil {
		return err
	}

	if err := newGoTypesDefinition(typeRepo); err != nil {
		return err
	}

	_, _ = sdkFileOut.WriteString(methodsDef)
	_, _ = sdkFileOut.WriteString(typeRepo.TypesDefinition())

	return nil
}
