package model

import (
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// DesignDocumentAsset 设计文档资产扩展表
type DesignDocumentAsset struct {
	ID              uint           `gorm:"primaryKey"`
	DesignType      string         `gorm:"size:50;not null;index"`
	Components      datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	Diagrams        datatypes.JSON `gorm:"type:jsonb"`
	Dependencies    datatypes.JSON `gorm:"type:jsonb"`
	TechnologyStack pq.StringArray `gorm:"type:varchar(100)[]"`
}

// TableName 指定表名
func (DesignDocumentAsset) TableName() string {
	return "assets_design_document"
}
