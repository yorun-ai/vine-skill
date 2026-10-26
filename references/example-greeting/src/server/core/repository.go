package core

import (
	skeltype "go.yorun.ai/skel/types"

	skeled "example.com/greeting/skeled/golang"
)

type GreetingRepository interface {
	Create(g skeled.Greeting) skeled.Greeting
	List(offset, limit int) []skeled.Greeting
}

type GreetingReceiptRepository interface {
	Prepare(jobId skeltype.UUID, greeting skeled.Greeting)
	Find(greetingId skeltype.UUID) *skeled.GreetingReceipt
}
