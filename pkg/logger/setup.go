package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var dateLayout = "02'01'06 3:04:05"

func Setup() zapcore.Core {
	return zapcore.NewCore(
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
	)
}

func Prod(writer zapcore.WriteSyncer) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "lvl",
			TimeKey:        "t",
			NameKey:        "name",
			CallerKey:      "call",
			StacktraceKey:  "stack",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}),
		zapcore.Lock(writer),
		zap.NewAtomicLevelAt(zap.DebugLevel),
	)
}

func ProdFile(name string) (zapcore.Core, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return Prod(file), err
}

func ConsoleAndFile(name string) *zap.Logger {
	console := Setup()
	file, err := ProdFile(name)
	if err != nil {
		zap.New(console, zap.AddStacktrace(zap.ErrorLevel)).Fatal("can't load logging file", zap.String("name", name))
	}
	return zap.New(zapcore.NewTee(console, file), zap.AddStacktrace(zap.ErrorLevel))
}
