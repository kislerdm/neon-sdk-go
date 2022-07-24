package sdk

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Error API error.
type Error struct {
	Code, Message string
	HTTPCode      int
}

func (e Error) Error() string {
	return fmt.Sprintf("[HTTP Code: %d][Error Code: %s] %s", e.HTTPCode, e.Code, e.Message)
}

// Database Neon Database metadata.
type Database struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	OwnerId int    `json:"owner_id"`
}

// Role Neon Role metadata.
type Role struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// AdditionalOptions additional options.
type AdditionalOptions map[string]string

// ProjectInfo metadata of a Project.
type ProjectInfo struct {
	CreatedAt      time.Time         `json:"created_at"`
	CurrentState   string            `json:"current_state"`
	Databases      []Database        `json:"databases"`
	Deleted        bool              `json:"deleted"`
	ID             string            `json:"id"`
	InstanceHandle string            `json:"instance_handle"`
	InstanceTypeID string            `json:"instance_type_id"`
	MaxProjectSize int               `json:"max_project_size"`
	Name           string            `json:"name"`
	ParentID       string            `json:"parent_id"`
	PendingState   string            `json:"pending_state"`
	PlatformID     string            `json:"platform_id"`
	PlatformName   string            `json:"platform_name"`
	PoolerEnabled  bool              `json:"pooler_enabled"`
	RegionID       string            `json:"region_id"`
	RegionName     string            `json:"region_name"`
	Roles          []Role            `json:"roles"`
	Settings       AdditionalOptions `json:"settings,omitempty"`
	Size           int               `json:"size"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ProjectSettingsRequestCreate settings to create a new Project.
type ProjectSettingsRequestCreate struct {
	Name           string            `json:"name"`
	PlatformID     string            `json:"platform_id"`
	InstanceHandle string            `json:"instance_handle"`
	RegionID       string            `json:"region_id"`
	Settings       AdditionalOptions `json:"settings"`
}

// ProjectSettingsRequestUpdate settings to update existing Project.
type ProjectSettingsRequestUpdate struct {
	InstanceTypeID string            `json:"instance_type_id"`
	Name           string            `json:"name"`
	PoolerEnabled  bool              `json:"pooler_enabled"`
	Settings       AdditionalOptions `json:"settings"`
}

// ProjectDeleteResponse response to the Project deletion request.
type ProjectDeleteResponse struct {
	Action        string    `json:"action"`
	CreatedAt     time.Time `json:"created_at"`
	Error         string    `json:"error"`
	FailuresCount int       `json:"failures_count"`
	ID            int       `json:"id"`
	ProjectID     string    `json:"project_id"`
	RetryAt       time.Time `json:"retry_at"`
	Status        string    `json:"status"`
	UpdatedAt     time.Time `json:"updated_at"`
	UUID          string    `json:"uuid"`
}

// Client to interact with Neon API.
type Client interface {
	// ValidateAPIKey makes a call to validate API access key.
	ValidateAPIKey() error

	// ListProjects returns existing Projects.
	ListProjects() ([]ProjectInfo, error)

	// ReadInfoProject returns Project info.
	ReadInfoProject(projectID string) (ProjectInfo, error)

	// CreateProject creates new Project.
	CreateProject(settings ProjectSettingsRequestCreate) (ProjectInfo, error)

	// UpdateProject updates existing Project.
	UpdateProject(settings ProjectSettingsRequestUpdate) (ProjectInfo, error)

	// DeleteProject deletes existing Project.
	DeleteProject(projectID string) (ProjectDeleteResponse, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Options struct {
	// APIKey API access key.
	APIKey string

	// HTTPClient client to communicate with the API over http.
	HTTPClient httpClient
}

type client struct {
	options Options
}

func (c client) ValidateAPIKey() error {
	//TODO implement me
	panic("implement me")
}

func (c client) ListProjects() ([]ProjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) ReadInfoProject(projectID string) (ProjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) CreateProject(settings ProjectSettingsRequestCreate) (ProjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) UpdateProject(settings ProjectSettingsRequestUpdate) (ProjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) DeleteProject(projectID string) (ProjectDeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

// NewClient initialised new client to communicate to Neon over http API.
// See details: https://neon.tech/docs/reference/about/
func NewClient(ctx context.Context, optFns ...func(*Options) error) (Client, error) {
	o := Options{}
	for _, fn := range optFns {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	resolveAPIKey(&o)
	resolveHTTPClient(&o)

	c := client{options: o}

	if err := c.ValidateAPIKey(); err != nil {
		return nil, fmt.Errorf(
			"invalid API access key. find details: https://neon.tech/docs/get-started-with-neon/using-api-keys/",
		)
	}

	return &c, nil
}

func resolveHTTPClient(o *Options) {
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: 5 * time.Minute}
	}
}

func resolveAPIKey(o *Options) {
	if o.APIKey == "" {
		o.APIKey = os.Getenv("NEON_API_KEY")
	}
}

// WithAPIKey sets the API access key.
func WithAPIKey(apiKey string) func(*Options) error {
	return func(o *Options) error {
		o.APIKey = apiKey
		return nil
	}
}
