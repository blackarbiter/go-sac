// pkg/domain/lifecycle.go
package domain

import "fmt"

// LifecyclePhase 定义DevSecOps生命周期阶段
type LifecyclePhase uint8

const (
	LifecyclePhasePlanning    LifecyclePhase = iota + 1 // 计划
	LifecyclePhaseDevelopment                           // 开发
	LifecyclePhaseBuilding                              // 构建
	LifecyclePhaseTesting                               // 测试
	LifecyclePhaseRelease                               // 发布
	LifecyclePhaseDelivery                              // 交付
	LifecyclePhaseDeployment                            // 部署
	LifecyclePhaseOperation                             // 运营
	LifecyclePhaseMonitoring                            // 监控
	LifecyclePhaseFeedback                              // 反馈

	// 自动补充安全关键阶段
	LifecyclePhaseSecurityHardening // 安全加固
	LifecyclePhaseComplianceAudit   // 合规审计
)

// String 返回可读阶段名称
func (p LifecyclePhase) String() string {
	switch p {
	case LifecyclePhasePlanning:
		return "Planning"
	case LifecyclePhaseDevelopment:
		return "Development"
	case LifecyclePhaseBuilding:
		return "Building"
	case LifecyclePhaseTesting:
		return "Testing"
	case LifecyclePhaseRelease:
		return "Release"
	case LifecyclePhaseDelivery:
		return "Delivery"
	case LifecyclePhaseDeployment:
		return "Deployment"
	case LifecyclePhaseOperation:
		return "Operation"
	case LifecyclePhaseMonitoring:
		return "Monitoring"
	case LifecyclePhaseFeedback:
		return "Feedback"
	case LifecyclePhaseSecurityHardening:
		return "SecurityHardening"
	case LifecyclePhaseComplianceAudit:
		return "ComplianceAudit"
	default:
		return "Unknown"
	}
}

// ParseLifecyclePhase 从字符串解析阶段
func ParseLifecyclePhase(s string) (LifecyclePhase, error) {
	switch s {
	case "Planning":
		return LifecyclePhasePlanning, nil
	case "Development":
		return LifecyclePhaseDevelopment, nil
	case "Building":
		return LifecyclePhaseBuilding, nil
	case "Testing":
		return LifecyclePhaseTesting, nil
	case "Release":
		return LifecyclePhaseRelease, nil
	case "Delivery":
		return LifecyclePhaseDelivery, nil
	case "Deployment":
		return LifecyclePhaseDeployment, nil
	case "Operation":
		return LifecyclePhaseOperation, nil
	case "Monitoring":
		return LifecyclePhaseMonitoring, nil
	case "Feedback":
		return LifecyclePhaseFeedback, nil
	case "SecurityHardening":
		return LifecyclePhaseSecurityHardening, nil
	case "ComplianceAudit":
		return LifecyclePhaseComplianceAudit, nil
	default:
		return 0, fmt.Errorf("invalid lifecycle phase: %s", s)
	}
}
