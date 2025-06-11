// pkg/domain/security_material.go
package domain

import (
	"fmt"
	"strings"
)

// SecurityMaterialType 定义安全物料分类
type SecurityMaterialType uint8

const (
	SecurityMaterialTypeUnknown          SecurityMaterialType = iota
	SecurityMaterialTypeSensitiveWords                        // 需求敏感词
	SecurityMaterialTypeThreatModel                           // 威胁模型
	SecurityMaterialTypeSecuritySpec                          // 安全规范
	SecurityMaterialTypeToolkit                               // 安全工具包
	SecurityMaterialTypeScanRule                              // 安全扫描规则
	SecurityMaterialTypeSecurityStandard                      // 安全标准（自动补充）
	SecurityMaterialTypeComplianceDoc                         // 合规文档（自动补充）
)

// String 返回可读类型名称
func (t SecurityMaterialType) String() string {
	switch t {
	case SecurityMaterialTypeSensitiveWords:
		return "SensitiveWords"
	case SecurityMaterialTypeThreatModel:
		return "ThreatModel"
	case SecurityMaterialTypeSecuritySpec:
		return "SecuritySpec"
	case SecurityMaterialTypeToolkit:
		return "Toolkit"
	case SecurityMaterialTypeScanRule:
		return "ScanRule"
	case SecurityMaterialTypeSecurityStandard:
		return "SecurityStandard"
	case SecurityMaterialTypeComplianceDoc:
		return "ComplianceDoc"
	default:
		return "Unknown"
	}
}

func ParseSecurityMaterialType(s string) (SecurityMaterialType, error) {
	// 统一转换为小写进行匹配（不改变原始错误提示）
	lowerInput := strings.ToLower(s)

	switch lowerInput {
	case "sensitivewords":
		return SecurityMaterialTypeSensitiveWords, nil
	case "threatmodel":
		return SecurityMaterialTypeThreatModel, nil
	case "securityspec":
		return SecurityMaterialTypeSecuritySpec, nil
	case "toolkit":
		return SecurityMaterialTypeToolkit, nil
	case "scanrule":
		return SecurityMaterialTypeScanRule, nil
	case "securitystandard":
		return SecurityMaterialTypeSecurityStandard, nil
	case "compliancedoc":
		return SecurityMaterialTypeComplianceDoc, nil
	default:
		// 错误信息保留原始输入（便于定位问题）
		return SecurityMaterialTypeUnknown, fmt.Errorf("unknown security material type: %s", s)
	}
}
