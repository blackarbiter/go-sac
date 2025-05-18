package logger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blackarbiter/go-sac/pkg/logger"
	"go.uber.org/zap"
)

func TestInitZap(t *testing.T) {
	// 准备测试
	logDir := "./logs"
	logFile := filepath.Join(logDir, "app.log")

	// 确保目录存在
	os.MkdirAll(logDir, 0755)

	// 删除可能存在的旧日志文件
	os.Remove(logFile)

	t.Run("开发环境配置", func(t *testing.T) {
		// 初始化开发环境日志
		logger.InitZap("development")

		// 验证全局logger是否已设置
		if logger.Logger == nil {
			t.Error("Logger未被初始化")
		}

		// 写入测试日志
		logger.Logger.Info("开发环境测试日志")

		// 验证日志文件是否已创建
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("日志文件未被创建")
		}
	})

	t.Run("生产环境配置", func(t *testing.T) {
		// 清理旧文件
		os.Remove(logFile)

		// 初始化生产环境日志
		logger.InitZap("production")

		// 写入测试日志
		logger.Logger.Info("生产环境测试日志")

		// 验证日志文件是否已创建
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("日志文件未被创建")
		}

		// 验证格式（这里只能间接验证）
		logger.Logger.Info("测试结构化日志", zap.String("key", "value"), zap.Int("code", 200))
	})
}

func TestZapGlobals(t *testing.T) {
	// 初始化日志
	logger.InitZap("development")

	// 测试全局日志记录器
	zap.L().Info("通过全局记录器写入的日志")

	// 如果没有panic就认为测试通过
	// 这主要测试全局替换是否成功
}
