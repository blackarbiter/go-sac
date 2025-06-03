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

// SASTScanner 静态代码分析扫描器
type SASTScanner struct {
	*BaseScanner
}

// NewSASTScanner 创建SAST扫描器
func NewSASTScanner(
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	config *config.Config,
	opts ...BaseScannerOption,
) scanner.TaskExecutor {
	s := &SASTScanner{}

	// 从配置获取参数
	resourceProfile, securityConfig, timeout := config.GetScannerConfig(domain.ScanTypeStaticCodeAnalysis)

	baseOpts := []BaseScannerOption{
		WithResourceProfile(resourceProfile),
		WithSecurityProfile(int(securityConfig.RunAsUser), int(securityConfig.RunAsGroup), !securityConfig.AllowPrivilegeEscalation),
		WithTimeout(timeout, 30*time.Second),
		WithCircuitBreaker(config),
	}
	baseOpts = append(baseOpts, opts...)

	s.BaseScanner = NewBaseScanner(
		domain.ScanTypeStaticCodeAnalysis,
		timeoutCtrl,
		logger,
		config,
		baseOpts...,
	)
	return s
}

// Scan 实现扫描接口
func (s *SASTScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	// 创建扫描结果
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeStaticCodeAnalysis, task.AssetID, task.AssetType)

	// 执行代码扫描命令
	s.logger.Info("SAST execute", zap.Duration("within time ", s.defaultTimeout))
	cmd := exec.CommandContext(ctx, "sleep", "3")
	if err := s.ExecuteCommand(ctx, task, cmd, ""); err != nil {
		result.SetFailed(err.Error())
		return result, err
	}
	// 设置成功结果
	result.SetSuccess(task.Options)
	return result, nil
}

// AsyncExecute 实现TaskExecutor接口
func (s *SASTScanner) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	return s.BaseScanner.AsyncExecuteWithResult(ctx, task, func(ctx context.Context) (*domain.ScanResult, error) {
		return s.Scan(ctx, task)
	})
}

// SyncExecute 实现 TaskExecutor 接口
func (s *SASTScanner) SyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	return s.BaseScanner.ExecuteWithResult(ctx, task, func(ctx context.Context) (*domain.ScanResult, error) {
		return s.Scan(ctx, task)
	})
}

// Cancel 实现TaskExecutor接口
func (s *SASTScanner) Cancel(handle string) error {
	return nil
}

// GetStatus 实现TaskExecutor接口
func (s *SASTScanner) GetStatus(handle string) (domain.TaskStatus, error) {
	return domain.TaskStatusCompleted, nil
}

// HealthCheck 实现TaskExecutor接口
func (s *SASTScanner) HealthCheck() error {
	return s.BaseScanner.HealthCheck()
}
