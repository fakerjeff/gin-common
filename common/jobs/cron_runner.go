package jobs

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"com.github.gin-common/common/loggers/cron_logger"

	"com.github.gin-common/util"
	"github.com/robfig/cron/v3"
)

var once sync.Once
var c *cron.Cron

func GetCron() *cron.Cron {
	once.Do(func() {

		loc, err := time.LoadLocation(util.GetDefaultEnv("LOCATION", "Asia/Shanghai"))

		if err != nil {
			panic(err)
		}
		l := util.GetDefaultEnv("CRON_LOG_LEVEL", "warn")
		logLevel := util.GetLogLevel(l)
		c = cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
		)), cron.WithLocation(loc), cron.WithLogger(cron_logger.New(cron_logger.Config{
			LogLevel: logLevel,
			Writer:   os.Stdout,
			Options:  []zap.Option{zap.AddCaller()},
		})))
	})
	return c
}
