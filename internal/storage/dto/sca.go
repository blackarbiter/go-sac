package dto

import "time"

// SCAResponse represents the SCA scan result response
type SCAResponse struct {
	TaskID    string `json:"task_id"`
	AssetID   string `json:"asset_id"`
	AssetType string `json:"asset_type"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`

	// SCA specific fields
	PackageName      string `json:"package_name,omitempty"`
	PackageVersion   string `json:"package_version,omitempty"`
	License          string `json:"license,omitempty"`
	Vulnerabilities  string `json:"vulnerabilities,omitempty"`
	DirectDependency bool   `json:"direct_dependency,omitempty"`
	DependencyPath   string `json:"dependency_path,omitempty"`
	LatestVersion    string `json:"latest_version,omitempty"`
	UpdateAvailable  bool   `json:"update_available,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SCAQueryRequest represents the SCA query request
type SCAQueryRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

// SCABatchQueryRequest represents the SCA batch query request
type SCABatchQueryRequest struct {
	TaskIDs []string `json:"task_ids" binding:"required"`
}
