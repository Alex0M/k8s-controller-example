package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type FrontendPageData struct {
	Image    string `json:"image"`
	Replicas int    `json:"replicas"`
	Content  string `json:"content"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome to frontendpage API"))
	})
	r.Get("/api/frontendpage", getFrontendPage)
	http.ListenAndServe(":3000", r)
}

func getFrontendPage(w http.ResponseWriter, r *http.Request) {
	fe := FrontendPageData{
		Image:    "nginx",
		Replicas: 3,
		Content:  "Content from API server",
	}
	respondwithJSON(w, http.StatusOK, fe)
}

func respondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
