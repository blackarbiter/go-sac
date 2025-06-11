package domain

import (
	"fmt"
	"strings"
)

// AssetType 定义资产分类标准
type AssetType uint8

const (
	AssetTypeUnknown        AssetType = iota // 默认未知类型
	AssetTypeRequirement                     // 需求文档
	AssetTypeDesignDocument                  // 设计文档
	AssetTypeRepository                      // 代码仓库
	AssetTypeUploadedFile                    // 上传文件
	AssetTypeImage                           // 容器镜像
	AssetTypeDomain                          // 域名
	AssetTypeIP                              // IP地址
)

// String 返回可读类型名称
func (t AssetType) String() string {
	switch t {
	case AssetTypeRequirement:
		return "Requirement"
	case AssetTypeDesignDocument:
		return "DesignDocument"
	case AssetTypeRepository:
		return "Repository"
	case AssetTypeUploadedFile:
		return "UploadedFile"
	case AssetTypeImage:
		return "Image"
	case AssetTypeDomain:
		return "Domain"
	case AssetTypeIP:
		return "IP"
	default:
		return "Unknown"
	}
}

func ParseAssetType(s string) (AssetType, error) {
	// 统一转换为小写进行匹配
	lowerInput := strings.ToLower(s)

	switch lowerInput {
	case "requirement":
		return AssetTypeRequirement, nil
	case "designdocument":
		return AssetTypeDesignDocument, nil
	case "repository":
		return AssetTypeRepository, nil
	case "uploadedfile":
		return AssetTypeUploadedFile, nil
	case "image":
		return AssetTypeImage, nil
	case "domain":
		return AssetTypeDomain, nil
	case "ip":
		return AssetTypeIP, nil
	default:
		return AssetTypeUnknown, fmt.Errorf("unknown asset type: %s", s)
	}
}
