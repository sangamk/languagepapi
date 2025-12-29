package main

import (
	"log"
	"net/http"

	"languagepapi/components"
	"languagepapi/internal/db"
)

func main() {
	// Initialize database
	if err := db.Init("languagepapi.db"); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Routes
	http.HandleFunc("GET /", handleHome)
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	components.Home().Render(r.Context(), w)
}
