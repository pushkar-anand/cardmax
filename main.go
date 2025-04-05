package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var templates *template.Template

func init() {
	templatesDir := "./templates"
	templates = template.Must(template.ParseGlob(filepath.Join(templatesDir, "*.html")))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "index.html", nil)
}

func cardsHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "cards.html", nil)
}

func recommendHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "recommend.html", nil)
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "transactions.html", nil)
}

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route handlers
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/cards", cardsHandler)
	http.HandleFunc("/recommend", recommendHandler)
	http.HandleFunc("/transactions", transactionsHandler)

	// Start server
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
