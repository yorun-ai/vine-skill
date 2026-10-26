package impl

import (
	skeltype "go.yorun.ai/skel/types"

	skeled "example.com/greeting/skeled/golang"
	"example.com/greeting/src/server/core"
)

type GreetingReceiptRunner struct {
	skeled.DefaultGreetingReceiptTaskRunner
	Receipts core.GreetingReceiptRepository `inject:""`
}

func (h *GreetingReceiptRunner) RunPrepare(jobId skeltype.UUID, greeting skeled.Greeting) {
	h.Receipts.Prepare(jobId, greeting)
}
