package impl

import (
	"context"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
)

// SASTScanner 静态代码分析扫描器
type SASTScanner struct {
	BaseScanner
}

// NewSASTScanner 创建SAST扫描器
func NewSASTScanner() scanner.Scanner {
	return &SASTScanner{
		BaseScanner: *NewBaseScanner(domain.ScanTypeStaticCodeAnalysis),
	}
}

// Scan 实现扫描接口
func (s *SASTScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	// 创建扫描结果
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeStaticCodeAnalysis, task.AssetID, task.AssetType)

	// 模拟扫描过程
	time.Sleep(2000 * time.Millisecond)

	// 设置成功结果
	result.SetSuccess(task.Options)

	return result, nil
}
