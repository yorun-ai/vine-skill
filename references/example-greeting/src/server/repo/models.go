package repo

import (
	"uuid"

	skeltype "go.yorun.ai/skel/types"
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
		Id:        skeltype.NewUUID(uuid.MustParse(r.UUID)),
		Message:   r.Message,
		CreatedAt: skeltype.NewTimestamp(r.CreatedAt),
	}
}

func greetingRecordFromSkeled(g skeled.Greeting) *GreetingRecord {
	return &GreetingRecord{
		UUID:    g.Id.String(),
		Message: g.Message,
	}
}
