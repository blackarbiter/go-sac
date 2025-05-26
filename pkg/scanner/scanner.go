package scanner

import (
	"context"

	"github.com/blackarbiter/go-sac/pkg/domain"
)

// Scanner 定义扫描器接口
type Scanner interface {
	// Scan 执行扫描任务
	Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error)
}

// ScannerFactory 定义扫描器工厂接口
type ScannerFactory interface {
	// GetScanner 根据扫描类型获取对应的扫描器
	GetScanner(scanType domain.ScanType) (Scanner, error)
}
