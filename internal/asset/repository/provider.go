package repository

import (
	"github.com/blackarbiter/go-sac/internal/asset/repository/migration"
	mysqlStorage "github.com/blackarbiter/go-sac/pkg/storage/mysql"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 是资产仓库提供者集合
var ProviderSet = wire.NewSet(
	mysqlStorage.ProviderSet,
	ProvideRepository,
)

// ProvideRepository 提供资产仓库实例
func ProvideRepository(db *gorm.DB) Repository {
	// 执行数据库迁移
	if err := migration.AutoMigrate(db); err != nil {
		panic(err) // 在启动时如果迁移失败，应该直接panic
	}

	return NewGormRepository(db)
}
