package service

import (
	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/google/wire"
)

// ProviderSet 是 service 层的依赖注入集合
var ProviderSet = wire.NewSet(
	ProvideAssetService,
)

// ProvideAssetService 提供资产服务实例
func ProvideAssetService(repo repository.AssetRepository, cfg *config.Config) AssetService {
	return NewAssetService(repo)
}
