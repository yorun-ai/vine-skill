package core

import (
	skeled "example.com/greeting/skeled/golang"
)

type GreetingRepository interface {
	Create(g skeled.Greeting) skeled.Greeting
	List(offset, limit int) []skeled.Greeting
}
