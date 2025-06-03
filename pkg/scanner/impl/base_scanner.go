package scanner_impl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/blackarbiter/go-sac/pkg/config"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"go.uber.org/zap"
)

// TaskStatusUpdater 定义任务状态更新接口
type TaskStatusUpdater interface {
	UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) error
}

// ResultPublisher 定义结果发布接口
type ResultPublisher interface {
	PublishScanResult(ctx context.Context, result []byte) error
}

// BaseScanner provides common functionality for all scanners
type BaseScanner struct {
	scanType           domain.ScanType
	timeoutCtrl        *scanner.TimeoutController
	logger             *zap.Logger
	metricsRecorder    MetricsRecorder
	cgroupManager      CgroupManager
	securityProfile    SecurityProfile
	processManager     processManager
	meta               scanner.ExecutorMeta
	mu                 sync.RWMutex
	taskStatusUpdater  TaskStatusUpdater
	resultPublisher    ResultPublisher
	circuitBreaker     *scanner.CircuitBreaker // 熔断器
	defaultTimeout     time.Duration           // 默认超时时间
	gracefulStopPeriod time.Duration           // 优雅停止时间
	resourceProfile    scanner.ResourceProfile
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

func NewBaseScanner(
	scanType domain.ScanType,
	timeoutCtrl *scanner.TimeoutController,
	logger *zap.Logger,
	config *config.Config,
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
		defaultTimeout:     5 * time.Minute,  // 默认超时时间
		gracefulStopPeriod: 30 * time.Second, // 默认优雅停止时间
	}

	for _, opt := range opts {
		opt(bs)
	}

	bs.setupSignalHandling()
	return bs
}

// AsyncExecuteWithResult 异步执行带结果的扫描任务
func (s *BaseScanner) AsyncExecuteWithResult(ctx context.Context, task *domain.ScanTaskPayload, scanFunc func(context.Context) (*domain.ScanResult, error)) (string, error) {
	go func() {
		_, _ = s.ExecuteWithResult(ctx, task, scanFunc)
	}()
	return task.TaskID, nil
}

// classifyError 分类错误类型
func (s *BaseScanner) classifyError(err error) (string, scanner.ErrorType) {
	switch {
	case err == nil:
		return "success", scanner.TransientError
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout", scanner.TransientError
	case errors.Is(err, context.Canceled):
		return "canceled", scanner.TransientError
	case strings.Contains(err.Error(), "permission denied"):
		return "permission_denied", scanner.CriticalError
	case strings.Contains(err.Error(), "resource unavailable"):
		return "resource_unavailable", scanner.CriticalError
	default:
		return "runtime", scanner.TransientError
	}
}

// ExecuteCommand executes a command with proper process management and resource control
func (s *BaseScanner) ExecuteCommand(ctx context.Context, task *domain.ScanTaskPayload, cmd *exec.Cmd, tag string) error {
	// 检查熔断器状态
	if s.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker is open, task rejected")
	}

	// 1. 准备执行环境
	execID := fmt.Sprintf("%s-%d", task.TaskID, time.Now().UnixNano())
	if !strings.EqualFold("healthCheck", tag) {
		s.logger.Info("starting command execution",
			zap.String("task_id", task.TaskID),
			zap.String("exec_id", execID),
			zap.Strings("command", cmd.Args))
	}

	// 2. 设置进程属性
	//s.setProcessAttributes(cmd)

	// 3. 注册进程
	s.processManager.activeProcesses.Store(execID, cmd)
	defer func() {
		s.processManager.activeProcesses.Delete(execID)
		// 确保进程被清理
		if cmd.Process != nil {
			s.KillProcessGroup(cmd, true)
		}
	}()

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
		// 确保cgroup资源被清理
		defer func() {
			if err := s.cgroupManager.Cleanup(); err != nil {
				s.logger.Error("cgroup cleanup failed", zap.Error(err))
			}
		}()
	}

	// 6. 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, s.defaultTimeout)
	defer cancel()

	// 7. 异步等待进程结束
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
		close(done)
	}()

	// 8. 处理完成、超时或取消
	select {
	case err := <-done:
		execDuration := time.Since(startTime)
		_, errType := s.classifyError(err)
		s.recordCommandMetrics(task, cmd, err, execDuration)

		if err != nil {
			s.circuitBreaker.RecordFailure(errType)
		} else {
			s.circuitBreaker.RecordSuccess()
		}

		return err

	case <-timeoutCtx.Done():
		s.logger.Warn("command execution timeout",
			zap.String("task_id", task.TaskID),
			zap.String("exec_id", execID),
			zap.Duration("timeout", s.defaultTimeout))

		s.circuitBreaker.RecordFailure(scanner.TransientError)

		// 先尝试优雅停止
		s.KillProcessGroup(cmd, false)

		// 等待进程结束或超时
		gracefulCtx, cancel := context.WithTimeout(context.Background(), s.gracefulStopPeriod)
		defer cancel()

		select {
		case <-done:
			s.logger.Info("process stopped gracefully",
				zap.String("task_id", task.TaskID),
				zap.String("exec_id", execID))
		case <-gracefulCtx.Done():
			s.logger.Warn("graceful stop timeout, forcing kill",
				zap.String("task_id", task.TaskID),
				zap.String("exec_id", execID))
			s.KillProcessGroup(cmd, true)
		}

		return timeoutCtx.Err()

	case <-ctx.Done():
		s.logger.Warn("command execution canceled",
			zap.String("task_id", task.TaskID),
			zap.String("exec_id", execID))

		s.circuitBreaker.RecordFailure(scanner.TransientError)

		s.KillProcessGroup(cmd, true)
		return ctx.Err()
	}
}

