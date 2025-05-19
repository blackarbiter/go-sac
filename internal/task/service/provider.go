package service

import (
	"github.com/google/wire"
)

// ProviderSet 是任务服务提供者集合
var ProviderSet = wire.NewSet(
	NewTaskService,
)
