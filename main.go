package main

import (
	"log"
	"net/http"

	"github.com/nathan-barry/query-quarry/handlers"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("POST /query", handlers.QueryHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
