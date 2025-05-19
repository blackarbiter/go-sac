package repository

import (
	"github.com/google/wire"
)

// ProviderSet 是任务仓库提供者集合
var ProviderSet = wire.NewSet(
	NewTaskRepository,
)
