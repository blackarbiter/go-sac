package model

// UploadedFileAsset 上传文件资产扩展表
type UploadedFileAsset struct {
	ID         uint   `gorm:"primaryKey"`
	FilePath   string `gorm:"size:1024;not null"`
	FileSize   int64  `gorm:"not null"`
	FileType   string `gorm:"size:100;not null;index"`
	Checksum   string `gorm:"size:128;not null"`
	PreviewURL string `gorm:"size:512"`
}

// TableName 指定表名
func (UploadedFileAsset) TableName() string {
	return "assets_uploaded_file"
}
