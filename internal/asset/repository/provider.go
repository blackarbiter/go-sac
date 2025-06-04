package repository

import (
	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/pkg/config"
	mysqlStorage "github.com/blackarbiter/go-sac/pkg/storage/mysql"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 是资产仓库提供者集合
var ProviderSet = wire.NewSet(
	ProvideAssetRepository,
	mysqlStorage.ProviderSet,
)

// ProvideAssetRepository 提供资产仓库实例
func ProvideAssetRepository(db *gorm.DB, cfg *config.Config) AssetRepository {
	// 自动迁移表结构
	if err := db.AutoMigrate(&dto.Asset{}); err != nil {
		panic(err) // 在启动时如果迁移失败，应该直接panic
	}

	return NewAssetRepository(db)
}
