package handlers

import (
	"github.com/blackarbiter/go-sac/internal/storage/service"
)

// 全局存储处理器实例
var storageHandler *StorageHandler

// InitHandlers 初始化所有处理程序
func InitHandlers(storageService service.StorageService) {
	storageHandler = NewStorageHandler(storageService)
}

// GetStorageHandler 获取存储处理器实例
func GetStorageHandler() *StorageHandler {
	return storageHandler
}
