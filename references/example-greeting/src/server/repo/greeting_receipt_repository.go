package repo

import (
	"uuid"

	skeltype "go.yorun.ai/skel/types"
	"go.yorun.ai/vine/core/ex"
	"go.yorun.ai/vine/infra/rdb"
	"gorm.io/gorm/clause"

	skeled "example.com/greeting/skeled/golang"
)

type GreetingReceiptRecord struct {
	rdb.Model
	JobUUID      string `gorm:"column:job_uuid;uniqueIndex"`
	GreetingUUID string `gorm:"column:greeting_uuid;uniqueIndex"`
	Message      string
}

type GreetingReceiptDao struct {
	rdb.Dao[*GreetingReceiptRecord]
}

type GreetingReceiptRepositoryImpl struct {
	Dao *GreetingReceiptDao `inject:""`
}

func (r *GreetingReceiptRepositoryImpl) Prepare(jobId skeltype.UUID, greeting skeled.Greeting) {
	// The unique job ID and insert form one atomic operation, including under retries.
	// A repeated job preserves the first receipt and its preparation time.
	record := &GreetingReceiptRecord{
		JobUUID:      jobId.String(),
		GreetingUUID: greeting.Id.String(),
		Message:      "Greeting received: " + greeting.Message,
	}
	ex.PanicIfError(r.Dao.GormDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "job_uuid"}},
		DoNothing: true,
	}).Create(record).Error)
}

func (r *GreetingReceiptRepositoryImpl) Find(greetingId skeltype.UUID) *skeled.GreetingReceipt {
	record, found := r.Dao.First("greeting_uuid = ?", greetingId.String())
	if !found {
		return nil
	}
	return &skeled.GreetingReceipt{
		JobId:      skeltype.NewUUID(uuid.MustParse(record.JobUUID)),
		GreetingId: skeltype.NewUUID(uuid.MustParse(record.GreetingUUID)),
		Message:    record.Message,
		PreparedAt: skeltype.NewTimestamp(record.CreatedAt),
	}
}
