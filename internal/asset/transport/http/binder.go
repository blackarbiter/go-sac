package http

import (
	"encoding/json"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/pkg/domain"
)

// AssetBinder 处理资产请求的动态绑定
type AssetBinder struct{}

// NewAssetBinder 创建资产绑定器实例
func NewAssetBinder() *AssetBinder {
	return &AssetBinder{}
}

// Bind 根据资产类型绑定请求数据
func (b *AssetBinder) Bind(assetType string, body []byte) (interface{}, error) {
	// 解析资产类型
	at, err := domain.ParseAssetType(assetType)
	if err != nil {
		return nil, fmt.Errorf("invalid asset type: %w", err)
	}

	// 根据资产类型绑定请求
	switch at {
	case domain.AssetTypeRequirement:
		var req dto.CreateRequirementRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal requirement request: %w", err)
		}
		return &req, nil

	case domain.AssetTypeDesignDocument:
		var req dto.CreateDesignDocumentRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal design document request: %w", err)
		}
		return &req, nil

	case domain.AssetTypeRepository:
		var req dto.CreateRepositoryRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal repository request: %w", err)
		}
		return &req, nil

	case domain.AssetTypeUploadedFile:
		var req dto.CreateUploadedFileRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal uploaded file request: %w", err)
		}
		return &req, nil

	case domain.AssetTypeImage:
		var req dto.CreateImageRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal image request: %w", err)
		}
		return &req, nil

	case domain.AssetTypeDomain:
		var req dto.CreateDomainRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal domain request: %w", err)
		}
		return &req, nil

	case domain.AssetTypeIP:
		var req dto.CreateIPRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal IP request: %w", err)
		}
		return &req, nil

	default:
		return nil, fmt.Errorf("unsupported asset type: %s", assetType)
	}
}
