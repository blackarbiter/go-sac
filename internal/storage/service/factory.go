package service

import (
	"fmt"
	"github.com/blackarbiter/go-sac/internal/storage/repository"
	"sync"

	"github.com/blackarbiter/go-sac/pkg/domain"
)

// ProcessorFactory 处理器工厂实现
type ProcessorFactory struct {
	processors map[string]StorageProcessor
	mu         sync.RWMutex
}

// NewProcessorFactory 创建处理器工厂
func NewProcessorFactory() *ProcessorFactory {
	return &ProcessorFactory{
		processors: make(map[string]StorageProcessor),
	}
}

// GetProcessor returns the processor for the given scan type
func (f *ProcessorFactory) GetProcessor(scanType domain.ScanType) (StorageProcessor, error) {
	processor, ok := f.processors[scanType.String()]
	if !ok {
		return nil, fmt.Errorf("no processor registered for scan type: %s", scanType)
	}
	return processor, nil
}

// RegisterProcessor registers a processor for a scan type
func (f *ProcessorFactory) RegisterProcessor(scanType domain.ScanType, processor StorageProcessor) {
	f.processors[scanType.String()] = processor
}

// RegisterDefaultProcessors 注册默认处理器
func (f *ProcessorFactory) RegisterDefaultProcessors(repo repository.Repository) {
	// 注册DAST处理器
	f.RegisterProcessor(domain.ScanTypeDast, NewDASTProcessor(repo))

	// 注册SAST处理器
	f.RegisterProcessor(domain.ScanTypeStaticCodeAnalysis, NewSASTProcessor(repo))

	// 注册SCA处理器
	f.RegisterProcessor(domain.ScanTypeSca, NewSCAProcessor(repo))
}
