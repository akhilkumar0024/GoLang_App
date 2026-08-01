package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRoot(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := setupRoutes()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Welcome to Go Application") {
		t.Errorf("handler returned unexpected body")
	}
}

func TestHandleHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := setupRoutes()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var res Response
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if res.Status != "UP" {
		t.Errorf("expected status 'UP', got '%s'", res.Status)
	}
}

func TestHandleTasks(t *testing.T) {
	handler := setupRoutes()

	// Test GET tasks
	reqGet, _ := http.NewRequest("GET", "/api/tasks", nil)
	rrGet := httptest.NewRecorder()
	handler.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Errorf("GET /api/tasks returned status %v", rrGet.Code)
	}

	// Test POST new task
	payload := []byte(`{"title": "Test new task"}`)
	reqPost, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(payload))
	reqPost.Header.Set("Content-Type", "application/json")
	rrPost := httptest.NewRecorder()
	handler.ServeHTTP(rrPost, reqPost)

	if rrPost.Code != http.StatusCreated {
		t.Errorf("POST /api/tasks returned status %v, want %v", rrPost.Code, http.StatusCreated)
	}

	var createdTask Task
	if err := json.Unmarshal(rrPost.Body.Bytes(), &createdTask); err != nil {
		t.Fatalf("failed to decode created task: %v", err)
	}

	if createdTask.Title != "Test new task" {
		t.Errorf("expected title 'Test new task', got '%s'", createdTask.Title)
	}
}
