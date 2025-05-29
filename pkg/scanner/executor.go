package scanner

import (
	"context"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
)

// ExecutorMeta represents executor metadata
type ExecutorMeta struct {
	Type            string
	Version         string
	SupportedTypes  []domain.ScanType
	ResourceProfile ResourceProfile
}

// ResourceProfile describes executor resource requirements
type ResourceProfile struct {
	MinCPU      int
	MaxCPU      int
	MemoryMB    int
	RequiresGPU bool
}

// TaskExecutor defines the interface for task execution
type TaskExecutor interface {
	Meta() ExecutorMeta
	AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (resultHandle string, err error)
	Cancel(handle string) error
	GetStatus(handle string) (domain.TaskStatus, error)
	HealthCheck() error
}

// ContainerConfig represents production-level container configuration
type ContainerConfig struct {
	DefaultImage      string
	ImagePullPolicy   string
	ResourceQuota     ResourceQuota
	SecurityContext   SecurityConfig
	SidecarContainers []SidecarConfig
	DeadlineBuffer    time.Duration
}

type ResourceQuota struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

type SecurityConfig struct {
	RunAsUser                int64
	ReadOnlyRootFs           bool
	Privileged               bool
	AllowPrivilegeEscalation bool
}

type SidecarConfig struct {
	Image       string
	Command     []string
	Environment map[string]string
}
