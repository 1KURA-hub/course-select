package initialize

import (
	"go-course/global" // 确保这里替换成你项目真实的 global 包路径
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 初始化日志
func InitLogger() {
	// 配置日志切割
	fileWriter := &lumberjack.Logger{
		Filename:   global.Settings.Log.Filename,
		MaxSize:    global.Settings.Log.MaxSize,
		MaxBackups: global.Settings.Log.MaxBackups,
		MaxAge:     global.Settings.Log.MaxAge,
		Compress:   global.Settings.Log.Compress,
	}

	// 解析日志级别
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(global.Settings.Log.Level)); err != nil {
		logLevel = zapcore.InfoLevel
	}

	var core zapcore.Core

	// 区分环境配置
	if global.Settings.Server.Mode == "release" {
		// 生产环境 只写入日志 不写入终端 也不写行号
		encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

		writeSyncer := zapcore.AddSync(fileWriter)

		core = zapcore.NewCore(encoder, writeSyncer, logLevel)
		global.Logger = zap.New(core)
	} else {
		// 开发环境 双写并显示行号
		encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

		writeSyncer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter))

		core = zapcore.NewCore(encoder, writeSyncer, logLevel)

		global.Logger = zap.New(core, zap.AddCaller())
	}

	zap.ReplaceGlobals(global.Logger)
}
