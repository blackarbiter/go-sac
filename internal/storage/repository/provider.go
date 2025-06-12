package repository

import (
	mysqlStorage "github.com/blackarbiter/go-sac/pkg/storage/mysql"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 是 repository 层的依赖注入集合
var ProviderSet = wire.NewSet(
	NewStorageRepository,
	mysqlStorage.ProviderSet,
)

// NewStorageRepository 创建存储服务的数据访问层实例
func NewStorageRepository(db *gorm.DB) StorageRepository {
	return &storageRepository{
		db: db,
	}
}
