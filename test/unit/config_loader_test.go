// test/unit/config_loader_test.go
package unit_test

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
	"os"
	"testing"

	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadBaseConfig(t *testing.T) {
	t.Cleanup(func() {
		os.Clearenv() // 清理环境变量
	})

	wd, _ := os.Getwd()
	t.Logf("Current working directory: %s", wd) // 打印当前路径

	// 执行配置加载
	cfg, err := config.Load("scan")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err) // 提前终止避免空指针
	}
	r1, r2, r3 := cfg.GetScannerConfig(domain.ScanTypeSca)
	t.Logf("%v,%v,%v", r1, r2, r3)
	g1, g2, g3 := cfg.GetCircuitBreakerConfig()
	t.Logf("%v,%v,%v", g1, g2, g3)
	h1, h2 := cfg.GetConcurrencyConfig()
	t.Logf("%v,%v", h1, h2)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Database.MySQL.Host)
	assert.Equal(t, 3306, cfg.Database.MySQL.Port)
	assert.Equal(t, 20, cfg.Database.MySQL.MaxOpenConns)
	assert.Equal(t, 10, cfg.MQ.RabbitMQ.Consumer.PrefetchCount)
}