// KillProcessGroup terminates a process group
func (s *BaseScanner) KillProcessGroup(cmd *exec.Cmd, force bool) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid

	// 检查进程是否已退出
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		s.logger.Debug("process already exited", zap.Int("pid", pid))
		return
	}

	// 实时检查进程是否存在
	if process, err := os.FindProcess(pid); err != nil || process == nil {
		s.logger.Debug("process not found", zap.Int("pid", pid))
		return
	}

	s.logger.Info("terminating process group",
		zap.Int("pid", pid),
		zap.Bool("force", force))

	switch runtime.GOOS {
	case "windows":
		s.killWindowsProcess(pid, force)
	default:
		s.killUnixProcess(pid, force)
	}

	// 确保进程状态更新
	if cmd.ProcessState == nil {
		if _, err := cmd.Process.Wait(); err != nil {
			s.logger.Debug("process wait result",
				zap.Int("pid", pid),
				zap.Error(err))
		}
	}
}

func (s *BaseScanner) killUnixProcess(pid int, force bool) {
	// 尝试终止整个进程组
	if err := syscall.Kill(-pid, 0); err == nil {
		sig := syscall.SIGTERM
		if force {
			sig = syscall.SIGKILL
		}

		if err := syscall.Kill(-pid, sig); err != nil && !errors.Is(err, syscall.ESRCH) {
			s.logger.Error("kill process group failed",
				zap.Int("pid", pid),
				zap.Error(err))
		}
		return
	}

	// 进程组不存在，尝试终止单个进程
	if err := syscall.Kill(pid, 0); err == nil {
		sig := syscall.SIGTERM
		if force {
			sig = syscall.SIGKILL
		}

		if err := syscall.Kill(pid, sig); err != nil && !errors.Is(err, syscall.ESRCH) {
			s.logger.Error("kill process failed",
				zap.Int("pid", pid),
				zap.Error(err))
		}
	} else {
		s.logger.Debug("process not found", zap.Int("pid", pid))
	}

	// 处理僵尸进程
	if _, err := syscall.Wait4(pid, nil, syscall.WNOHANG, nil); err != nil {
		s.logger.Debug("reaped zombie process", zap.Int("pid", pid))
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
	var wg sync.WaitGroup
	var activeCount int

	// 创建进程快照避免并发修改
	s.processManager.activeProcesses.Range(func(key, value interface{}) bool {
		if cmd, ok := value.(*exec.Cmd); ok && cmd.Process != nil {
			activeCount++
			wg.Add(1)
			go func(c *exec.Cmd) {
				defer wg.Done()
				s.KillProcessGroup(c, true)
			}(cmd)
		}
		return true
	})

	s.logger.Info("cleaning up processes",
		zap.Int("count", activeCount))

	// 等待所有终止操作完成
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Debug("all processes cleaned up")
	case <-time.After(5 * time.Second):
		s.logger.Warn("process cleanup timed out")
	}

	// 清理后重置活动进程表
	s.processManager.activeProcesses = sync.Map{}
}

