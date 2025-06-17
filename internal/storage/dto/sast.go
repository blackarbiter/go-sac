package dto

import "time"

// SASTResponse represents the SAST scan result response
type SASTResponse struct {
	TaskID    string `json:"task_id"`
	AssetID   string `json:"asset_id"`
	AssetType string `json:"asset_type"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`

	// SAST specific fields
	FilePath      string `json:"file_path,omitempty"`
	LineNumber    int    `json:"line_number,omitempty"`
	Severity      string `json:"severity,omitempty"`
	RuleID        string `json:"rule_id,omitempty"`
	RuleName      string `json:"rule_name,omitempty"`
	Description   string `json:"description,omitempty"`
	CWEID         string `json:"cwe_id,omitempty"`
	FixSuggestion string `json:"fix_suggestion,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SASTQueryRequest represents the SAST query request
type SASTQueryRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

// SASTBatchQueryRequest represents the SAST batch query request
type SASTBatchQueryRequest struct {
	TaskIDs []string `json:"task_ids" binding:"required"`
}
