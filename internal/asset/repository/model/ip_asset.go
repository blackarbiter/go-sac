package model

// IPAsset IP地址资产扩展表
type IPAsset struct {
	ID          uint   `gorm:"primaryKey"`
	IPAddress   string `gorm:"type:inet;not null;index"`
	SubnetMask  string `gorm:"size:100"`
	Gateway     string `gorm:"size:100"`
	DHCPEnabled bool   `gorm:"default:false"`
	DeviceType  string `gorm:"size:50;index"`
	MACAddress  string `gorm:"size:20"`
}

// TableName 指定表名
func (IPAsset) TableName() string {
	return "assets_ip"
}
