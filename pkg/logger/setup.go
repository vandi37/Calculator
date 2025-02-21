package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var dateLayout = "02'01'06 3:04:05"

func Setup() *zap.Logger {
	return zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey:     ">>",
			LevelKey:       "!!",
			TimeKey:        "#",
			NameKey:        "",
			CallerKey:      "",
			StacktraceKey:  "stack",
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout(dateLayout),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}),
		zapcore.Lock(os.Stderr),
		zap.NewAtomicLevelAt(zap.DebugLevel),
	), zap.AddStacktrace(zap.ErrorLevel))
}
