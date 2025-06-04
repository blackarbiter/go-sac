package handlers

import (
	"github.com/blackarbiter/go-sac/internal/asset/service"
)

// 全局资产处理器实例
var assetHandler *AssetHandler

// InitHandlers 初始化所有处理程序
func InitHandlers(assetService service.AssetService) {
	assetHandler = NewAssetHandler(assetService)
}

// GetAssetHandler 获取资产处理器实例
func GetAssetHandler() *AssetHandler {
	return assetHandler
}
