package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	
	"./models"
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
	
	// API endpoints
	http.HandleFunc("/api/predefined-cards", predefinedCardsHandler)

	// Start server
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// API handler for predefined cards
func predefinedCardsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check if this is a request for a specific card
	if r.URL.Path != "/api/predefined-cards" {
		// Extract the card ID from the path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) == 4 {
			idStr := pathParts[3]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid card ID", http.StatusBadRequest)
				return
			}
			
			// Get the specific card
			card, err := models.GetPredefinedCardByID(id)
			if err != nil {
				http.Error(w, "Card not found", http.StatusNotFound)
				return
			}
			
			json.NewEncoder(w).Encode(card)
			return
		}
		
		http.NotFound(w, r)
		return
	}
	
	// Return all cards
	cards, err := models.LoadPredefinedCards()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(cards)
}
