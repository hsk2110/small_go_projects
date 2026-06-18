package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

// the todo struct consists of id, title, whether it's finished and description
type Todo struct {
	ID          int    `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Finished    bool   `json:"finished" db:"finished"`
	Description string `json:"description" db:"description"`
}

type App struct {
	db *pgx.Conn
}

type User struct {
	Username string `json:"username" db:"username"`
	Password string `json:"pwd" db:"pwd"`
}

var todos = []Todo{}

var myUser = User{
	Username: "admin",
	Password: "password",
}

// this handler is responsible for the GET and POST methods.
// if it's GET then we display all the todos
// if it's POST then we append the new todo and display them all
func (a *App) handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		rows, err := a.db.Query(r.Context(), "SELECT id, title, finished, description FROM todos")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[Todo])

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(todos)

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
		w.Header().Set("Content-Type", "application/json")

		err = a.db.QueryRow(r.Context(), "INSERT INTO todos (title, finished, description) VALUES ($1, $2, $3) RETURNING id", todo.Title, todo.Finished, todo.Description).Scan(&todo.ID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(todo)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

// for DELETE requests
func (a *App) handleTodosDelete(w http.ResponseWriter, r *http.Request) {
	requestedID, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	/* index := slices.IndexFunc(todos, func(t Todo) bool {
		return t.ID == requestedID
	}) */
	commandTag, err := a.db.Exec(r.Context(), "DELETE FROM todos WHERE id = $1", requestedID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// if the index doesnt exist
	if commandTag.RowsAffected() != 1 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

}

func (a *App) handleTodosUpdate(w http.ResponseWriter, r *http.Request) {

	requestedID, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	/* index := slices.IndexFunc(todos, func(t Todo) bool {
		return t.ID == requestedID
	}) */
	todo := &Todo{}

	err = json.NewDecoder(r.Body).Decode(todo)

	todo.ID = requestedID

	if err != nil {
		http.Error(w, "invalid id", http.StatusInternalServerError)
		return
	}

	commandTag, err := a.db.Exec(r.Context(),
		"UPDATE todos SET title=$1, finished=$2, description=$3 WHERE id=$4",
		todo.Title, todo.Finished, todo.Description, todo.ID)

	if err != nil {
		http.Error(w, "invalid id", http.StatusInternalServerError)
		return
	}

	if commandTag.RowsAffected() != 1 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	} else {

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(todo)
	}
}

func (a *App) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	user := &User{}
	err := json.NewDecoder(r.Body).Decode(user)

	if err != nil {
		http.Error(w, "error decoding user", http.StatusInternalServerError)
		return
	}

	if user.Password != myUser.Password || user.Username != myUser.Username {
		http.Error(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "admin",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	})

	tokenstring, err := token.SignedString([]byte("my-secret-key"))

	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenstring})

}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt))
		if err != nil {
			log.Fatal(err)
		}

		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		slog.Info("request info: ", "method", r.Method, "path", r.URL.Path, "requestID", requestID, "duration", duration)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenstring := r.Header.Get("Authorization")
		if tokenstring == "" {
			http.Error(w, "No Authorization", http.StatusUnauthorized)
			return
		}

		tokenstring = strings.Replace(tokenstring, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
			return []byte("my-secret-key"), nil
		})

		if err != nil {
			http.Error(w, "No Authorization", http.StatusUnauthorized)
			return
		}

		if _, ok := token.Claims.(jwt.MapClaims); ok {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "invalid token", http.StatusUnauthorized)
		}

	})
}

func main() {

	mux := http.NewServeMux()
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/todos"
	}
	port = fmt.Sprintf(":%s", port)

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	slog.Info("Database connection success!")

	app := App{conn}

	mux.Handle("/todo", authMiddleware(http.HandlerFunc(app.handleTodos)))
	mux.Handle("DELETE /todo/{id}", authMiddleware(http.HandlerFunc(app.handleTodosDelete)))
	mux.Handle("PUT /todo/{id}", authMiddleware(http.HandlerFunc(app.handleTodosUpdate)))
	mux.HandleFunc("GET /health", app.handleHealthCheck)
	mux.HandleFunc("POST /login", app.handleLogin)

	s := &http.Server{
		Addr:    port,
		Handler: loggingMiddleware(mux),
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
