package app

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Start the real two-App bundle once; all calls enter through the local Portal.
func TestDomains(t *testing.T) {
	seed, err := os.ReadFile("../seed/hub.yaml")
	if err != nil {
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
	seedPath := filepath.Join(t.TempDir(), "hub.yaml")
	if err := os.WriteFile(seedPath, []byte(strings.ReplaceAll(string(seed), "7388", fmt.Sprint(port))), 0600); err != nil {
		t.Fatal(err)
	}
	runtime := New(seedPath, "")
	runtime.Start()
	t.Cleanup(runtime.StopGracefully)
	client := &http.Client{Timeout: 5 * time.Second}
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/api/invoke/shop.orders.OrderApiService/", port)

	invoke := func(t *testing.T, method string, params any) (int, string, map[string]any) {
		t.Helper()
		body, err := json.Marshal(map[string]any{"params": params})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, endpoint+method, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Accept", "application/vrpc+json")
		req.Header.Set("Content-Type", "application/vrpc+json")
		req.Header.Set("Vrpc-Trace", "id=123e4567e89b12d3a456426614174000")
		req.Header.Set("Vrpc-Client", "name=domains.test,version=1.0.0,instanceId=123e4567-e89b-12d3-a456-426614174001")
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var envelope map[string]any
		if err := json.UnmarshalRead(response.Body, &envelope); err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, response.Header.Get("Vrpc-Status"), envelope
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		status, code, _ := invoke(t, "getOrder", map[string]any{"orderId": "123e4567-e89b-12d3-a456-426614174099"})
		if status == 404 && code == "NOT_FOUND" {
			break
		}
		if status != 503 || time.Now().After(deadline) {
			t.Fatalf("orders API did not become ready: HTTP %d, %s", status, code)
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Run("two-way Rpc and events", func(t *testing.T) {
		for _, quantity := range []int{2, 3} {
			t.Run(fmt.Sprintf("quantity=%d", quantity), func(t *testing.T) {
				t.Parallel()
				status, code, response := invoke(t, "placeOrder", map[string]any{"sku": "notebook", "quantity": quantity})
				if status != 200 || code != "OK" {
					t.Fatalf("place order failed: HTTP %d, %s, %v", status, code, response)
				}
				order := response["result"].(map[string]any)
				product := order["product"].(map[string]any)
				if product["sku"] != "notebook" || product["title"] != "Vine notebook" || product["unitPriceCents"] != float64(1290) {
					t.Fatalf("catalog Rpc or imported DTO mismatch: %v", product)
				}
				id := order["id"].(string)
				deadline := time.Now().Add(10 * time.Second)
				for {
					status, code, response = invoke(t, "getOrder", map[string]any{"orderId": id})
					if status != 200 || code != "OK" {
						t.Fatalf("get order failed: HTTP %d, %s", status, code)
					}
					order = response["result"].(map[string]any)
					if order["confirmed"] == true {
						break
					}
					if time.Now().After(deadline) {
						t.Fatalf("two-domain event round trip did not complete: %v", order)
					}
					time.Sleep(50 * time.Millisecond)
				}
				// OrderPlaced carries IDs only. This text requires the catalog's
				// reverse Rpc to OrderService, then its event back to the orders App.
				if order["id"] != id || order["quantity"] != float64(quantity) || order["confirmation"] != fmt.Sprintf("Reserved %d × Vine notebook", quantity) {
					t.Fatalf("reverse Rpc or confirmation event mismatch: %v", order)
				}
			})
		}
	})
	t.Run("cross-App error propagation", func(t *testing.T) {
		status, code, _ := invoke(t, "placeOrder", map[string]any{"sku": "missing", "quantity": 1})
		if status != 404 || code != "NOT_FOUND" {
			t.Fatalf("catalog error lost across orders and Portal: HTTP %d, %s", status, code)
		}
		status, code, _ = invoke(t, "placeOrder", map[string]any{"sku": "notebook", "quantity": 0})
		if status != 422 || code != "VALIDATION_FAILED" {
			t.Fatalf("quantity validation failed: HTTP %d, %s", status, code)
		}
	})
}
