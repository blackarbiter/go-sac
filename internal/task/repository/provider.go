package repository

import (
	"github.com/blackarbiter/go-sac/pkg/config"
	mysqlStorage "github.com/blackarbiter/go-sac/pkg/storage/mysql"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 是任务仓库提供者集合
var ProviderSet = wire.NewSet(
	ProvideTaskRepository,
	mysqlStorage.ProviderSet,
)

// ProvideTaskRepository 提供任务仓库实例
func ProvideTaskRepository(db *gorm.DB, cfg *config.Config) TaskRepository {
	// 自动迁移表结构
	if err := db.AutoMigrate(&TaskEntity{}); err != nil {
		panic(err) // 在启动时如果迁移失败，应该直接panic
	}

	return NewTaskRepository(db, cfg)
}
