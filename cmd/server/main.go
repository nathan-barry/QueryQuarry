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

	// API endpoint
	http.HandleFunc("POST /query", handlers.QueryHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
