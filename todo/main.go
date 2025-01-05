package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const port = ":3000"
const cacheDir = "/app/cache"
const imagePath = "/app/cache/current.jpg"

func downloadImage() error {
	resp, err := http.Get("https://picsum.photos/1200")
	if (err != nil) {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func ensureCache() error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return downloadImage()
	}

	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		return err
	}

	if time.Since(fileInfo.ModTime()) > 60*time.Minute {
		return downloadImage()
	}

	return nil
}

func main() {
	if err := ensureCache(); err != nil {
		log.Printf("Initial cache setup failed: %v", err)
	}

	go func() {
		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := downloadImage(); err != nil {
				log.Printf("Failed to update image: %v", err)
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
				<head>
					<title>Welcome</title>
					<style>
						.todo-container {
							margin: 20px 0;
						}
						.todo-item {
							margin: 5px 0;
							padding: 5px;
							border-bottom: 1px solid #ccc;
						}
					</style>
				</head>
				<body>
					<h1>Welcome to the Todo Server</h1>
					<p>This is a simple web server built with Go.</p>
					<img src="/image" alt="Random Picture" style="max-width: 100%;">
					
					<div class="todo-container">
						 <form id="todoForm" onsubmit="return submitTodo(event)">
                            <input type="text" id="todoInput" maxlength="140" placeholder="Enter your todo (max 140 chars)" required>
                            <button type="submit">Create Todo</button>
                        </form>
                        
                        <h2>Existing Todos:</h2>
                        <div id="todoList" class="todo-list"></div>
                    </div>

                    <script>
                        function loadTodos() {
                            fetch('/todos')
                                .then(response => response.json())
                                .then(todos => {
                                    const todoList = document.getElementById('todoList');
                                    todoList.innerHTML = todos.map(todo => 
                                        '<div class="todo-item">' + todo.text + '</div>'
                                    ).join('');
                                })
                                .catch(error => console.error('Error loading todos:', error));
                        }

                        function submitTodo(event) {
                            event.preventDefault();
                            const input = document.getElementById('todoInput');
                            const text = input.value.trim();
                            
                            if (text) {
                                fetch('/todos', {
                                    method: 'POST',
                                    headers: {'Content-Type': 'application/json'},
                                    body: JSON.stringify({text})
                                })
                                .then(() => {
                                    input.value = '';
                                    loadTodos();
                                })
                                .catch(error => console.error('Error creating todo:', error));
                            }
                            return false;
                        }

                        // Load todos when page loads
                        loadTodos();
                    </script>
				</body>
			</html>
		`)
	})

	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		if err := ensureCache(); err != nil {
			http.Error(w, "Failed to ensure cache", http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, imagePath)
	})

	fmt.Printf("Server started on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
