package app

import (
	"os"
	"testing"

	"go.yorun.ai/vine/app/testkit"
	"go.yorun.ai/vine/core/ex"
	"go.yorun.ai/vine/core/meta"
	"go.yorun.ai/vine/core/skel"

	skeled "example.com/greeting/skeled/golang"
)

func startRuntime(t *testing.T) *testkit.Runtime {
	t.Helper()
	// Use the test-only seed (no portalSites/portalRules) so the standalone does
	// not bind real portal ports (7288/7299). Tests stay independent of local
	// port availability.
	// The business RDB uses a fixed SQLite file path; remove any data left by
	// earlier runs so each test package starts from a clean database.
	_ = os.Remove("greeting.sqlite")
	return testkit.StartStandalone[*GreetingApp](t, testkit.Option{
		SeedYAMLFile: "../seed/test-hub.yaml",
	})
}

func TestGreetingSuite(t *testing.T) {
	r := startRuntime(t)
	execution := r.NewExecution(testkit.ExecutionOption{
		Actor: meta.NewAnonymousActor(),
	})
	client := testkit.NewClient[skeled.GreetingServiceClient](execution)
	auth := testkit.NewClient[skeled.AuthServiceClient](execution)
	welcome := testkit.NewClient[skeled.WelcomeServiceClient](execution)

	t.Run("login with default credentials", func(t *testing.T) {
		result := auth.Login("admin", "admin123")
		if !result.Success {
			t.Fatalf("expected successful login with default credentials, got %+v", result)
		}
		if result.Username == nil || *result.Username != "admin" {
			t.Fatalf("expected username admin, got %+v", result.Username)
		}
		if result.Token == nil || *result.Token == "" {
			t.Fatal("expected a session token on successful login")
		}
	})

	t.Run("login rejected on wrong password", func(t *testing.T) {
		result := auth.Login("admin", "wrong")
		if result.Success {
			t.Fatalf("expected failed login for wrong password, got %+v", result)
		}
	})

	t.Run("welcome requires an authenticated actor", func(t *testing.T) {
		// An anonymous execution has no actor info, so the welcome method
		// panics with Unauthorized (plain-style error handling).
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected a panic without authentication, got none")
			}
			err, ok := r.(ex.Error)
			if !ok {
				t.Fatalf("expected ex.Error panic, got %T: %v", r, r)
			}
			if err.Code() != ex.Unauthorized {
				t.Fatalf("expected Unauthorized, got %s", err.Code())
			}
		}()
		welcome.GetWelcome()
	})

	t.Run("welcome accepts actor info after admission", func(t *testing.T) {
		// Core tests cover token validation. This test covers the post-admission
		// path where Portal has already supplied authenticated actor info.
		authed := r.NewExecution(testkit.ExecutionOption{
			Actor: meta.NewAuthenticatedActor(&skeled.GreetingActorInfo{Username: "admin"}),
		})
		w := testkit.NewClient[skeled.WelcomeServiceClient](authed)
		got := w.GetWelcome()
		if got.Username != "admin" {
			t.Fatalf("expected welcome for admin, got %+v", got)
		}
		if got.Message != "Welcome back, admin" {
			t.Fatalf("unexpected welcome message %q", got.Message)
		}
	})

	t.Run("create and list greeting", func(t *testing.T) {
		created := client.CreateGreeting("hello from vine")
		if created.Id == (skel.UUID{}) {
			t.Fatal("expected a generated greeting id")
		}
		if created.Message != "hello from vine" {
			t.Fatalf("unexpected message %q", created.Message)
		}

		page := client.ListGreetings(nil, nil)
		if len(page.Items) != 1 || page.Items[0].Id != created.Id {
			t.Fatalf("expected 1 greeting, got %+v", page.Items)
		}
	})

	t.Run("empty message rejected", func(t *testing.T) {
		// Default (non-ER) style expresses business errors by panic; the plain
		// client re-panics via ex.PanicIfError, so assert via recover.
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected a panic for empty message, got none")
			}
			err, ok := r.(ex.Error)
			if !ok {
				t.Fatalf("expected ex.Error panic, got %T: %v", r, r)
			}
			if err.Code() != ex.ValidationFailed {
				t.Fatalf("expected ValidationFailed, got %s", err.Code())
			}
		}()
		client.CreateGreeting("")
	})
}
