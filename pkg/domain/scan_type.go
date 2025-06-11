// pkg/domain/scan_type.go
package domain

import (
	"fmt"
	"strings"
)

// ScanType 定义安全扫描类型
type ScanType uint8

const (
	ScanTypeUnknown             ScanType = iota
	ScanTypeRequirementAnalysis          // 风险需求识别
	ScanTypeThreatModeling               // 威胁建模
	ScanTypeSecuritySpecCheck            // 安全规范扫描
	ScanTypeStaticCodeAnalysis           // 静态代码扫描
	ScanTypeContainerImageScan           // 镜像扫描
	ScanTypeHostSecurityCheck            // 主机安全扫描
	ScanTypeBlackBoxTesting              // 黑盒扫描
	ScanTypePortScanning                 // 端口扫描
	ScanTypeDast                         // 动态应用扫描（DAST）
	ScanTypeSca                          // 软件成分分析（SCA）
	ScanTypeSecretsDetection             // 敏感信息扫描
	ScanTypeComplianceAudit              // 合规性扫描
)

// String 返回可读类型名称
func (t ScanType) String() string {
	switch t {
	case ScanTypeRequirementAnalysis:
		return "RequirementAnalysis"
	case ScanTypeThreatModeling:
		return "ThreatModeling"
	case ScanTypeSecuritySpecCheck:
		return "SecuritySpecCheck"
	case ScanTypeStaticCodeAnalysis:
		return "SAST"
	case ScanTypeContainerImageScan:
		return "ContainerImageScan"
	case ScanTypeHostSecurityCheck:
		return "HostSecurityCheck"
	case ScanTypeBlackBoxTesting:
		return "BlackBoxTesting"
	case ScanTypePortScanning:
		return "PortScanning"
	case ScanTypeDast:
		return "DAST"
	case ScanTypeSca:
		return "SCA"
	case ScanTypeSecretsDetection:
		return "SecretsDetection"
	case ScanTypeComplianceAudit:
		return "ComplianceAudit"
	default:
		return "Unknown"
	}
}

func ParseScanType(s string) (ScanType, error) {
	// 统一转换为小写进行匹配（保留原始错误信息）
	lowerInput := strings.ToLower(s)

	switch lowerInput {
	case "requirementanalysis":
		return ScanTypeRequirementAnalysis, nil
	case "threatmodeling":
		return ScanTypeThreatModeling, nil
	case "securityspeccheck":
		return ScanTypeSecuritySpecCheck, nil
	case "sast":
		return ScanTypeStaticCodeAnalysis, nil
	case "containerimagescan":
		return ScanTypeContainerImageScan, nil
	case "hostsecuritycheck":
		return ScanTypeHostSecurityCheck, nil
	case "blackboxtesting":
		return ScanTypeBlackBoxTesting, nil
	case "portscanning":
		return ScanTypePortScanning, nil
	case "dast":
		return ScanTypeDast, nil
	case "sca":
		return ScanTypeSca, nil
	case "secretsdetection":
		return ScanTypeSecretsDetection, nil
	case "complianceaudit":
		return ScanTypeComplianceAudit, nil
	default:
		// 错误信息保留原始输入（便于调试）
		return ScanTypeUnknown, fmt.Errorf("unknown scan type: %s", s)
	}
}
