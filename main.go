package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Response struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type InfoResponse struct {
	AppName   string    `json:"app_name"`
	Version   string    `json:"version"`
	GoVersion string    `json:"go_version"`
	Uptime    string    `json:"uptime"`
	Timestamp time.Time `json:"timestamp"`
}

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskStore struct {
	sync.RWMutex
	tasks  []Task
	nextID int
}

var (
	startTime = time.Now()
	store     = &TaskStore{
		tasks: []Task{
			{ID: 1, Title: "Learn Go programming", Completed: true, CreatedAt: time.Now().Add(-24 * time.Hour)},
			{ID: 2, Title: "Build RESTful API in Go", Completed: false, CreatedAt: time.Now().Add(-12 * time.Hour)},
			{ID: 3, Title: "Containerize Go App with Docker", Completed: false, CreatedAt: time.Now()},
		},
		nextID: 4,
	}
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Application</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; }
        body { background: #0f172a; color: #f8fafc; display: flex; justify-content: center; align-items: center; min-height: 100vh; padding: 20px; }
        .card { background: #1e293b; border-radius: 16px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); padding: 40px; max-width: 600px; width: 100%; border: 1px solid #334155; text-align: center; }
        h1 { color: #38bdf8; font-size: 2.2rem; margin-bottom: 12px; }
        p { color: #94a3b8; font-size: 1.05rem; margin-bottom: 24px; line-height: 1.6; }
        .badge { display: inline-block; background: #0284c7; color: #fff; padding: 6px 16px; border-radius: 20px; font-weight: 600; font-size: 0.9rem; margin-bottom: 24px; }
        .endpoints { text-align: left; background: #0f172a; border-radius: 10px; padding: 20px; border: 1px solid #334155; }
        .endpoints h3 { color: #cbd5e1; font-size: 1.1rem; margin-bottom: 12px; }
        .endpoint-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid #1e293b; }
        .endpoint-item:last-child { border-bottom: none; }
        .method { font-weight: bold; font-size: 0.85rem; padding: 3px 8px; border-radius: 4px; }
        .get { background: #065f46; color: #34d399; }
        .post { background: #1e40af; color: #60a5fa; }
        a { color: #38bdf8; text-decoration: none; font-family: monospace; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="card">
        <div class="badge">🚀 GoLang HTTP Server</div>
        <h1>Welcome to Go Application</h1>
        <p>A fast, lightweight, and containerized Go web server application.</p>
        
        <div class="endpoints">
            <h3>Available Endpoints:</h3>
            <div class="endpoint-item">
                <span class="method get">GET</span>
                <a href="/api/health" target="_blank">/api/health</a>
            </div>
            <div class="endpoint-item">
                <span class="method get">GET</span>
                <a href="/api/info" target="_blank">/api/info</a>
            </div>
            <div class="endpoint-item">
                <span class="method get">GET</span>
                <a href="/api/tasks" target="_blank">/api/tasks</a>
            </div>
            <div class="endpoint-item">
                <span class="method post">POST</span>
                <span>/api/tasks</span>
            </div>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Status:    "UP",
		Message:   "Service is healthy and running",
		Timestamp: time.Now(),
	})
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(InfoResponse{
		AppName:   "GoLang_App",
		Version:   "1.0.0",
		GoVersion: "go1.21",
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		Timestamp: time.Now(),
	})
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		store.RLock()
		defer store.RUnlock()
		json.NewEncoder(w).Encode(store.tasks)

	case http.MethodPost:
		var input struct {
			Title string `json:"title"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
			http.Error(w, `{"error": "Invalid request payload. 'title' is required."}`, http.StatusBadRequest)
			return
		}

		store.Lock()
		newTask := Task{
			ID:        store.nextID,
			Title:     input.Title,
			Completed: false,
			CreatedAt: time.Now(),
		}
		store.nextID++
		store.tasks = append(store.tasks, newTask)
		store.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newTask)

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/info", handleInfo)
	mux.HandleFunc("/api/tasks", handleTasks)

	return loggingMiddleware(mux)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      setupRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on http://localhost%s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", addr, err)
		}
	}()

	<-stop
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
