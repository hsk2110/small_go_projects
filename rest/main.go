package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
)

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Finished    bool   `json:"finished"`
	Description string `json:"description"`
}

var todos = []Todo{}
var id int

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

func handleTodosDelete(w http.ResponseWriter, r *http.Request) {
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

func main() {
	http.HandleFunc("/todo", handleTodos)
	http.HandleFunc("DELETE /todo/{id}", handleTodosDelete)
	http.HandleFunc("PUT /todo/{id}", handleTodosUpdate)
	log.Fatal(http.ListenAndServe(":8888", nil))

}
