package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleTodos(t *testing.T) {
	todos = []Todo{}
	w := httptest.NewRecorder()

	r := httptest.NewRequest("GET", "/todo", nil)

	handleTodos(w, r)

	if w.Code != 200 {
		t.Errorf("Got: %v, want: %v", w.Code, 200)
	}

	if w.Body.String() == "" {
		t.Errorf("Got: %v, want: non empty body", w.Body.String())
	}
}

func TestHandleTodosPost(t *testing.T) {
	todos = []Todo{}
	body := strings.NewReader(`{"title":"Buy milk"}`)

	w := httptest.NewRecorder()

	r := httptest.NewRequest("POST", "/todo", body)

	handleTodos(w, r)

	if w.Code != 201 {
		t.Errorf("Got: %v, want: %v", w.Code, 201)
	}

	if w.Body.String() == "" {
		t.Errorf("Got: %v, want: non empty body", w.Body.String())
	}
}

func TestHandleTodosDelete(t *testing.T) {
	todos = []Todo{}
	id = 0
	body := strings.NewReader(`{"title":"Buy milk"}`)

	w := httptest.NewRecorder()

	r := httptest.NewRequest("POST", "/todo", body)

	handleTodos(w, r)

	w = httptest.NewRecorder()

	r = httptest.NewRequest("DELETE", "/todo/0", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /todo/{id}", handleTodosDelete)
	mux.ServeHTTP(w, r)

	if w.Code != 204 {
		t.Errorf("Got: %v, want: %v", w.Code, 204)
	}
}

func TestHandleTodosUpdate(t *testing.T) {
	todos = []Todo{}
	id = 0
	body := strings.NewReader(`{"title":"Buy milk"}`)

	w := httptest.NewRecorder()

	r := httptest.NewRequest("POST", "/todo", body)

	handleTodos(w, r)

	w = httptest.NewRecorder()

	body = strings.NewReader(`{"title":"Buy cheese"}`)

	r = httptest.NewRequest("PUT", "/todo/0", body)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /todo/{id}", handleTodosUpdate)
	mux.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("Got: %v, want: %v", w.Code, 200)
	}

	if !strings.Contains(w.Body.String(), `"title":"Buy cheese"`) {
		t.Errorf("Got: %v, want: %v inside", w.Body.String(), `"title":"Buy cheese"`)
	}
}
