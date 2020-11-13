package migrate

import (
	"com.github.gin-common/app/model"
	"com.github.gin-common/tools/db_tool"
)

func doMigrate(dst ...interface{}) {
	err := db_tool.GetDB().AutoMigrate(dst...)
	if err != nil {
		panic(err)
	}
}

func Migrate() {
	doMigrate(model.User{})
}
