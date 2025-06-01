package docs

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	User     string
	Password string
}

func RegisterRoutes(r chi.Router, cfg Config) {
	// Настройка Basic Auth
	docsAuth := middleware.BasicAuth("Restricted Docs", map[string]string{
		cfg.User: cfg.Password,
	})

	projectRoot := getProjectRoot()

	r.Route("/swagger", func(r chi.Router) {
		r.Use(docsAuth)

		// Swagger JSON
		r.Get("/doc.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, filepath.Join(projectRoot, "docs/swagger.json"))
		})
	})

	// Группа защищенных роутов для документации
	r.Route("/docs", func(r chi.Router) {
		r.Use(docsAuth)

		// RapiDoc UI
		staticPath := filepath.Join(projectRoot, "docs")
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticPath, "rapidoc.html"))
		})
		r.Handle("/*", http.StripPrefix("/docs", http.FileServer(http.Dir(staticPath))))
	})
}

// getProjectRoot возвращает корневую директорию проекта
func getProjectRoot() string {
	dir, _ := os.Getwd()
	return dir
}
