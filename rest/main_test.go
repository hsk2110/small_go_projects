package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

var testApp App

func TestMain(m *testing.M) {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:password@localhost:5432/todos")
	if err != nil {
		log.Fatal("could not connect to test database:", err)
	}
	testApp = App{db: conn}
	os.Exit(m.Run())
}

func TestHandleTodos(t *testing.T) {
	testApp.db.Exec(context.Background(), "DELETE FROM todos")
	w := httptest.NewRecorder()

	r := httptest.NewRequest("GET", "/todo", nil)

	testApp.handleTodos(w, r)

	if w.Code != 200 {
		t.Errorf("Got: %v, want: %v", w.Code, 200)
	}

	if w.Body.String() == "" {
		t.Errorf("Got: %v, want: non empty body", w.Body.String())
	}
}

func TestHandleTodosPost(t *testing.T) {
	testApp.db.Exec(context.Background(), "DELETE FROM todos")
	body := strings.NewReader(`{"title":"Buy milk"}`)
	w := httptest.NewRecorder()

	r := httptest.NewRequest("POST", "/todo", body)

	testApp.handleTodos(w, r)

	if w.Code != 201 {
		t.Errorf("Got: %v, want: %v", w.Code, 201)
	}

	if w.Body.String() == "" {
		t.Errorf("Got: %v, want: non empty body", w.Body.String())
	}
}

func TestHandleTodosDelete(t *testing.T) {
	testApp.db.Exec(context.Background(), "DELETE FROM todos")
	body := strings.NewReader(`{"title":"Buy milk"}`)

	w := httptest.NewRecorder()

	r := httptest.NewRequest("POST", "/todo", body)

	testApp.handleTodos(w, r)

	w = httptest.NewRecorder()

	r = httptest.NewRequest("DELETE", "/todo/1", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /todo/1", testApp.handleTodosDelete)
	mux.ServeHTTP(w, r)

	if w.Code != 204 {
		t.Errorf("Got: %v, want: %v", w.Code, 204)
	}
}

func TestHandleTodosUpdate(t *testing.T) {
	testApp.db.Exec(context.Background(), "DELETE FROM todos")
	body := strings.NewReader(`{"title":"Buy milk"}`)

	w := httptest.NewRecorder()

	r := httptest.NewRequest("POST", "/todo", body)

	testApp.handleTodos(w, r)

	w = httptest.NewRecorder()

	body = strings.NewReader(`{"title":"Buy cheese"}`)

	r = httptest.NewRequest("PUT", "/todo/1", body)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /todo/1", testApp.handleTodosUpdate)
	mux.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("Got: %v, want: %v", w.Code, 200)
	}

	if !strings.Contains(w.Body.String(), `"title":"Buy cheese"`) {
		t.Errorf("Got: %v, want: %v inside", w.Body.String(), `"title":"Buy cheese"`)
	}
}
