package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/dale-tomson/dogs-api/db"
	"github.com/dale-tomson/dogs-api/handlers"
	appMiddleware "github.com/dale-tomson/dogs-api/middleware"
)

func getContentType(path string) string {
	switch {
	case len(path) > 5 && path[len(path)-5:] == ".html":
		return "text/html; charset=utf-8"
	case len(path) > 3 && path[len(path)-3:] == ".js":
		return "application/javascript"
	case len(path) > 4 && path[len(path)-4:] == ".css":
		return "text/css"
	case len(path) > 4 && path[len(path)-4:] == ".json":
		return "application/json"
	case len(path) > 4 && path[len(path)-4:] == ".svg":
		return "image/svg+xml"
	case len(path) > 4 && path[len(path)-4:] == ".png":
		return "image/png"
	case (len(path) > 4 && path[len(path)-4:] == ".jpg") || (len(path) > 5 && path[len(path)-5:] == ".jpeg"):
		return "image/jpeg"
	case len(path) > 4 && path[len(path)-4:] == ".gif":
		return "image/gif"
	case len(path) > 4 && path[len(path)-4:] == ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./dogs.db"
	}

	db.Init(dbPath, "dogs.json")

	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "X-API-Key"},
	}))

	r.Get("/api/health", handlers.HealthCheck)
	r.Route("/api/dogs", func(r chi.Router) {
		r.Use(httprate.LimitByIP(100, time.Minute)) // 100 req/min per IP
		r.Use(appMiddleware.APIKeyAuth)
		r.Get("/", handlers.ListDogs)
		r.Post("/", handlers.CreateDog)
		r.Get("/{breed}", handlers.GetDog)
		r.Put("/{breed}", handlers.UpdateDog)
		r.Delete("/{breed}", handlers.DeleteDog)
	})

	// Determine frontend path (local machine or Docker)
	var frontendPath string
	if _, err := os.Stat("/app/frontend/dist"); err == nil {
		frontendPath = "/app/frontend/dist"
	} else {
		frontendPath = "frontend/dist"
	}
	slog.Info("serving frontend", "path", frontendPath)

	// SPA handler - serve frontend files from filesystem
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		if path == "" {
			path = "index.html"
		}

		filePath := filepath.Join(frontendPath, path)
		content, err := os.ReadFile(filePath)
		if err != nil {
			slog.Warn("file not found, serving index.html", "requested_path", path, "error", err)
			content, err = os.ReadFile(filepath.Join(frontendPath, "index.html"))
			if err != nil {
				slog.Error("failed to read index.html", "error", err)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			path = "index.html"
		}

		w.Header().Set("Content-Type", getContentType(path))
		w.Write(content)
	}))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("shutting down — waiting for in-flight requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}
	slog.Info("server stopped cleanly")
}
