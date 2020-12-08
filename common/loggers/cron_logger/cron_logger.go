package cron_logger

import (
	"fmt"
	"io"

	"com.github.gin-common/util"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CronLogger struct {
	logger *zap.SugaredLogger
	config Config
}

func New(config Config) *CronLogger {
	cronLogger := util.GetLogger(config.LogLevel, config.Writer, config.Options...)

	return &CronLogger{
		logger: cronLogger.Sugar(),
		config: config,
	}
}

type Config struct {
	LogLevel zapcore.Level
	Options  []zap.Option
	Writer   io.Writer
}

func (l *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infof(msg, keysAndValues)
}

func (l *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.logger.Errorf(fmt.Sprintf("%s:%s", msg, err.Error()), keysAndValues)
}
