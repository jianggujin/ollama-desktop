package log

import (
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"ollama-desktop/internal/config"
	"os"
)

var (
	Logger zerolog.Logger
)

func init() {
	level, err := zerolog.ParseLevel(config.Config.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = config.Config.Logging.TimeFormat
	// consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: config.Config.Logging.TimeFormat}
	// 未配置日志文件，必须初始化控制台日志
	if config.Config.Logging.Filename == "" {
		Logger = zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
		return
	}
	lumberjackLogger := &lumberjack.Logger{
		Filename:   config.Config.Logging.Filename,   // 日志文件的位置
		MaxSize:    config.Config.Logging.MaxSize,    // 文件最大尺寸（以MB为单位）
		MaxBackups: config.Config.Logging.MaxBackups, // 保留的最大旧文件数量
		MaxAge:     config.Config.Logging.MaxAge,     // 保留旧文件的最大天数
		Compress:   config.Config.Logging.Compress,   // 是否压缩/归档旧文件
		LocalTime:  config.Config.Logging.LocalTime,  // 使用本地时间创建时间戳
	}
	Logger = zerolog.New(lumberjackLogger).Level(level).With().Timestamp().Logger()
}

func Info() *zerolog.Event {
	return Logger.Info()
}
func Debug() *zerolog.Event {
	return Logger.Debug()
}

func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Error() *zerolog.Event {
	return Logger.Error()
}

func Fatal() *zerolog.Event {
	return Logger.Fatal()
}
func Trace() *zerolog.Event {
	return Logger.Trace()
}

//func ParseLogLevel(l string) zerolog.Level {
//	if l == "" {
//		return zerolog.GlobalLevel()
//	}
//	level, err := zerolog.ParseLevel(l)
//	if err != nil {
//		return zerolog.GlobalLevel()
//	}
//	return level
//}
