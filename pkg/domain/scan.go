// pkg/domain/scan.go
package domain

import "fmt"

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

// ParseScanType 从字符串解析扫描类型
func ParseScanType(s string) (ScanType, error) {
	switch s {
	case "RequirementAnalysis":
		return ScanTypeRequirementAnalysis, nil
	case "ThreatModeling":
		return ScanTypeThreatModeling, nil
	case "SecuritySpecCheck":
		return ScanTypeSecuritySpecCheck, nil
	case "SAST":
		return ScanTypeStaticCodeAnalysis, nil
	case "ContainerImageScan":
		return ScanTypeContainerImageScan, nil
	case "HostSecurityCheck":
		return ScanTypeHostSecurityCheck, nil
	case "BlackBoxTesting":
		return ScanTypeBlackBoxTesting, nil
	case "PortScanning":
		return ScanTypePortScanning, nil
	case "DAST":
		return ScanTypeDast, nil
	case "SCA":
		return ScanTypeSca, nil
	case "SecretsDetection":
		return ScanTypeSecretsDetection, nil
	case "ComplianceAudit":
		return ScanTypeComplianceAudit, nil
	default:
		return ScanTypeUnknown, fmt.Errorf("unknown scan type: %s", s)
	}
}
