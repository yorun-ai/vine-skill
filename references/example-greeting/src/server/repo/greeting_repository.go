package repo

import (
	"github.com/google/uuid"

	"go.yorun.ai/vine/core/skel"

	skeled "example.com/greeting/skeled/golang"
)

type GreetingRepositoryImpl struct {
	Dao *GreetingDao `inject:""`
}

func (r *GreetingRepositoryImpl) Create(g skeled.Greeting) skeled.Greeting {
	g.Id = skel.NewUUID(uuid.New())
	record := greetingRecordFromSkeled(g)
	r.Dao.Create(record)
	return record.toSkeled()
}

func (r *GreetingRepositoryImpl) List(offset, limit int) []skeled.Greeting {
	q := r.Dao.Query()
	if limit > 0 {
		q = q.Limit(limit)
	}
	records := q.Offset(offset).Order("id DESC").List()
	greetings := make([]skeled.Greeting, 0, len(records))
	for _, record := range records {
		greetings = append(greetings, record.toSkeled())
	}
	return greetings
}
