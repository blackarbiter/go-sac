package scanner_impl

import (
	"context"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"go.uber.org/zap"
)

// DASTScanner 动态应用安全测试扫描器
type DASTScanner struct {
	*BaseScanner
}

// NewDASTScanner 创建DAST扫描器
func NewDASTScanner(
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	opts ...BaseScannerOption,
) scanner.TaskExecutor {
	s := &DASTScanner{}
	baseOpts := []BaseScannerOption{
		WithResourceProfile(scanner.ResourceProfile{
			MinCPU:   2,
			MaxCPU:   4,
			MemoryMB: 2048,
		}),
		WithSecurityProfile(1001, 1001, true),
	}
	baseOpts = append(baseOpts, opts...)

	s.BaseScanner = NewBaseScanner(
		domain.ScanTypeDast,
		timeoutCtrl,
		logger,
		baseOpts...,
	)
	return s
}

// Scan 实现扫描接口
func (d *DASTScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeDast, task.AssetID, task.AssetType)

	d.logger.Info("starting DAST scan",
		zap.String("task_id", task.TaskID),
		zap.String("scan_type", string(task.ScanType)))

	// 使用超时控制执行扫描
	err := d.ExecuteWithTimeout(ctx, task, func(ctx context.Context) error {
		// 2. 模拟扫描过程
		select {
		case <-time.After(10 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	if err != nil {
		result.SetFailed(err.Error())
		return result, err
	}

	result.SetSuccess(task.Options)
	return result, nil
}

// AsyncExecute 实现TaskExecutor接口
func (d *DASTScanner) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	go func() {
		_ = d.BaseScanner.ExecuteWithResult(ctx, task, func(ctx context.Context) (*domain.ScanResult, error) {
			return d.Scan(ctx, task)
		})
	}()
	return task.TaskID, nil
}

// Cancel 实现TaskExecutor接口
func (s *DASTScanner) Cancel(handle string) error {
	return nil
}

// GetStatus 实现TaskExecutor接口
func (s *DASTScanner) GetStatus(handle string) (domain.TaskStatus, error) {
	return domain.TaskStatusCompleted, nil
}

// HealthCheck 实现TaskExecutor接口
func (s *DASTScanner) HealthCheck() error {
	return s.BaseScanner.HealthCheck()
}
