// pkg/storage/mysql/migration.go
package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type MigrationConfig struct {
	MigrationsPath string        // 迁移文件路径
	Timeout        time.Duration // 超时设置
	TargetVersion  uint          // 目标迁移版本，0表示迁移到最新
}

type DBMigration struct {
	logger *zap.Logger
	config MigrationConfig
}

func NewMigration(cfg MigrationConfig, logger *zap.Logger) *DBMigration {
	return &DBMigration{
		logger: logger.Named("mysql.migration"),
		config: cfg,
	}
}

// 规范化DSN，确保mysql驱动能够正确识别
func formatDSN(dsn string) string {
	// 添加mysql://前缀，确保migrate能够正确识别
	return "mysql://" + dsn
}

func (m *DBMigration) Run(dsn string) error {
	startTime := time.Now()
	m.logger.Info("开始执行数据库迁移",
		zap.String("dsn", dsn),
		zap.String("migrations_path", m.config.MigrationsPath))

	_, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

	// 格式化DSN
	formattedDSN := formatDSN(dsn)

	migrator, err := migrate.New(
		"file://"+m.config.MigrationsPath,
		formattedDSN,
	)
	if err != nil {
		m.logger.Error("创建迁移实例失败",
			zap.Error(err))
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}

	// 如果指定了目标版本，迁移到特定版本，否则迁移到最新
	var migrationErr error
	if m.config.TargetVersion > 0 {
		migrationErr = migrator.Migrate(m.config.TargetVersion)
		if migrationErr != nil && migrationErr != migrate.ErrNoChange {
			m.logger.Error("迁移到特定版本失败",
				zap.Uint("target_version", m.config.TargetVersion),
				zap.Error(migrationErr))
			return fmt.Errorf("迁移到版本 %d 失败: %w", m.config.TargetVersion, migrationErr)
		}
	} else {
		migrationErr = migrator.Up()
		if migrationErr != nil && migrationErr != migrate.ErrNoChange {
			m.logger.Error("数据库迁移执行失败",
				zap.Error(migrationErr))
			return fmt.Errorf("数据库迁移执行失败: %w", migrationErr)
		}
	}

	version, dirty, verr := migrator.Version()
	if verr != nil && verr != migrate.ErrNilVersion {
		m.logger.Error("获取当前迁移版本失败",
			zap.Error(verr))
	} else {
		m.logger.Info("当前数据库迁移版本",
			zap.Uint("current_version", version),
			zap.Bool("dirty", dirty))
	}

	m.logger.Info("数据库迁移完成",
		zap.Duration("duration", time.Since(startTime)))
	return nil
}

// 执行特定的迁移方向（Up或Down）
func (m *DBMigration) RunWithDirection(dsn string, up bool) error {
	// 格式化DSN
	formattedDSN := formatDSN(dsn)

	migrator, err := migrate.New(
		"file://"+m.config.MigrationsPath,
		formattedDSN,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}

	if up {
		err = migrator.Up()
	} else {
		err = migrator.Down()
	}

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("迁移执行失败: %w", err)
	}
	return nil
}

// 执行向上迁移N步
func (m *DBMigration) StepUp(dsn string, steps int) error {
	formattedDSN := formatDSN(dsn)
	migrator, err := migrate.New(
		"file://"+m.config.MigrationsPath,
		formattedDSN,
	)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}

	err = migrator.Steps(steps)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("迁移步骤执行失败: %w", err)
	}
	return nil
}

// 获取当前迁移版本
func (m *DBMigration) GetVersion(dsn string) (uint, bool, error) {
	formattedDSN := formatDSN(dsn)
	migrator, err := migrate.New(
		"file://"+m.config.MigrationsPath,
		formattedDSN,
	)
	if err != nil {
		return 0, false, fmt.Errorf("创建迁移实例失败: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("获取版本失败: %w", err)
	}
	return version, dirty, nil
}
