package docs

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed swagger/swagger.json swagger/swagger.yaml swagger/rapidoc.html
var swaggerFiles embed.FS

func Routes() chi.Router {
	r := chi.NewRouter()

	swaggerFS, err := fs.Sub(swaggerFiles, "swagger")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(swaggerFS))

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			data, err := fs.ReadFile(swaggerFS, "rapidoc.html")
			if err != nil {
				http.Error(w, "RapiDoc UI not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
		})

		r.Get("/doc.json", func(w http.ResponseWriter, r *http.Request) {
			data, err := fs.ReadFile(swaggerFS, "swagger.json")
			if err != nil {
				http.Error(w, "Swagger spec not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write(data)
		})

		r.Handle("/*", fileServer)
	})

	return r
}