// setProcessAttributes is implemented in platform-specific files
func (s *BaseScanner) setProcessAttributes(cmd *exec.Cmd) {
	// 通用属性
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	// Unix系统设置进程组
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr.Setpgid = true
		// 设置进程组ID为父进程PID
		// 确保整个进程组能被正确终止
		cmd.SysProcAttr.Pgid = syscall.Getpid()
	}

	// 安全相关设置
	if s.securityProfile.RunAsUser != nil {
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid:    uint32(*s.securityProfile.RunAsUser),
			Gid:    uint32(*s.securityProfile.RunAsGroup),
			Groups: []uint32{},
		}
	}

	// Linux 特定的安全设置
	if runtime.GOOS == "linux" && s.securityProfile.NoNewPrivs {
		// 在 Linux 上，我们需要使用 prctl 系统调用来设置 NoNewPrivs
		// 这需要在进程启动后通过其他方式实现
		s.logger.Info("NoNewPrivs requested, will be set after process start")
	}
}

// recordCommandMetrics 记录命令执行指标
func (s *BaseScanner) recordCommandMetrics(task *domain.ScanTaskPayload, cmd *exec.Cmd, err error, duration time.Duration) {
	if s.metricsRecorder == nil {
		return
	}

	errorType, _ := s.classifyError(err)
	tags := map[string]string{
		"scanner_type": s.scanType.String(),
		"task_id":      task.TaskID,
		"exit_code":    strconv.Itoa(cmd.ProcessState.ExitCode()),
		"error_type":   errorType,
		"asset_type":   string(task.AssetType),
		"asset_id":     task.AssetID,
		"command":      cmd.Path,
	}

	// 记录执行时间
	s.metricsRecorder.Record("command_duration_seconds", duration.Seconds(), tags)
	s.metricsRecorder.Record("command_cpu_seconds", cmd.ProcessState.SystemTime().Seconds(), tags)
	s.metricsRecorder.Record("command_user_seconds", cmd.ProcessState.UserTime().Seconds(), tags)

	// 记录命令状态
	if err != nil {
		s.metricsRecorder.Record("command_errors", 1, tags)
	} else {
		s.metricsRecorder.Record("command_success", 1, tags)
	}

	// 记录资源使用情况
	if s.cgroupManager != nil {
		// 这里可以添加资源使用指标的记录
		s.metricsRecorder.Gauge("command_memory_usage_bytes", 0, tags) // 需要实现实际的内存统计
		s.metricsRecorder.Gauge("command_cpu_usage_percent", 0, tags)  // 需要实现实际的 CPU 统计
	}
}

// Meta implements TaskExecutor interface
func (s *BaseScanner) Meta() scanner.ExecutorMeta {
	return s.meta
}

// HealthCheck implements TaskExecutor interface
func (s *BaseScanner) HealthCheck() error {
	// 检查cgroups是否可用
	if s.cgroupManager != nil {
		if err := s.cgroupManager.Apply(os.Getpid()); err != nil {
			return fmt.Errorf("cgroup health check failed: %w", err)
		}
	}

	// 检查进程清理能力
	testCmd := exec.Command("sleep", "0")
	if err := s.ExecuteCommand(context.Background(), &domain.ScanTaskPayload{}, testCmd, "healthCheck"); err != nil {
		return fmt.Errorf("process execution test failed: %w", err)
	}

	return nil
}

// WithResourceProfile sets the resource profile
func WithResourceProfile(rp scanner.ResourceProfile) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.meta.ResourceProfile = rp
		bs.resourceProfile = rp
	}
}

// WithSecurityProfile sets the security profile
func WithSecurityProfile(user, group int, noNewPrivs bool) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.securityProfile = SecurityProfile{
			RunAsUser:  &user,
			RunAsGroup: &group,
			NoNewPrivs: noNewPrivs,
		}
	}
}

// WithCgroupManager sets the cgroup manager
func WithCgroupManager(cm CgroupManager) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.cgroupManager = cm
	}
}

// WithMetricsRecorder sets the metrics recorder
func WithMetricsRecorder(mr MetricsRecorder) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.metricsRecorder = mr
	}
}

// WithCircuitBreaker 设置熔断器配置
func WithCircuitBreaker(config *config.Config) BaseScannerOption {
	threshold, criticalThreshold, resetTimeout := config.GetCircuitBreakerConfig()
	return func(bs *BaseScanner) {
		bs.circuitBreaker = scanner.NewCircuitBreaker(threshold, criticalThreshold, resetTimeout)
	}
}

