package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStatusHealthy(t *testing.T) {
	s := New()
	s.SetBroker(true)
	s.RecordPoll(nil)

	if !s.Healthy(time.Minute) {
		t.Fatal("expected healthy")
	}

	s.SetBroker(false)
	if s.Healthy(time.Minute) {
		t.Fatal("expected unhealthy when broker down")
	}
}

func TestHealthzHandler(t *testing.T) {
	s := New()
	s.SetBroker(true)
	s.RecordPoll(nil)

	srv := NewServer(":0", s)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["broker_up"] != true {
		t.Fatalf("expected broker_up=true, got %v", body["broker_up"])
	}
}
