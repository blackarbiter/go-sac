package impl

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
)

// BaseScanner 基础扫描器实现
type BaseScanner struct {
	scanType domain.ScanType
}

// NewBaseScanner 创建基础扫描器
func NewBaseScanner(scanType domain.ScanType) *BaseScanner {
	return &BaseScanner{
		scanType: scanType,
	}
}
