package app

import (
	vineapp "go.yorun.ai/vine/app"

	"example.com/greeting/src/server/repo"
)

// Migrator creates the business tables before the App serves traffic.
// Alternatively, a DAO can implement EnsureSchema; its default hook is empty.
type Migrator struct {
	vineapp.BaseModule

	GreetingDao *repo.GreetingDao        `inject:""`
	ReceiptDao  *repo.GreetingReceiptDao `inject:""`
}

func (m *Migrator) BeforeAppStart() error {
	if err := m.GreetingDao.GormDB().AutoMigrate(&repo.GreetingRecord{}); err != nil {
		return err
	}
	return m.ReceiptDao.GormDB().AutoMigrate(&repo.GreetingReceiptRecord{})
}
