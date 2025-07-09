package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var tasks []Task
var nextID = 1

func main() {
	// Khởi tạo sample data
	initSampleData()

	// Tạo router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", healthCheck).Methods("GET")
	api.HandleFunc("/tasks", getTasks).Methods("GET")
	api.HandleFunc("/tasks", createTask).Methods("POST")
	api.HandleFunc("/tasks/{id}", getTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")

	// Root route
	r.HandleFunc("/", homeHandler).Methods("GET")

	// CORS middleware
	r.Use(corsMiddleware)

	// Lấy port từ environment variable (Railway sẽ set PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Printf("📡 API available at: http://localhost:%s/api\n", port)
	fmt.Printf("🏠 Home page: http://localhost:%s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initSampleData() {
	tasks = []Task{
		{
			ID:          1,
			Title:       "Learn Go",
			Description: "Study Go programming language",
			Completed:   false,
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			Title:       "Deploy to Railway",
			Description: "Deploy Go API to Railway platform",
			Completed:   false,
			CreatedAt:   time.Now(),
		},
	}
	nextID = 3
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func sendResponse(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	response := APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	
	json.NewEncoder(w).Encode(response)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Go Task API</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
			.endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
			.method { display: inline-block; padding: 2px 8px; border-radius: 3px; font-weight: bold; }
			.get { background: #4CAF50; color: white; }
			.post { background: #2196F3; color: white; }
			.put { background: #FF9800; color: white; }
			.delete { background: #f44336; color: white; }
		</style>
	</head>
	<body>
		<h1>🚀 Go Task API</h1>
		<p>Simple REST API for managing tasks</p>
		
		<h2>Available Endpoints:</h2>
		
		<div class="endpoint">
			<span class="method get">GET</span> <code>/api/health</code> - Health check
		</div>
		
		<div class="endpoint">
			<span class="method get">GET</span> <code>/api/tasks</code> - Get all tasks
		</div>
		
		<div class="endpoint">
			<span class="method post">POST</span> <code>/api/tasks</code> - Create new task
		</div>
		
		<div class="endpoint">
			<span class="method get">GET</span> <code>/api/tasks/{id}</code> - Get task by ID
		</div>
		
		<div class="endpoint">
			<span class="method put">PUT</span> <code>/api/tasks/{id}</code> - Update task
		</div>
		
		<div class="endpoint">
			<span class="method delete">DELETE</span> <code>/api/tasks/{id}</code> - Delete task
		</div>
		
		<h2>Example Usage:</h2>
		<pre>
# Get all tasks
curl -X GET /api/tasks

# Create a new task
curl -X POST /api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "New Task", "description": "Task description"}'

# Update a task
curl -X PUT /api/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Task", "completed": true}'
		</pre>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, http.StatusOK, true, "API is running", map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, http.StatusOK, true, "Tasks retrieved successfully", tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendResponse(w, http.StatusBadRequest, false, "Invalid JSON format", nil)
		return
	}
	
	if task.Title == "" {
		sendResponse(w, http.StatusBadRequest, false, "Title is required", nil)
		return
	}
	
	task.ID = nextID
	nextID++
	task.CreatedAt = time.Now()
	task.Completed = false
	
	tasks = append(tasks, task)
	
	sendResponse(w, http.StatusCreated, true, "Task created successfully", task)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendResponse(w, http.StatusBadRequest, false, "Invalid task ID", nil)
		return
	}
	
	for _, task := range tasks {
		if task.ID == id {
			sendResponse(w, http.StatusOK, true, "Task found", task)
			return
		}
	}
	
	sendResponse(w, http.StatusNotFound, false, "Task not found", nil)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendResponse(w, http.StatusBadRequest, false, "Invalid task ID", nil)
		return
	}
	
	var updatedTask Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		sendResponse(w, http.StatusBadRequest, false, "Invalid JSON format", nil)
		return
	}
	
	for i, task := range tasks {
		if task.ID == id {
			if updatedTask.Title != "" {
				tasks[i].Title = updatedTask.Title
			}
			if updatedTask.Description != "" {
				tasks[i].Description = updatedTask.Description
			}
			tasks[i].Completed = updatedTask.Completed
			
			sendResponse(w, http.StatusOK, true, "Task updated successfully", tasks[i])
			return
		}
	}
	
	sendResponse(w, http.StatusNotFound, false, "Task not found", nil)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendResponse(w, http.StatusBadRequest, false, "Invalid task ID", nil)
		return
	}
	
	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			sendResponse(w, http.StatusOK, true, "Task deleted successfully", nil)
			return
		}
	}
	
	sendResponse(w, http.StatusNotFound, false, "Task not found", nil)
}