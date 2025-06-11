package model

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// RequirementAsset 需求文档资产扩展表
type RequirementAsset struct {
	ID                 uint           `gorm:"primaryKey"`
	BusinessValue      string         `gorm:"type:text"`
	Stakeholders       datatypes.JSON `gorm:"type:jsonb;not null"`
	Priority           int            `gorm:"not null;default:0;index"`
	AcceptanceCriteria datatypes.JSON `gorm:"type:jsonb"`
	RelatedDocuments   datatypes.JSON `gorm:"type:jsonb"`
	Version            string         `gorm:"size:20;not null"`
}

// TableName 指定表名
func (RequirementAsset) TableName() string {
	return "assets_requirement"
}

// GetStakeholders 获取利益相关者列表
func (r *RequirementAsset) GetStakeholders() ([]string, error) {
	var stakeholders []string
	err := json.Unmarshal(r.Stakeholders, &stakeholders)
	return stakeholders, err
}

// SetStakeholders 设置利益相关者列表
func (r *RequirementAsset) SetStakeholders(stakeholders []string) error {
	data, err := json.Marshal(stakeholders)
	if err != nil {
		return err
	}
	r.Stakeholders = data
	return nil
}
