package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseAsset 资产基表模型
type BaseAsset struct {
	ID             uint      `gorm:"primaryKey"`
	AssetType      string    `gorm:"size:50;not null;index"`
	Name           string    `gorm:"size:255;not null"`
	Status         string    `gorm:"size:50;not null;default:'active'"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
	CreatedBy      string    `gorm:"size:100;not null"`
	UpdatedBy      string    `gorm:"size:100;not null"`
	ProjectID      uint      `gorm:"index"`
	OrganizationID uint      `gorm:"not null;index"`
	Tags           string    `gorm:"type:TEXT"`
}

// TableName 指定表名
func (BaseAsset) TableName() string {
	return "assets_base"
}

// BeforeCreate 创建前的钩子
func (b *BaseAsset) BeforeCreate(tx *gorm.DB) error {
	if b.Status == "" {
		b.Status = "active"
	}
	return nil
}
