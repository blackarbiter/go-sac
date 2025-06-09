package service

import (
	"fmt"
	"sync"

	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/pkg/domain"
)

// ProcessorFactory 处理器工厂实现
type ProcessorFactory struct {
	processors map[string]AssetProcessor
	mu         sync.RWMutex
}

// NewProcessorFactory 创建处理器工厂
func NewProcessorFactory() *ProcessorFactory {
	return &ProcessorFactory{
		processors: make(map[string]AssetProcessor),
	}
}

// GetProcessor 获取指定类型的处理器
func (f *ProcessorFactory) GetProcessor(assetType string) (AssetProcessor, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	processor, exists := f.processors[assetType]
	if !exists {
		return nil, fmt.Errorf("no processor found for asset type: %s", assetType)
	}
	return processor, nil
}

// RegisterProcessor 注册处理器
func (f *ProcessorFactory) RegisterProcessor(assetType string, processor AssetProcessor) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.processors[assetType] = processor
}

// RegisterDefaultProcessors 注册默认处理器
func (f *ProcessorFactory) RegisterDefaultProcessors(repo repository.Repository) {
	// 注册需求文档处理器
	f.RegisterProcessor(domain.AssetTypeRequirement.String(), NewRequirementProcessor(repo))

	// 注册设计文档处理器
	f.RegisterProcessor(domain.AssetTypeDesignDocument.String(), NewDesignDocumentProcessor(repo))

	// 注册代码仓库处理器
	f.RegisterProcessor(domain.AssetTypeRepository.String(), NewRepositoryProcessor(repo))

	// 注册上传文件处理器
	f.RegisterProcessor(domain.AssetTypeUploadedFile.String(), NewUploadedFileProcessor(repo))

	// 注册容器镜像处理器
	f.RegisterProcessor(domain.AssetTypeImage.String(), NewImageProcessor(repo))

	// 注册域名处理器
	f.RegisterProcessor(domain.AssetTypeDomain.String(), NewDomainProcessor(repo))

	// 注册IP地址处理器
	f.RegisterProcessor(domain.AssetTypeIP.String(), NewIPProcessor(repo))
}
