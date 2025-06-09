package model

import (
	"time"

	"gorm.io/datatypes"
)

// RepositoryAsset 代码仓库资产扩展表
type RepositoryAsset struct {
	ID             uint   `gorm:"primaryKey"`
	RepoURL        string `gorm:"size:512;not null;index"`
	Branch         string `gorm:"size:100;not null;default:'main'"`
	LastCommitHash string `gorm:"size:100"`
	LastCommitTime time.Time
	Language       string         `gorm:"size:50;not null;index"`
	CICDConfig     datatypes.JSON `gorm:"type:jsonb"`
}

// TableName 指定表名
func (RepositoryAsset) TableName() string {
	return "assets_repository"
}
