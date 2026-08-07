package repo

import (
	"github.com/google/uuid"

	"go.yorun.ai/vine/core/skel"
	"go.yorun.ai/vine/infra/rdb"

	skeled "example.com/greeting/skeled/golang"
)

type GreetingRecord struct {
	rdb.Model
	UUID    string `gorm:"column:uuid;uniqueIndex"`
	Message string `gorm:"column:message"`
}

type GreetingDao struct{ rdb.Dao[*GreetingRecord] }

func (r *GreetingRecord) toSkeled() skeled.Greeting {
	return skeled.Greeting{
		Id:        skel.NewUUID(uuid.MustParse(r.UUID)),
		Message:   r.Message,
		CreatedAt: skel.NewTimestamp(r.CreatedAt),
	}
}

func greetingRecordFromSkeled(g skeled.Greeting) *GreetingRecord {
	return &GreetingRecord{
		UUID:    g.Id.String(),
		Message: g.Message,
	}
}
