package scanner_impl

import (
	"github.com/blackarbiter/go-sac/pkg/config"
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
	cfg *config.Config,
) map[domain.ScanType]scanner.TaskExecutor {
	commonOpts := []BaseScannerOption{
		WithMetricsRecorder(metrics),
		WithCgroupManager(cgroup),
	}

	return map[domain.ScanType]scanner.TaskExecutor{
		domain.ScanTypeStaticCodeAnalysis: NewSASTScanner(timeoutCtrl, logger, cfg, commonOpts...),
		domain.ScanTypeDast:               NewDASTScanner(timeoutCtrl, logger, cfg, commonOpts...),
		domain.ScanTypeSca:                NewSCAScanner(timeoutCtrl, logger, cfg, commonOpts...),
	}
}
