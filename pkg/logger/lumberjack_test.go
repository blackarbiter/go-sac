package logger_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blackarbiter/go-sac/pkg/logger"
	"go.uber.org/zap"
)

func TestLogRotation(t *testing.T) {
	// 准备测试
	logDir := "./logs"
	logFile := filepath.Join(logDir, "app.log")

	// 确保目录存在
	os.MkdirAll(logDir, 0755)

	// 清理可能存在的旧日志文件
	os.Remove(logFile)

	// 初始化带轮换的日志
	logger.InitZapWithRotation("production")

	t.Run("基本日志写入", func(t *testing.T) {
		// 写入测试日志
		logger.Logger.Info("测试轮换日志系统")

		// 验证日志文件是否已创建
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("日志文件未被创建")
		}

		// 读取日志文件内容
		content, err := ioutil.ReadFile(logFile)
		if err != nil {
			t.Fatalf("无法读取日志文件: %v", err)
		}

		// 验证日志内容
		if !strings.Contains(string(content), "测试轮换日志系统") {
			t.Error("日志内容未正确写入")
		}
	})
}

func TestGetLogWriter(t *testing.T) {
	// 获取日志写入器
	writer := logger.GetLogWriter()

	// 简单校验写入器是否可用
	_, err := writer.Write([]byte("测试日志写入器\n"))
	if err != nil {
		t.Errorf("写入器写入失败: %v", err)
	}
}

func TestGetLogWriterWithOptions(t *testing.T) {
	// 获取带选项的日志写入器
	writer := logger.GetLogWriterWithOptions()

	// 验证写入功能
	_, err := writer.Write([]byte("测试带选项的日志写入器\n"))
	if err != nil {
		t.Errorf("带选项的写入器写入失败: %v", err)
	}
}

// 注意：完整测试日志轮换功能需要写入大量数据或模拟时间经过
// 这里提供一个简化版本的测试，实际可能需要更多的模拟和配置
func TestLogRotationSimulation(t *testing.T) {
	// 清理环境
	logDir := "./logs"
	logFile := filepath.Join(logDir, "app.log")
	os.MkdirAll(logDir, 0755)
	os.Remove(logFile)

	// 初始化带轮换的日志系统
	logger.InitZapWithRotation("production")

	// 写入少量日志
	for i := 0; i < 100; i++ {
		logger.Logger.Info("模拟日志数据",
			zap.Int("index", i),
			zap.String("data", strings.Repeat("测试数据", 100))) // 添加一些数据量
	}

	// 检查日志文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("主日志文件未创建")
	}

	// 注意：在单元测试中很难真实测试轮换，因为：
	// 1. 需要写入足够多的数据触发大小限制
	// 2. 需要等待足够长时间触发时间限制
	// 一个完整测试可能需要集成测试或模拟lumberjack内部行为
}
