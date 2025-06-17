package model

import "time"

// SASTModel represents the SAST scan result
type SASTModel struct {
	ID        uint   `gorm:"primaryKey"`
	TaskID    string `gorm:"type:varchar(64);not null;index"`
	AssetID   string `gorm:"type:varchar(64);not null;index"`
	AssetType string `gorm:"type:varchar(32);not null"`
	Status    string `gorm:"type:varchar(16);not null"`
	Error     string `gorm:"type:text"`

	// SAST specific fields
	FilePath      string `gorm:"type:varchar(512)"`
	LineNumber    int    `gorm:"type:int"`
	Severity      string `gorm:"type:varchar(16)"`
	RuleID        string `gorm:"type:varchar(64)"`
	RuleName      string `gorm:"type:varchar(128)"`
	Description   string `gorm:"type:text"`
	CWEID         string `gorm:"type:varchar(16)"`
	FixSuggestion string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for SASTModel
func (SASTModel) TableName() string {
	return "sast_results"
}
