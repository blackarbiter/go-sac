package impl

import (
	"context"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
)

// DASTScanner 动态应用安全测试扫描器
type DASTScanner struct {
	BaseScanner
}

// NewDASTScanner 创建DAST扫描器
func NewDASTScanner() scanner.Scanner {
	return &DASTScanner{
		BaseScanner: *NewBaseScanner(domain.ScanTypeDast),
	}
}

// Scan 实现扫描接口
func (s *DASTScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	// 创建扫描结果
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeDast, task.AssetID, task.AssetType)

	// 模拟扫描过程
	time.Sleep(2000 * time.Millisecond)

	// 设置成功结果
	result.SetSuccess(task.Options)

	return result, nil
}
