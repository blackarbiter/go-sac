package dto

import "time"

// DASTResponse represents the DAST scan result response
type DASTResponse struct {
	TaskID    string `json:"task_id"`
	AssetID   string `json:"asset_id"`
	AssetType string `json:"asset_type"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`

	// DAST specific fields
	URL         string  `json:"url,omitempty"`
	Method      string  `json:"method,omitempty"`
	Parameter   string  `json:"parameter,omitempty"`
	Payload     string  `json:"payload,omitempty"`
	Severity    string  `json:"severity,omitempty"`
	VulnType    string  `json:"vuln_type,omitempty"`
	CVSSScore   float64 `json:"cvss_score,omitempty"`
	Remediation string  `json:"remediation,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DASTQueryRequest represents the DAST query request
type DASTQueryRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

// DASTBatchQueryRequest represents the DAST batch query request
type DASTBatchQueryRequest struct {
	TaskIDs []string `json:"task_ids" binding:"required"`
}
