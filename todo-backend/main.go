package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const (
	port          = ":3000"
	maxTodoLength = 140
)

type Todo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

var db *sql.DB

func initDB() error {
	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todo (
			id SERIAL PRIMARY KEY,
			todo TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return err
}

func getTodos() ([]Todo, error) {
	var todos []Todo
	rows, err := db.Query("SELECT id, todo FROM todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Text); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}

func createTodo(text string) (Todo, error) {
	if len(text) > maxTodoLength {
		return Todo{}, fmt.Errorf("todo text exceeds maximum length of %d characters", maxTodoLength)
	}
	var todo Todo
	err := db.QueryRow(`
		INSERT INTO todo (todo)
		VALUES ($1)
		RETURNING id, todo
	`, text).Scan(&todo.ID, &todo.Text)
	return todo, err
}

func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler(w, r)
		log.Printf(
			"Method: %s, Path: %s, Duration: %v",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	}
}

func main() {

	if err := initDB(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/todos", logRequest(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			todos, err := getTodos()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(todos)
		case http.MethodPost:
			var todo Todo
			if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
				log.Printf("Error decoding request: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			newTodo, err := createTodo(todo.Text)
			if err != nil {
				log.Printf("Error creating todo: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			log.Printf("Created new todo: ID=%d, Text='%s'", newTodo.ID, newTodo.Text)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newTodo)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	fmt.Printf("Server started on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
