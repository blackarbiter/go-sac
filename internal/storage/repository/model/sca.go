package model

import "time"

// SCAModel represents the SCA scan result
type SCAModel struct {
	ID        uint   `gorm:"primaryKey"`
	TaskID    string `gorm:"type:varchar(64);not null;index"`
	AssetID   string `gorm:"type:varchar(64);not null;index"`
	AssetType string `gorm:"type:varchar(32);not null"`
	Status    string `gorm:"type:varchar(16);not null"`
	Error     string `gorm:"type:text"`

	// SCA specific fields
	PackageName      string `gorm:"type:varchar(256)"`
	PackageVersion   string `gorm:"type:varchar(64)"`
	License          string `gorm:"type:varchar(128)"`
	Vulnerabilities  string `gorm:"type:json"`
	DirectDependency bool   `gorm:"type:boolean"`
	DependencyPath   string `gorm:"type:text"`
	LatestVersion    string `gorm:"type:varchar(64)"`
	UpdateAvailable  bool   `gorm:"type:boolean"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for SCAModel
func (SCAModel) TableName() string {
	return "sca_results"
}
