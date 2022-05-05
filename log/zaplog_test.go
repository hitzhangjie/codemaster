package log_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func NewStructuredLogger(options ...zap.Option) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, zap.DebugLevel)
	return zap.New(core).WithOptions(options...)
}

func NewConsoleLogger(options ...zap.Option) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), os.Stdout, zap.DebugLevel)
	return zap.New(core).WithOptions(options...)
}

type consoleEncoder struct {
	zapcore.Encoder
	pool buffer.Pool
}

func (c consoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buffer := c.pool.Get()

	buffer.AppendString(entry.Time.Format("2006/01/02 15:04:05") + " ")
	buffer.AppendString(strings.ToUpper(entry.Level.String()) + " ")
	c.appendFields(buffer, fields)
	buffer.AppendString(entry.Message + "\n")

	return buffer, nil
}

func (c consoleEncoder) appendFields(buffer *buffer.Buffer, fields []zapcore.Field) {
	for _, f := range fields {
		buffer.AppendString(fmt.Sprintf("%s:%s", f.Key, f.String))
	}
	buffer.AppendString(" ")
}

func NewConsoleLogger2(options ...zap.Option) *zap.Logger {
	core := zapcore.NewCore(&consoleEncoder{
		Encoder: zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			ConsoleSeparator: "\t",
		}),
		pool: buffer.NewPool(),
	}, os.Stdout, zap.DebugLevel)
	return zap.New(core).WithOptions(options...)
}

func TestZapLogger(t *testing.T) {
	f := []zap.Field{zap.String("uid", "zhijiezhang")}

	l := NewStructuredLogger()
	l.Info("helloworld", f...)

	l = NewConsoleLogger()
	l.Info("helloworld", f...)

	l = NewConsoleLogger2()
	l.Info("helloworld", f...)
}
