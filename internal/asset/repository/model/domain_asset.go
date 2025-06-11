package model

import (
	"time"
)

// DomainAsset 域名资产扩展表
type DomainAsset struct {
	ID            uint      `gorm:"primaryKey"`
	DomainName    string    `gorm:"size:255;not null;index"`
	Registrar     string    `gorm:"size:100"`
	ExpiryDate    time.Time `gorm:"not null;index"`
	DNSServers    string    `gorm:"type:TEXT"`
	SSLExpiryDate time.Time
}

// TableName 指定表名
func (DomainAsset) TableName() string {
	return "assets_domain"
}
