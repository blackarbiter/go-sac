package logger

import (
	"os"

	"github.com/natefinch/lumberjack/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogWriter() zapcore.WriteSyncer {
	// 建议将日志文件大小调整为更合理的值（例如100MB）
	lumberJackLogger, err := lumberjack.NewRoller(
		"./logs/app.log",
		100*1024*1024, // 100MB（单位：字节）
		nil,           // 使用默认选项
	)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(lumberJackLogger)
}

// 可选：如需配置日志保留策略，可创建Options结构体
func GetLogWriterWithOptions() zapcore.WriteSyncer {
	opt := &lumberjack.Options{
		MaxAge:     30,   // 保留30天
		MaxBackups: 10,   // 保留10个备份
		Compress:   true, // 启用压缩
	}

	lumberJackLogger, err := lumberjack.NewRoller(
		"./logs/app.log",
		100*1024*1024, // 100MB
		opt,
	)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(lumberJackLogger)
}

// 保持原有初始化逻辑不变
func InitZapWithRotation(env string) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			GetLogWriterWithOptions(),
		),
		zap.InfoLevel,
	)

	Logger = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(Logger)
}
