package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bannerfingerprint/internal/engine"
	"bannerfingerprint/internal/model"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	e, err := engine.New()
	if err != nil {
		t.Fatal(err)
	}
	return NewHandler(e)
}

func TestBatchEndpoint(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t))
	defer srv.Close()

	payload, _ := json.Marshal([]model.Input{{IP: "1.2.3.4", Port: 22, Banner: "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"}})

	resp, err := http.Post(srv.URL+"/fingerprint", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var got []model.Result
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Protocol != "SSH" || got[0].Product != "OpenSSH" || got[0].Version != "8.9p1" || got[0].OSHint != "Ubuntu" {
		t.Errorf("unexpected result: %+v", got[0])
	}
}

func TestBatchEndpointInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t))
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/fingerprint", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestBatchEndpointMethodNotAllowed(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
}

func TestSingleEndpoint(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t))
	defer srv.Close()

	payload, _ := json.Marshal(model.Input{IP: "1.2.3.7", Port: 3306, Banner: "J\\x00\\x00\\x00\n8.0.32\\x00"})

	resp, err := http.Post(srv.URL+"/fingerprint/single", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got model.Result
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Protocol != "MySQL" || got.Product != "MySQL" || got.Version != "8.0.32" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(newTestHandler(t))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
