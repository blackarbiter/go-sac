package model

import "time"

// DASTModel represents the DAST scan result
type DASTModel struct {
	ID        uint   `gorm:"primaryKey"`
	TaskID    string `gorm:"type:varchar(64);not null;index"`
	AssetID   string `gorm:"type:varchar(64);not null;index"`
	AssetType string `gorm:"type:varchar(32);not null"`
	Status    string `gorm:"type:varchar(16);not null"`
	Error     string `gorm:"type:text"`

	// DAST specific fields
	URL         string  `gorm:"type:varchar(1024)"`
	Method      string  `gorm:"type:varchar(16)"`
	Parameter   string  `gorm:"type:varchar(512)"`
	Payload     string  `gorm:"type:text"`
	Severity    string  `gorm:"type:varchar(16)"`
	VulnType    string  `gorm:"type:varchar(64)"`
	CVSSScore   float64 `gorm:"type:decimal(3,1)"`
	Remediation string  `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for DASTModel
func (DASTModel) TableName() string {
	return "dast_results"
}
