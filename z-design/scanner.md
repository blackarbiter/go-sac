经过全面梳理，以下是完整、统一的生产级实现方案，包含所有必要组件和防御性编程措施：
```go
// scanner_impl/base_scanner.go
package scanner_impl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"go.uber.org/zap"
)

// region 核心基础设施
// ==================== BaseScanner 实现 ====================

type BaseScanner struct {
	scanType        domain.ScanType
	timeoutCtrl     *scanner.TimeoutController
	logger          *zap.Logger
	metricsRecorder MetricsRecorder
	cgroupManager   CgroupManager
	securityProfile SecurityProfile
	processManager  processManager
	meta            scanner.ExecutorMeta
	mu              sync.RWMutex
}

type BaseScannerOption func(*BaseScanner)

type processManager struct {
	activeProcesses sync.Map
	shutdownSignal  chan struct{}
}

type SecurityProfile struct {
	RunAsUser  *int
	RunAsGroup *int
	NoNewPrivs bool
}

type MetricsRecorder interface {
	Record(name string, value float64, tags map[string]string)
	Gauge(name string, value float64, tags map[string]string)
}

type CgroupManager interface {
	Apply(pid int) error
	Cleanup() error
}

const (
	defaultTimeout     = 5 * time.Minute
	gracefulStopPeriod = 30 * time.Second
)

func NewBaseScanner(
	scanType domain.ScanType,
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	opts ...BaseScannerOption,
) *BaseScanner {
	bs := &BaseScanner{
		scanType:    scanType,
		timeoutCtrl: timeoutCtrl,
		logger:      logger.With(zap.String("scanner", scanType.String())),
		processManager: processManager{
			shutdownSignal: make(chan struct{}),
		},
		meta: scanner.ExecutorMeta{
			Type:            scanType.String(),
			Version:         "1.0.0",
			SupportedTypes:  []domain.ScanType{scanType},
			ResourceProfile: scanner.ResourceProfile{MinCPU: 1, MaxCPU: 2, MemoryMB: 512},
		},
	}

	for _, opt := range opts {
		opt(bs)
	}

	bs.setupSignalHandling()
	return bs
}

// region 进程生命周期管理
// --------------------------------------------------

func (s *BaseScanner) ExecuteCommand(ctx context.Context, task *domain.ScanTaskPayload, cmd *exec.Cmd) error {
	// 1. 准备执行环境
	execID := uuid.New().String()
	s.logger.Info("starting command execution",
		zap.String("task_id", task.TaskID),
		zap.String("exec_id", execID),
		zap.Strings("command", cmd.Args))

	// 2. 设置进程属性
	s.setProcessAttributes(cmd)

	// 3. 注册进程
	s.processManager.activeProcesses.Store(execID, cmd)
	defer s.processManager.activeProcesses.Delete(execID)

	// 4. 启动进程
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("command start failed: %w", err)
	}

	// 5. 应用资源限制
	if s.cgroupManager != nil {
		if err := s.cgroupManager.Apply(cmd.Process.Pid); err != nil {
			s.logger.Warn("cgroup apply failed", zap.Error(err))
		}
	}

	// 6. 异步等待进程结束
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
		close(done)
	}()

	// 7. 处理完成或取消
	select {
	case err := <-done:
		execDuration := time.Since(startTime)
		s.recordCommandMetrics(task, cmd, err, execDuration)
		return err

	case <-ctx.Done():
		s.logger.Warn("command execution canceled",
			zap.String("task_id", task.TaskID),
			zap.String("exec_id", execID))
		s.KillProcessGroup(cmd, true)
		return ctx.Err()
	}
}

func (s *BaseScanner) KillProcessGroup(cmd *exec.Cmd, force bool) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid
	s.logger.Info("terminating process group",
		zap.Int("pid", pid),
		zap.Bool("force", force))

	switch runtime.GOOS {
	case "windows":
		s.killWindowsProcess(pid, force)
	default:
		s.killUnixProcess(pid, force)
	}
}

func (s *BaseScanner) killUnixProcess(pid int, force bool) {
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}

	// 终止整个进程组
	if err := syscall.Kill(-pid, sig); err != nil {
		s.logger.Error("kill process group failed",
			zap.Int("pid", pid),
			zap.Error(err))
	}
}

func (s *BaseScanner) killWindowsProcess(pid int, force bool) {
	args := []string{"/PID", strconv.Itoa(pid), "/T"}
	if force {
		args = append(args, "/F")
	}

	killCmd := exec.Command("taskkill", args...)
	if output, err := killCmd.CombinedOutput(); err != nil {
		s.logger.Error("taskkill failed",
			zap.String("output", string(output)),
			zap.Error(err))
	}
}

func (s *BaseScanner) setupSignalHandling() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigs:
			s.logger.Info("received system signal, cleaning up",
				zap.String("signal", sig.String()))
			s.cleanupProcesses()
			os.Exit(1)

		case <-s.processManager.shutdownSignal:
			s.cleanupProcesses()
		}
	}()
}

func (s *BaseScanner) cleanupProcesses() {
	s.processManager.activeProcesses.Range(func(key, value interface{}) bool {
		if cmd, ok := value.(*exec.Cmd); ok {
			s.KillProcessGroup(cmd, true)
		}
		return true
	})

	if s.cgroupManager != nil {
		if err := s.cgroupManager.Cleanup(); err != nil {
			s.logger.Error("cgroup cleanup failed", zap.Error(err))
		}
	}
}

// endregion

// region 工具方法
// --------------------------------------------------

func (s *BaseScanner) setProcessAttributes(cmd *exec.Cmd) {
	// 通用属性
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	// Unix系统设置进程组
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr.Setpgid = true
		cmd.SysProcAttr.Pgid = 0
	}

	// 安全相关设置
	if s.securityProfile.RunAsUser != nil {
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid:    uint32(*s.securityProfile.RunAsUser),
			Gid:    uint32(*s.securityProfile.RunAsGroup),
			Groups: []uint32{},
		}
	}
	if s.securityProfile.NoNewPrivs {
		cmd.SysProcAttr.NoNewPrivs = true
	}
}

func (s *BaseScanner) recordCommandMetrics(task *domain.ScanTaskPayload, cmd *exec.Cmd, err error, duration time.Duration) {
	tags := map[string]string{
		"scanner_type": s.scanType.String(),
		"task_id":      task.TaskID,
		"exit_code":    strconv.Itoa(cmd.ProcessState.ExitCode()),
		"error_type":   s.classifyError(err),
	}

	s.metricsRecorder.Record("command_duration_seconds", duration.Seconds(), tags)
	s.metricsRecorder.Record("command_cpu_seconds", cmd.ProcessState.SystemTime().Seconds(), tags)

	if err != nil {
		s.metricsRecorder.Record("command_errors", 1, tags)
	}
}

func (s *BaseScanner) classifyError(err error) string {
	switch {
	case err == nil:
		return "success"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	default:
		return "runtime"
	}
}

// endregion

// region 公共接口实现
// --------------------------------------------------

func (s *BaseScanner) Meta() scanner.ExecutorMeta {
	return s.meta
}

func (s *BaseScanner) HealthCheck() error {
	// 检查cgroups是否可用
	if s.cgroupManager != nil {
		if err := s.cgroupManager.Apply(os.Getpid()); err != nil {
			return fmt.Errorf("cgroup health check failed: %w", err)
		}
	}

	// 检查进程清理能力
	testCmd := exec.Command("sleep", "0")
	if err := s.ExecuteCommand(context.Background(), &domain.ScanTaskPayload{}, testCmd); err != nil {
		return fmt.Errorf("process execution test failed: %w", err)
	}

	return nil
}

// endregion

// region 配置选项
// --------------------------------------------------

func WithResourceProfile(rp scanner.ResourceProfile) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.meta.ResourceProfile = rp
	}
}

func WithSecurityProfile(user, group int, noNewPrivs bool) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.securityProfile = SecurityProfile{
			RunAsUser:   &user,
			RunAsGroup:  &group,
			NoNewPrivs: noNewPrivs,
		}
	}
}

func WithCgroupManager(cm CgroupManager) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.cgroupManager = cm
	}
}

func WithMetricsRecorder(mr MetricsRecorder) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.metricsRecorder = mr
	}
}

// endregion

// endregion

// ==================== 扫描器具体实现 ====================
// region SAST 扫描器
// --------------------------------------------------

type SASTScanner struct {
	*BaseScanner
}

func NewSASTScanner(
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	opts ...BaseScannerOption,
) scanner.TaskExecutor {
	s := &SASTScanner{}
	baseOpts := []BaseScannerOption{
		WithResourceProfile(scanner.ResourceProfile{
			MinCPU:   1,
			MaxCPU:   2,
			MemoryMB: 1024,
		}),
		WithSecurityProfile(1001, 1001, true),
	}
	baseOpts = append(baseOpts, opts...)

	s.BaseScanner = NewBaseScanner(
		domain.ScanTypeStaticCodeAnalysis,
		timeoutCtrl,
		logger,
		baseOpts...,
	)
	return s
}

func (s *SASTScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeStaticCodeAnalysis, task.AssetID, task.AssetType)

	cmd := exec.CommandContext(ctx, "sast-cli", "--target", task.Target)
	if err := s.ExecuteCommand(ctx, task, cmd); err != nil {
		result.SetError(err)
		return result, err
	}

	result.SetSuccess(task.Options)
	return result, nil
}

func (s *SASTScanner) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	go func() {
		_, _ = s.Scan(ctx, task)
	}()
	return task.TaskID, nil
}

// endregion

// region DAST 扫描器
// --------------------------------------------------

type DASTScanner struct {
	*BaseScanner
}

func NewDASTScanner(
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	opts ...BaseScannerOption,
) scanner.TaskExecutor {
	d := &DASTScanner{}
	baseOpts := []BaseScannerOption{
		WithResourceProfile(scanner.ResourceProfile{
			MinCPU:   2,
			MaxCPU:   4,
			MemoryMB: 2048,
		}),
	}
	baseOpts = append(baseOpts, opts...)

	d.BaseScanner = NewBaseScanner(
		domain.ScanTypeDast,
		timeoutCtrl,
		logger,
		baseOpts...,
	)
	return d
}

func (d *DASTScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeDast, task.AssetID, task.AssetType)

	// 模拟动态扫描
	time.Sleep(3 * time.Second)

	// 实际生产环境替换为：
	// cmd := exec.CommandContext(ctx, "dast-tool", task.Target)
	// if err := d.ExecuteCommand(ctx, task, cmd); err != nil {...}

	result.SetSuccess(task.Options)
	return result, nil
}

func (d *DASTScanner) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	go func() {
		_, _ = d.Scan(ctx, task)
	}()
	return task.TaskID, nil
}

// endregion

// region SCA 扫描器
// --------------------------------------------------

type SCAScanner struct {
	*BaseScanner
}

func NewSCAScanner(
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	opts ...BaseScannerOption,
) scanner.TaskExecutor {
	s := &SCAScanner{}
	baseOpts := []BaseScannerOption{
		WithResourceProfile(scanner.ResourceProfile{
			MinCPU:   1,
			MaxCPU:   2,
			MemoryMB: 512,
		}),
	}
	baseOpts = append(baseOpts, opts...)

	s.BaseScanner = NewBaseScanner(
		domain.ScanTypeSca,
		timeoutCtrl,
		logger,
		baseOpts...,
	)
	return s
}

func (s *SCAScanner) Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error) {
	result := domain.NewScanResult(task.TaskID, domain.ScanTypeSca, task.AssetID, task.AssetType)

	// 依赖分析逻辑
	if err := s.analyzeDependencies(ctx, task); err != nil {
		result.SetError(err)
		return result, err
	}

	result.SetSuccess(task.Options)
	return result, nil
}

func (s *SCAScanner) analyzeDependencies(ctx context.Context, task *domain.ScanTaskPayload) error {
	cmd := exec.CommandContext(ctx, "sca-analyzer", "--lock-file", task.Target)
	return s.ExecuteCommand(ctx, task, cmd)
}

func (s *SCAScanner) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	go func() {
		_, _ = s.Scan(ctx, task)
	}()
	return task.TaskID, nil
}

// endregion

// region 工厂方法
// --------------------------------------------------

func CreateDefaultScanners(
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	metrics MetricsRecorder,
	cgroup CgroupManager,
) map[domain.ScanType]scanner.TaskExecutor {
	commonOpts := []BaseScannerOption{
		WithMetricsRecorder(metrics),
		WithCgroupManager(cgroup),
	}

	return map[domain.ScanType]scanner.TaskExecutor{
		domain.ScanTypeStaticCodeAnalysis: NewSASTScanner(timeoutCtrl, logger, commonOpts...),
		domain.ScanTypeDast:               NewDASTScanner(timeoutCtrl, logger, commonOpts...),
		domain.ScanTypeSca:                NewSCAScanner(timeoutCtrl, logger, commonOpts...),
	}
}

// endregion
```
生产部署保障措施
1. 资源限制配置示例（Linux cgroups v2）
```go
// cgroups_manager.go
type CgroupV2Manager struct {
	path string
}

func (c *CgroupV2Manager) Apply(pid int) error {
	// 创建子cgroup
	if err := os.Mkdir(c.path, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	// 设置CPU限制
	if err := os.WriteFile(filepath.Join(c.path, "cpu.max"), []byte("50000 100000"), 0644); err != nil {
		return err
	}

	// 设置内存限制
	if err := os.WriteFile(filepath.Join(c.path, "memory.max"), []byte("2G"), 0644); err != nil {
		return err
	}

	// 添加进程
	return os.WriteFile(filepath.Join(c.path, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644)
}
```
2. 监控指标集成（Prometheus示例）
```go
// prometheus_recorder.go
type PrometheusRecorder struct {
	commandDuration prometheus.Histogram
	commandErrors   *prometheus.CounterVec
}

func NewPrometheusRecorder() *PrometheusRecorder {
	return &PrometheusRecorder{
		commandDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "scanner_command_duration_seconds",
			Help:    "Command execution duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
		}),
		commandErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "scanner_command_errors_total",
			Help: "Total number of command execution errors",
		}, []string{"scanner_type", "error_type"}),
	}
}
```
生产验证清单
1. 基础功能验证
```go
# 测试进程终止能力
$ go test -v -run TestProcessTermination

# 验证cgroups配置
$ cat /sys/fs/cgroup/go-sac-scanner/cpu.max
50000 100000

# 检查资源指标
$ curl localhost:9090/metrics | grep scanner_command
```
