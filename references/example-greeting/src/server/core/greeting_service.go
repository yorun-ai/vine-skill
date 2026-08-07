package core

import (
	"strconv"

	"go.yorun.ai/vine/core/ex"

	skeled "example.com/greeting/skeled/golang"
)

type GreetingService struct {
	Repo   GreetingRepository     `inject:""`
	Config *skeled.GreetingConfig `inject:""`
}

func (s *GreetingService) Create(message string) skeled.Greeting {
	if message == "" {
		ex.PanicNew(ex.ValidationFailed, "message must not be empty")
	}
	return s.Repo.Create(skeled.Greeting{Message: message})
}

func (s *GreetingService) List(pageToken *string, pageSize *int) skeled.GreetingPage {
	offset, limit := s.page(pageToken, pageSize)
	greetings := s.Repo.List(offset, limit)
	return skeled.GreetingPage{
		Items:     greetings,
		NextToken: nextPageToken(len(greetings), limit, offset),
	}
}

func (s *GreetingService) page(pageToken *string, pageSize *int) (offset, limit int) {
	limit = s.Config.DefaultPageSize
	if pageSize != nil && *pageSize > 0 {
		limit = *pageSize
	}
	if limit < 1 {
		limit = 20
	}
	offset = 0
	if pageToken != nil {
		if n, err := strconv.Atoi(*pageToken); err == nil && n > 0 {
			offset = n
		}
	}
	return offset, limit
}

func nextPageToken(itemCount, limit, offset int) *string {
	if limit > 0 && itemCount == limit {
		token := strconv.Itoa(offset + limit)
		return &token
	}
	return nil
}
