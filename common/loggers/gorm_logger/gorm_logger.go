package gorm_logger

// 使用zap实现的gorm Logger

import (
	"context"
	"fmt"
	"io"
	"time"

	"gorm.io/gorm/utils"

	"com.github.gin-common/util"

	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"

	"gorm.io/gorm/logger"
)

type GormLogger struct {
	logger *zap.Logger
	config Config
}

type Config struct {
	SlowThreshold time.Duration
	LogLevel      zapcore.Level
	Options       []zap.Option
	Writer        io.Writer
}

var gormZapLogLevelMapping = map[logger.LogLevel]zapcore.Level{
	logger.Silent: zap.DebugLevel,
	logger.Info:   zap.InfoLevel,
	logger.Warn:   zap.WarnLevel,
	logger.Error:  zap.ErrorLevel,
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	var ll zapcore.Level
	if logLevel, ok := gormZapLogLevelMapping[level]; ok {
		ll = logLevel
	} else {
		ll = zap.ErrorLevel
	}
	var options = make([]zap.Option, len(l.config.Options))
	copy(options, l.config.Options)

	options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		syncer := zapcore.AddSync(l.config.Writer)
		return zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), syncer, ll)
	}))
	l.logger = l.logger.WithOptions(options...)
	l.config.LogLevel = ll
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, data...), zap.String("sql_caller", utils.FileWithLineNum()))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Warn(fmt.Sprintf(msg, data...), zap.String("sql_caller", utils.FileWithLineNum()))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, data...), zap.String("sql_caller", utils.FileWithLineNum()))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	if l.config.LogLevel == zap.DebugLevel {
		sql, rows := fc()
		l.logger.Debug("", zap.String("sql", sql), zap.Int64("rows", rows), zap.Float64("cost", float64(elapsed.Nanoseconds())/1e6), zap.String("logType", "debugQuery"), zap.String("sql_caller", utils.FileWithLineNum()))
		return
	}

	if err != nil {
		sql, rows := fc()
		l.logger.Warn(err.Error(), zap.String("sql", sql), zap.Int64("rows", rows), zap.Float64("cost", float64(elapsed.Nanoseconds())/1e6), zap.String("logType", "errQuery"), zap.String("sql_caller", utils.FileWithLineNum()))
		return
	}
	if elapsed > l.config.SlowThreshold && l.config.SlowThreshold != 0 {
		sql, rows := fc()
		l.logger.Warn("", zap.String("sql", sql), zap.Int64("rows", rows), zap.Float64("cost", float64(elapsed.Nanoseconds())/1e6), zap.String("logType", "slowQuery"), zap.String("sql_caller", utils.FileWithLineNum()))
		return
	}
}

func New(config Config) logger.Interface {
	gormLogger := util.GetLogger(config.LogLevel, config.Writer, config.Options...)
	return &GormLogger{
		logger: gormLogger,
		config: config,
	}
}
