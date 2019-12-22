package log

import (
	"io"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	l    *zap.Logger
	once *sync.Once
)

func init() {
	once = new(sync.Once)
}

// Init intialize *zap.Logger
func Init(output io.Writer) {
	once.Do(func() {
		encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})
		core := zapcore.NewCore(encoder, zapcore.AddSync(output), zapcore.InfoLevel)
		l = zap.New(core)
	})
}

// L returns *zap.Logger
func L() *zap.Logger {
	return l
}
