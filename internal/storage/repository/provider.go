package repository

import (
	mysqlStorage "github.com/blackarbiter/go-sac/pkg/storage/mysql"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 是 repository 层的依赖注入集合
var ProviderSet = wire.NewSet(
	ProvideRepository,
	mysqlStorage.ProviderSet,
)

// ProvideRepository 提供资产仓库实例
func ProvideRepository(db *gorm.DB) Repository {
	gp := NewGormRepository(db)
	err := gp.AutoMigrate()
	if err != nil {
		panic(err)
	}
	return gp
}
