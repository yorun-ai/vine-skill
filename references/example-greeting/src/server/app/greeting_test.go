package app

import (
	"bytes"
	"database/sql"
	"encoding/json/v2"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.yorun.ai/vine/app/testkit"
	"gopkg.in/yaml.v3"
)

// Exercise API services and admission through their real Portal boundary,
// using the checked-in seed on an isolated loopback port.
func TestGreetingSuite(t *testing.T) {
	seed, err := os.ReadFile("../seed/hub.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := yaml.Unmarshal(seed, &config); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	config["portalEntries"].([]any)[0].(map[string]any)["http"].(map[string]any)["httpPort"] = port
	seed, err = yaml.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	workdir := t.TempDir()
	seedPath := filepath.Join(workdir, "hub.yaml")
	if err := os.WriteFile(seedPath, seed, 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(workdir)
	testkit.StartStandalone[*greetingTestApp](t, testkit.Option{SeedHubDataFile: seedPath})
	client := &http.Client{Timeout: 5 * time.Second}
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/api/invoke/", port)

	invoke := func(service string, method string, params any, token string) (int, string, map[string]any) {
		t.Helper()
		body, err := json.Marshal(map[string]any{"params": params})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, endpoint+"greeting."+service+"/"+method, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Accept", "application/vrpc+json")
		req.Header.Set("Content-Type", "application/vrpc+json")
		req.Header.Set("Vrpc-Trace", "id=123e4567e89b12d3a456426614174000")
		req.Header.Set("Vrpc-Client", "name=greeting.test,version=1.0.0,instanceId=123e4567-e89b-12d3-a456-426614174001")
		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		payload, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		var envelope map[string]any
		if err := json.Unmarshal(payload, &envelope); err != nil {
			t.Fatalf("invalid vRPC response: %v", err)
		}
		return resp.StatusCode, resp.Header.Get("Vrpc-Status"), envelope
	}

	var login map[string]any
	deadline := time.Now().Add(10 * time.Second)
	for {
		status, code, envelope := invoke("AuthApiService", "login", map[string]any{"username": "admin", "password": "admin123"}, "")
		if status == 200 && code == "OK" {
			login = envelope["result"].(map[string]any)
			break
		}
		if status != 503 || time.Now().After(deadline) {
			t.Fatalf("Portal did not become ready: HTTP %d, %s", status, code)
		}
		time.Sleep(50 * time.Millisecond)
	}
	token, ok := login["token"].(string)
	if !ok || token == "" {
		t.Fatal("login did not issue a token")
	}

	t.Run("admission", func(t *testing.T) {
		for _, supplied := range []string{"", "invalid-token"} {
			status, code, _ := invoke("WelcomeApiService", "getWelcome", map[string]any{}, supplied)
			if status != 401 || code != "UNAUTHORIZED" {
				t.Fatalf("expected rejected credential: HTTP %d, %s", status, code)
			}
		}
		status, code, response := invoke("WelcomeApiService", "getWelcome", map[string]any{}, token)
		if status != 200 || code != "OK" {
			t.Fatalf("valid credential rejected: HTTP %d, %s", status, code)
		}
		if response["result"].(map[string]any)["username"] != "admin" {
			t.Fatal("authenticated actor info was not propagated")
		}
		// optional admits omitted credentials and rejects invalid supplied credentials.
		status, code, _ = invoke("AuthApiService", "login", map[string]any{"username": "admin", "password": "admin123"}, "invalid-token")
		if status != 401 || code != "UNAUTHORIZED" {
			t.Fatalf("optional auth accepted invalid credentials: HTTP %d, %s", status, code)
		}
	})
	t.Run("persistence and business error", func(t *testing.T) {
		status, code, response := invoke("GreetingApiService", "createGreeting", map[string]any{"message": "hello from vine"}, "")
		if status != 200 || code != "OK" {
			t.Fatalf("create failed: HTTP %d, %s", status, code)
		}
		id := response["result"].(map[string]any)["id"]
		status, code, response = invoke("GreetingApiService", "listGreetings", map[string]any{"pageToken": nil, "pageSize": nil}, "")
		if status != 200 || code != "OK" {
			t.Fatalf("list failed: HTTP %d, %s", status, code)
		}
		items := response["result"].(map[string]any)["items"].([]any)
		if len(items) != 1 || items[0].(map[string]any)["id"] != id {
			t.Fatal("created greeting was not persisted")
		}
		status, code, _ = invoke("GreetingApiService", "createGreeting", map[string]any{"message": ""}, "")
		if status != 422 || code != "VALIDATION_FAILED" {
			t.Fatalf("wrong business error mapping: HTTP %d, %s", status, code)
		}
	})
	t.Run("event to task receipt and retries", func(t *testing.T) {
		status, code, response := invoke("GreetingApiService", "getGreetingReceipt",
			map[string]any{"greetingId": "123e4567-e89b-12d3-a456-426614174099"}, "")
		if status != 200 || code != "OK" || response["result"] != nil {
			t.Fatalf("missing receipt should be null: HTTP %d, %s, %v", status, code, response)
		}
		status, code, response = invoke("GreetingApiService", "createGreeting", map[string]any{"message": retryMessage}, "")
		if status != 200 || code != "OK" {
			t.Fatalf("create failed: HTTP %d, %s", status, code)
		}
		id := response["result"].(map[string]any)["id"].(string)
		deadline := time.Now().Add(20 * time.Second)
		for {
			status, code, response = invoke("GreetingApiService", "getGreetingReceipt", map[string]any{"greetingId": id}, "")
			if status != 200 || code != "OK" {
				t.Fatalf("receipt lookup failed: HTTP %d, %s", status, code)
			}
			counts := asyncDeliveries.snapshot(id)
			// Two event attempts launch two tasks; the failed first task also retries.
			if counts.Events >= 2 && counts.Tasks >= 3 && response["result"] != nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("async delivery did not recover: attempts=%+v, response=%v", counts, response)
			}
			time.Sleep(50 * time.Millisecond)
		}
		receipt := response["result"].(map[string]any)
		if receipt["jobId"] != id || receipt["greetingId"] != id || receipt["message"] != "Greeting received: "+retryMessage || receipt["preparedAt"] == nil {
			t.Fatalf("incorrect receipt: %v", receipt)
		}
		// Inspect the durable side effect after the real Event and Task redeliveries.
		db, err := sql.Open("sqlite", filepath.Join(workdir, "greeting.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM greeting_receipt_records WHERE job_uuid = ?", id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("duplicate side effects: %d receipts", count)
		}
	})
}
