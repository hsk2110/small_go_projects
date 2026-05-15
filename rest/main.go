package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"
	"time"
)

// the todo struct consists of id, title, whether it's finished and description
type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Finished    bool   `json:"finished"`
	Description string `json:"description"`
}

var todos = []Todo{}
var id int

// this handler is responsible for the GET and POST methods.
// if it's GET then we display all the todos
// if it's POST then we append the new todo and display them all
func handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(todos)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		todo := &Todo{}
		err := json.NewDecoder(r.Body).Decode(todo)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		todo.ID = id
		w.Header().Set("Content-Type", "application/json")
		id++
		todos = append(todos, *todo)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(todo)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

// for DELETE requests
func handleTodosDelete(w http.ResponseWriter, r *http.Request) {
	requestedID, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	index := slices.IndexFunc(todos, func(t Todo) bool {
		return t.ID == requestedID
	})
	// if the index doesnt exist
	if index < 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else {
		todos = slices.Delete(todos, index, index+1)
		w.WriteHeader(http.StatusNoContent)
	}

}

func handleTodosUpdate(w http.ResponseWriter, r *http.Request) {

	requestedID, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	index := slices.IndexFunc(todos, func(t Todo) bool {
		return t.ID == requestedID
	})

	if index < 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else {
		todo := &Todo{}

		err = json.NewDecoder(r.Body).Decode(todo)

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusBadRequest)
			return
		}
		todo.ID = requestedID
		todos[index] = *todo
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(todo)
	}
}

func myMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Method: ", r.Method, "\nPath: ", r.URL.Path)
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		fmt.Println("Duration: ", duration)
	})
}

func main() {

	mux := http.NewServeMux()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = fmt.Sprintf(":%s", port)

	mux.HandleFunc("/todo", handleTodos)
	mux.HandleFunc("DELETE /todo/{id}", handleTodosDelete)
	mux.HandleFunc("PUT /todo/{id}", handleTodosUpdate)

	s := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()

	ctx, stop = context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	s.Shutdown(ctx)
}
