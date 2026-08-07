package app

import (
	"go.yorun.ai/vine/infra/rdb"

	"example.com/greeting/src/server/repo"
)

type MainDatabase struct{ rdb.Database }

func (*MainDatabase) InitOption(option *rdb.Option) {
	option.ConnURL = "sqlite://./greeting.sqlite"
	option.MaxOpenConn = 10
}

func (*MainDatabase) InitDao(add rdb.TypeAdder) {
	add(rdb.T[*repo.GreetingDao]())
}
