package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nathan-barry/QueryQuarry/handlers"
)

func main() {
	fmt.Println("Starting Server...")

	// Serve home page static HTML file
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Custom handle for about page
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/about.html")
	})

	// API endpoints
	http.HandleFunc("POST /count", handlers.CountHandler)
	http.HandleFunc("POST /csv", handlers.CSVHandler)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
