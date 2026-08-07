package core

import (
	"testing"

	"go.yorun.ai/vine/core/ex"

	skeled "example.com/greeting/skeled/golang"
)

func TestSessionTokenAuthentication(t *testing.T) {
	sessions := NewSessionStore()
	login := &AuthService{Sessions: sessions}
	auth := &GreetingActorAuthService{Sessions: sessions}

	result := login.Login(DefaultUsername, DefaultPassword)
	if !result.Success || result.Token == nil || *result.Token == "" {
		t.Fatalf("expected login to issue a token, got %+v", result)
	}

	info := auth.Auth(skeled.GreetingActorCredential{Token: *result.Token})
	if info.Username != DefaultUsername {
		t.Fatalf("expected authenticated username %q, got %q", DefaultUsername, info.Username)
	}

	t.Run("reject invalid token", func(t *testing.T) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				t.Fatal("expected invalid token to panic with Unauthorized")
			}
			err, ok := recovered.(ex.Error)
			if !ok {
				t.Fatalf("expected ex.Error panic, got %T: %v", recovered, recovered)
			}
			if err.Code() != ex.Unauthorized {
				t.Fatalf("expected Unauthorized, got %s", err.Code())
			}
		}()
		auth.Auth(skeled.GreetingActorCredential{Token: "invalid-token"})
	})
}
