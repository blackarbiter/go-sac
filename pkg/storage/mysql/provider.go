package mysql

import (
	"time"

	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet 是数据库提供者集合
var ProviderSet = wire.NewSet(
	ProvideGormDB,
)

// ProvideGormDB 提供GORM数据库连接
func ProvideGormDB(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	// 创建MySQL连接器配置
	mysqlConfig := ConnectorConfig{
		DSN:             cfg.GetMySQLDSN(),
		MaxOpenConns:    cfg.Database.MySQL.MaxOpenConns,
		MaxIdleConns:    10, // 默认空闲连接数
		ConnMaxLifetime: time.Hour,
		ConnTimeout:     10 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      time.Second,
	}

	// 创建MySQL连接器
	connector, err := NewConnector(mysqlConfig, logger)
	if err != nil {
		return nil, err
	}

	// 获取数据库连接
	sqlDB := connector.GetDB()

	// 使用gorm打开连接
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return db, nil
}
