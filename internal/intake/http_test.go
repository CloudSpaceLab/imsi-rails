package intake

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer() (*httptest.Server, *InMemoryEventSink) {
	service, _, events, _ := testService()
	mux := http.NewServeMux()
	NewHandler(service).Register(mux)
	return httptest.NewServer(mux), events
}

func TestSubmitTransactionHTTP(t *testing.T) {
	t.Parallel()

	server, events := testServer()
	defer server.Close()

	body, _ := json.Marshal(validRequest())
	response, err := http.Post(server.URL+"/v1/transactions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.StatusCode)
	}

	var payload SubmitTransactionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.TransactionID == "" {
		t.Fatal("expected transaction id")
	}
	if payload.SelectedRouteID != "r1" {
		t.Fatalf("expected r1, got %s", payload.SelectedRouteID)
	}
	if len(events.Events()) != 2 {
		t.Fatalf("expected 2 emitted events, got %d", len(events.Events()))
	}
}

func TestSubmitTransactionHTTPIdempotentReplay(t *testing.T) {
	t.Parallel()

	server, _ := testServer()
	defer server.Close()

	body, _ := json.Marshal(validRequest())
	first, err := http.Post(server.URL+"/v1/transactions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("first post failed: %v", err)
	}
	defer first.Body.Close()

	second, err := http.Post(server.URL+"/v1/transactions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("second post failed: %v", err)
	}
	defer second.Body.Close()

	if second.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 replay, got %d", second.StatusCode)
	}

	var payload SubmitTransactionResponse
	if err := json.NewDecoder(second.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !payload.IDempotentReplay {
		t.Fatal("expected idempotent replay")
	}
}

func TestSubmitTransactionHTTPValidationError(t *testing.T) {
	t.Parallel()

	server, _ := testServer()
	defer server.Close()

	request := validRequest()
	request.IDempotencyKey = ""
	body, _ := json.Marshal(request)
	response, err := http.Post(server.URL+"/v1/transactions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.StatusCode)
	}
}
