package logging

import (
	"fmt"
	"mignon-ssh-port-forworder-dev/app/pkg/constant"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 是全局导出的日志器实例。其他包导入 mylog 后即可使用 mylog.Logger.Info(...)
var Logger *zap.Logger

// 定义自定义时间格式
const customTimeLayout = "2006-01-02:15:04:05.000"

// CustomTimeEncoder 将时间格式化为 yyyy-mm-dd:hh:mm:ss 毫秒三位

func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(customTimeLayout))
}

func init() {
	Logger = InitLogger(constant.IconstantInstance.LoggerPath)
}

// InitLogger 初始化并返回配置好的 Zap Logger
func InitLogger(logFilePath string) *zap.Logger {
	// --- 1. 确保日志目录存在 ---
	logDir := filepath.Dir(logFilePath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("无法创建日志目录: %s, 错误: %v\n", logDir, err)
			// 如果失败，返回一个功能正常的空日志器，不影响程序运行
			return zap.NewNop()
		}
	}

	// --- 2. 配置控制台 Core (带颜色和自定义格式) ---
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoderConfig.EncodeTime = CustomTimeEncoder
	consoleEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)
	consoleLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, consoleLevel)

	// --- 3. 配置文件 Core (轮转和自定义文本格式) ---
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	fileEncoderConfig := zap.NewDevelopmentEncoderConfig()
	fileEncoderConfig.EncodeTime = CustomTimeEncoder
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 文件不带颜色
	fileEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	fileEncoder := zapcore.NewConsoleEncoder(fileEncoderConfig)
	fileWriter := zapcore.AddSync(lumberJackLogger)
	fileLevel := zap.NewAtomicLevelAt(zap.InfoLevel)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, fileLevel)

	// --- 4. 组合并返回 Logger ---
	multiCore := zapcore.NewTee(consoleCore, fileCore)

	logger := zap.New(multiCore,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return logger
}

// Close 负责在程序退出前同步日志缓冲区。
// 必须在主程序的 main 函数中通过 defer 调用。
func Close() {
	if Logger != nil {
		if err := Logger.Sync(); err != nil {
			// 在 Sync 失败时，降级使用标准库 log 打印错误
			fmt.Printf("FATAL: Failed to sync logger buffer: %v\n", err)
		}
	}
}
