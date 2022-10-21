package main

import (
	"fmt"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger
var WarnColor = "\033[1;31m%s\033[0m\n"

func init() {
	// writeSyncer := getLogWriter()
	writeSyncer := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		getLogWriter(),
	)
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	Logger = zap.New(core, zap.AddCaller()).Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	// encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	logPath := os.Getenv("LogPath")
	if logPath == "" {
		fmt.Printf(WarnColor, "logpath is empty,use log path in ./update.log")
		logPath = "./update.log"
	}
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    5,
		MaxBackups: 5,
		MaxAge:     24,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}
func Info(args ...interface{}) {
	Logger.Info(args...)
}
func Error(args ...interface{}) {
	Logger.Error(args...)
}
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}
