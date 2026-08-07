package app

import (
	vineapp "go.yorun.ai/vine/app"

	"example.com/greeting/src/server/repo"
)

// Migrator creates the business tables before the App serves traffic.
// Vine RDB never migrates business tables automatically.
type Migrator struct {
	vineapp.BaseModule

	GreetingDao *repo.GreetingDao `inject:""`
}

func (m *Migrator) BeforeAppStart() error {
	return m.GreetingDao.GormDB().AutoMigrate(&repo.GreetingRecord{})
}
