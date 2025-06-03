package scanner_impl

import (
	"context"
	"os/exec"
	"time"

	"github.com/blackarbiter/go-sac/pkg/config"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"go.uber.org/zap"
)

// SCAScanner 软件成分分析扫描器
type SCAScanner struct {
	*BaseScanner
}

// NewSCAScanner 创建SCA扫描器
func NewSCAScanner(
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	config *config.Config,
	opts ...BaseScannerOption,
) scanner.TaskExecutor {
	s := &SCAScanner{}

	// 从配置获取参数
	resourceProfile, securityConfig, timeout := config.GetScannerConfig(domain.ScanTypeSca)
	baseOpts := []BaseScannerOption{
		WithResourceProfile(resourceProfile),
		WithSecurityProfile(int(securityConfig.RunAsUser), int(securityConfig.RunAsGroup), !securityConfig.AllowPrivilegeEscalation),
		WithTimeout(timeout, 30*time.Second),
		WithCircuitBreaker(config),
	}
	baseOpts = append(baseOpts, opts...)

	s.BaseScanner = NewBaseScanner(
		domain.ScanTypeSca,
		timeoutCtrl,
		logger,
		config,
		baseOpts...,
	)
	return s
}

// Scan 实现扫描接口
func (s *SCAScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	// 创建扫描结果
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeSca, task.AssetID, task.AssetType)

	// 执行目录扫描命令
	cmd := exec.CommandContext(ctx, "sleep", "5")
	if err := s.ExecuteCommand(ctx, task, cmd, ""); err != nil {
		result.SetFailed(err.Error())
		return result, err
	}
	// 设置成功结果
	result.SetSuccess(task.Options)
	return result, nil
}

// AsyncExecute 实现TaskExecutor接口
func (s *SCAScanner) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	return s.BaseScanner.AsyncExecuteWithResult(ctx, task, func(ctx context.Context) (*domain.ScanResult, error) {
		return s.Scan(ctx, task)
	})
}

// SyncExecute 实现 TaskExecutor 接口
func (s *SCAScanner) SyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	return s.BaseScanner.ExecuteWithResult(ctx, task, func(ctx context.Context) (*domain.ScanResult, error) {
		return s.Scan(ctx, task)
	})
}

// Cancel 实现TaskExecutor接口
func (s *SCAScanner) Cancel(handle string) error {
	return nil
}

// GetStatus 实现TaskExecutor接口
func (s *SCAScanner) GetStatus(handle string) (domain.TaskStatus, error) {
	return domain.TaskStatusCompleted, nil
}

// HealthCheck 实现TaskExecutor接口
func (s *SCAScanner) HealthCheck() error {
	return s.BaseScanner.HealthCheck()
}
