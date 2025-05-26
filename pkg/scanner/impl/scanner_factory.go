package impl

import (
	"fmt"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
)

// ScannerFactoryImpl 扫描器工厂实现
type ScannerFactoryImpl struct {
	scanners map[domain.ScanType]scanner.Scanner
}

// NewScannerFactory 创建扫描器工厂
func NewScannerFactory() *ScannerFactoryImpl {
	return &ScannerFactoryImpl{
		scanners: map[domain.ScanType]scanner.Scanner{
			domain.ScanTypeStaticCodeAnalysis: NewSASTScanner(),
			domain.ScanTypeDast:               NewDASTScanner(),
			domain.ScanTypeSca:                NewSCAScanner(),
		},
	}
}

// GetScanner 实现工厂接口
func (f *ScannerFactoryImpl) GetScanner(scanType domain.ScanType) (scanner.Scanner, error) {
	if s, ok := f.scanners[scanType]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("unsupported scanner type: %s", scanType)
}
