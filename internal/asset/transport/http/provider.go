package http

import "github.com/google/wire"

// ProviderSet 是 HTTP 层的依赖注入集合
var ProviderSet = wire.NewSet(
	NewServer,
	NewAssetBinder,
)
