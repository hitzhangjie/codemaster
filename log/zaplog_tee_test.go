package log

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test_Tee(t *testing.T) {
	os.RemoveAll("mylog.debug")
	os.RemoveAll("mylog.error")

	// Create log files for debug and error levels
	debugLogFile, err := os.Create("mylog.debug")
	if err != nil {
		panic(err)
	}
	errorLogFile, err := os.Create("mylog.error")
	if err != nil {
		panic(err)
	}

	// Create encoders for log files
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create cores for log files
	debugCore := zapcore.NewCore(encoder, zapcore.AddSync(debugLogFile), zapcore.DebugLevel)
	errorCore := zapcore.NewCore(encoder, zapcore.AddSync(zapcore.AddSync(errorLogFile)), zapcore.ErrorLevel)

	// Create a tee core that duplicates log messages to both debugCore and errorCore
	teeCore := zapcore.NewTee(debugCore, errorCore)

	// Create a logger that writes to both debugCore and errorCore
	logger := zap.New(teeCore, zap.AddCaller(), zap.AddCallerSkip(1))

	// Change log level
	// 这个logger的日志级别提升到error后，复制的时候就不会复制到debugcore，知会复制到errorcore
	logger = logger.WithOptions(zap.IncreaseLevel(zapcore.ErrorLevel))

	// Write log messages with different severity levels
	logger.Debug("debug message")
	logger.Error("error message")

	// Sync log files to ensure all log messages are written
	debugLogFile.Sync()
	errorLogFile.Sync()
}
