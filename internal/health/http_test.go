package health

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testHealthServer() *httptest.Server {
	service, _, _ := testHealthService()
	mux := http.NewServeMux()
	NewHandler(service).Register(mux)
	return httptest.NewServer(mux)
}

func TestIngestHealthSampleHTTP(t *testing.T) {
	t.Parallel()

	server := testHealthServer()
	defer server.Close()

	body, _ := json.Marshal(IngestSampleRequest{
		ProviderID: "ria",
		RouteID:    "eu-ng-account",
		SignalType: SignalErrorRate,
		ErrorRate:  &RateSignal{RateBps: uint16Ptr(650)},
	})
	response, err := http.Post(server.URL+"/v1/health/samples", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.StatusCode)
	}

	var payload IngestSampleResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Sample.State != StateWatch {
		t.Fatalf("expected watch state, got %s", payload.Sample.State)
	}
	if payload.StateChange == nil {
		t.Fatal("expected state change")
	}
}

func TestGetRouteHealthHTTP(t *testing.T) {
	t.Parallel()

	server := testHealthServer()
	defer server.Close()

	body, _ := json.Marshal(IngestSampleRequest{
		ProviderID:  "remitly",
		RouteID:     "uk-ng-account",
		SignalType:  SignalCallbackLag,
		CallbackLag: &CallbackLagSignal{P95LagMS: 90_000},
	})
	response, err := http.Post(server.URL+"/v1/health/samples", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	response.Body.Close()

	getResponse, err := http.Get(server.URL + "/v1/health/routes/uk-ng-account")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer getResponse.Body.Close()

	if getResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", getResponse.StatusCode)
	}

	var payload RouteHealthResponse
	if err := json.NewDecoder(getResponse.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.State != StateWatch {
		t.Fatalf("expected watch state, got %s", payload.State)
	}
	if payload.Snapshot.P95LatencyMS != 90_000 {
		t.Fatalf("expected p95 latency 90000, got %d", payload.Snapshot.P95LatencyMS)
	}
}

func TestIngestHealthSampleHTTPValidationError(t *testing.T) {
	t.Parallel()

	server := testHealthServer()
	defer server.Close()

	body, _ := json.Marshal(IngestSampleRequest{
		ProviderID: "ria",
		SignalType: SignalErrorRate,
		ErrorRate:  &RateSignal{RateBps: uint16Ptr(11_000)},
	})
	response, err := http.Post(server.URL+"/v1/health/samples", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.StatusCode)
	}
}
