// pkg/domain/security_material.go
package domain

import "fmt"

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

// ParseSecurityMaterialType 从字符串解析物料类型
func ParseSecurityMaterialType(s string) (SecurityMaterialType, error) {
	switch s {
	case "SensitiveWords":
		return SecurityMaterialTypeSensitiveWords, nil
	case "ThreatModel":
		return SecurityMaterialTypeThreatModel, nil
	case "SecuritySpec":
		return SecurityMaterialTypeSecuritySpec, nil
	case "Toolkit":
		return SecurityMaterialTypeToolkit, nil
	case "ScanRule":
		return SecurityMaterialTypeScanRule, nil
	case "SecurityStandard":
		return SecurityMaterialTypeSecurityStandard, nil
	case "ComplianceDoc":
		return SecurityMaterialTypeComplianceDoc, nil
	default:
		return SecurityMaterialTypeUnknown, fmt.Errorf("unknown security material type: %s", s)
	}
}