// ExecuteWithTimeout 执行带超时控制的通用任务
func (s *BaseScanner) ExecuteWithTimeout(ctx context.Context, task *domain.ScanTaskPayload, fn func(context.Context) error) error {
	// 1. 准备执行环境
	execID := fmt.Sprintf("%s-%d", task.TaskID, time.Now().UnixNano())
	s.logger.Info("starting task execution",
		zap.String("task_id", task.TaskID),
		zap.String("exec_id", execID))

	// 2. 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, s.defaultTimeout)
	defer cancel()

	// 3. 创建结果通道
	done := make(chan error, 1)
	startTime := time.Now()

	// 4. 异步执行任务
	go func() {
		done <- fn(timeoutCtx)
		close(done)
	}()

	// 5. 处理完成、超时或取消
	select {
	case err := <-done:
		execDuration := time.Since(startTime)
		s.recordTaskMetrics(task, err, execDuration)
		return err

	case <-timeoutCtx.Done():
		s.logger.Warn("task execution timeout",
			zap.String("task_id", task.TaskID),
			zap.String("exec_id", execID),
			zap.Duration("timeout", s.defaultTimeout))
		s.recordTaskMetrics(task, timeoutCtx.Err(), time.Since(startTime))
		return timeoutCtx.Err()

	case <-ctx.Done():
		s.logger.Warn("task execution canceled",
			zap.String("task_id", task.TaskID),
			zap.String("exec_id", execID))
		s.recordTaskMetrics(task, ctx.Err(), time.Since(startTime))
		return ctx.Err()
	}
}

// recordTaskMetrics 记录任务执行指标
func (s *BaseScanner) recordTaskMetrics(task *domain.ScanTaskPayload, err error, duration time.Duration) {
	if s.metricsRecorder == nil {
		return
	}

	errorType, _ := s.classifyError(err)
	tags := map[string]string{
		"scanner_type": s.scanType.String(),
		"task_id":      task.TaskID,
		"error_type":   errorType,
		"asset_type":   string(task.AssetType),
		"asset_id":     task.AssetID,
	}

	// 记录执行时间
	s.metricsRecorder.Record("task_duration_seconds", duration.Seconds(), tags)

	// 记录任务状态
	if err != nil {
		s.metricsRecorder.Record("task_errors", 1, tags)
	} else {
		s.metricsRecorder.Record("task_success", 1, tags)
	}

	// 记录资源使用情况
	if s.cgroupManager != nil {
		// 这里可以添加资源使用指标的记录
		s.metricsRecorder.Gauge("task_memory_usage_bytes", 0, tags) // 需要实现实际的内存统计
		s.metricsRecorder.Gauge("task_cpu_usage_percent", 0, tags)  // 需要实现实际的 CPU 统计
	}
}

// ExecuteWithResult 执行扫描任务并处理结果
func (b *BaseScanner) ExecuteWithResult(ctx context.Context, task *domain.ScanTaskPayload, scanFunc func(context.Context) (*domain.ScanResult, error)) (*domain.ScanResult, error) {
	// 更新任务状态为运行中
	if err := b.UpdateTaskStatus(ctx, task.TaskID, domain.TaskStatusRunning); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// 执行扫描任务
	result, err := scanFunc(ctx)
	if err != nil {
		// 更新任务状态为失败
		_ = b.UpdateTaskStatus(ctx, task.TaskID, domain.TaskStatusFailed)
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// 更新任务状态为完成
	if err := b.UpdateTaskStatus(ctx, task.TaskID, domain.TaskStatusCompleted); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// 发布扫描结果
	if err := b.PublishScanResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to publish scan result: %w", err)
	}

	return result, nil
}

// UpdateTaskStatus 更新任务状态
func (b *BaseScanner) UpdateTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus) error {
	if b.taskStatusUpdater == nil {
		return nil
	}
	return b.taskStatusUpdater.UpdateTaskStatus(ctx, taskID, status)
}

// PublishScanResult 发布扫描结果
func (b *BaseScanner) PublishScanResult(ctx context.Context, result *domain.ScanResult) error {
	if b.resultPublisher == nil {
		return nil
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	return b.resultPublisher.PublishScanResult(ctx, resultBytes)
}

// SetTaskStatusUpdater 设置任务状态更新器
func (b *BaseScanner) SetTaskStatusUpdater(updater TaskStatusUpdater) {
	b.taskStatusUpdater = updater
}

// SetResultPublisher 设置结果发布器
func (b *BaseScanner) SetResultPublisher(publisher ResultPublisher) {
	b.resultPublisher = publisher
}

// WithTimeout 设置超时选项
func WithTimeout(defaultTimeout, gracefulStopPeriod time.Duration) BaseScannerOption {
	return func(bs *BaseScanner) {
		bs.defaultTimeout = defaultTimeout
		bs.gracefulStopPeriod = gracefulStopPeriod
	}
}
