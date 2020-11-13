package gorm_plugin

import "gorm.io/gorm"

type GormCachePlugin struct{}

func (GormCachePlugin) Name() string {
	panic("implement me")
}

func (GormCachePlugin) Initialize(db *gorm.DB) error {
	return nil
}
