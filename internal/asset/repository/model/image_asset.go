package model

import (
	"gorm.io/datatypes"
)

// ImageAsset 容器镜像资产扩展表
type ImageAsset struct {
	ID              uint           `gorm:"primaryKey"`
	RegistryURL     string         `gorm:"size:512;not null;index"`
	ImageName       string         `gorm:"size:255;not null;index"`
	Tag             string         `gorm:"size:100;not null;default:'latest'"`
	Digest          string         `gorm:"size:200;not null"`
	Size            int64          `gorm:"not null"`
	Vulnerabilities datatypes.JSON `gorm:"type:jsonb"`
}

// TableName 指定表名
func (ImageAsset) TableName() string {
	return "assets_image"
}
