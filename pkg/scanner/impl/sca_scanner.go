package impl

import (
	"context"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
)

// SCAScanner 软件成分分析扫描器
type SCAScanner struct {
	BaseScanner
}

// NewSCAScanner 创建SCA扫描器
func NewSCAScanner() scanner.Scanner {
	return &SCAScanner{
		BaseScanner: *NewBaseScanner(domain.ScanTypeSca),
	}
}

// Scan 实现扫描接口
func (s *SCAScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	// 创建扫描结果
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeSca, task.AssetID, task.AssetType)

	// 模拟扫描过程
	time.Sleep(100 * time.Millisecond)

	// 设置成功结果
	result.SetSuccess(task.Options)

	return result, nil
}
