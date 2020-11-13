package db_tool

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"

	"com.github.gin-common/common/loggers/gorm_logger"
	"go.uber.org/zap"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"com.github.gin-common/util"
)

var once sync.Once
var db *gorm.DB

func getMySqlDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		util.GetDefaultEnv("DATABASE_USER", ""),
		util.GetDefaultEnv("DATABASE_PASSWORD", ""),
		util.GetDefaultEnv("DATABASE_HOST", "127.0.0.1"),
		util.GetDefaultEnv("DATABASE_PORT", "3306"),
		util.GetDefaultEnv("DATABASE_NAME", ""),
		util.GetDefaultEnv("DATABASE_LOC", "Asia%2FShanghai"))
}

func getGormConfig(writer io.Writer, options ...zap.Option) (*gorm.Config, error) {
	var slowThreshold time.Duration
	var logLevel zapcore.Level
	t := util.GetDefaultEnv("GORM_SLOW_THRESHOLD", "200")
	l := util.GetDefaultEnv("GORM_LOG_LEVEL", "warn")
	logLevel = util.GetLogLevel(l)

	if t != "" {
		ivt, err := strconv.Atoi(t)
		if err != nil {
			return nil, err
		}
		slowThreshold = time.Millisecond * time.Duration(ivt)
	}
	return &gorm.Config{
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: gorm_logger.New(gorm_logger.Config{
			SlowThreshold: slowThreshold,
			LogLevel:      logLevel,
			Options:       options,
			Writer:        writer,
		}),
		SkipDefaultTransaction: true,
	}, nil
}

func GetDB() *gorm.DB {
	// 获取数据库连接（单例），连接池
	once.Do(func() {
		var err error
		var sqlDB *sql.DB
		var maxIdle, maxConn, maxLifeTime int
		var config *gorm.Config

		config, err = getGormConfig(os.Stdout, zap.AddCaller())

		util.PanicError(err)
		db, err = gorm.Open(mysql.Open(getMySqlDsn()), config)
		util.PanicError(err)

		// 获取通用数据库对象 sql.DB ，然后使用其提供的功能
		sqlDB, err = db.DB()
		util.PanicError(err)

		maxIdle, err = strconv.Atoi(util.GetDefaultEnv("DATABASE_MAX_IDLE", "20"))
		util.PanicError(err)
		maxConn, err = strconv.Atoi(util.GetDefaultEnv("DATABASE_MAX_CONN", "20"))
		util.PanicError(err)

		maxLifeTime, err = strconv.Atoi(util.GetDefaultEnv("DATABASE_MAX_LIFETIME", strconv.Itoa(2*60*60)))
		util.PanicError(err)
		// 设置最大空闲连接数
		sqlDB.SetMaxIdleConns(maxIdle)
		// 设置打开数据库连接的最大数量。
		sqlDB.SetMaxOpenConns(maxConn)
		// 设置了连接可复用的最大时间（单位秒）。
		sqlDB.SetConnMaxLifetime(time.Duration(maxLifeTime) * time.Second)
	})

	return db
}
