package model

import (
	"gorm.io/datatypes"
)

// DesignDocumentAsset 设计文档资产扩展表
type DesignDocumentAsset struct {
	ID              uint           `gorm:"primaryKey"`
	DesignType      string         `gorm:"size:50;not null;index"`
	Components      datatypes.JSON `gorm:"type:jsonb;not null"`
	Diagrams        datatypes.JSON `gorm:"type:jsonb"`
	Dependencies    datatypes.JSON `gorm:"type:jsonb"`
	TechnologyStack []string       `gorm:"type:json"`
}

// TableName 指定表名
func (DesignDocumentAsset) TableName() string {
	return "assets_design_document"
}
