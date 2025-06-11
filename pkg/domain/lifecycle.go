// pkg/domain/lifecycle.go
package domain

import (
	"fmt"
	"strings"
)

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

func ParseLifecyclePhase(s string) (LifecyclePhase, error) {
	// 统一转换为小写进行匹配（保留原始错误信息）
	lowerInput := strings.ToLower(s)

	switch lowerInput {
	case "planning":
		return LifecyclePhasePlanning, nil
	case "development":
		return LifecyclePhaseDevelopment, nil
	case "building":
		return LifecyclePhaseBuilding, nil
	case "testing":
		return LifecyclePhaseTesting, nil
	case "release":
		return LifecyclePhaseRelease, nil
	case "delivery":
		return LifecyclePhaseDelivery, nil
	case "deployment":
		return LifecyclePhaseDeployment, nil
	case "operation":
		return LifecyclePhaseOperation, nil
	case "monitoring":
		return LifecyclePhaseMonitoring, nil
	case "feedback":
		return LifecyclePhaseFeedback, nil
	case "securityhardening":
		return LifecyclePhaseSecurityHardening, nil
	case "complianceaudit":
		return LifecyclePhaseComplianceAudit, nil
	default:
		// 错误信息保留原始输入值
		return 0, fmt.Errorf("invalid lifecycle phase: %s", s)
	}
}
