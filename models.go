package sdk

import "time"

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

// ProjectStatus response to project start/stop requests.
type ProjectStatus struct {
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
