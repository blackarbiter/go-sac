package scanner_impl

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"go.uber.org/zap"
)

// CreateDefaultScanners creates the default set of scanners
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
