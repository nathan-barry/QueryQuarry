package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nathan-barry/query-quarry/search"
)

const WIKI_40B = "data/wiki40b.test"

func QueryHandler(w http.ResponseWriter, r *http.Request) {

	t := time.Now()

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error reading body")
		http.Error(w, "Error reading body", http.StatusInternalServerError)
	}

	if len(body) == 0 {
		log.Fatal("Body is empty")
		http.Error(w, "Body is empty", http.StatusBadRequest)
	}

	// Unmarshal request data
	var data RequestData
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}
	fmt.Println("Query:", data.Query)

	// Count occurrences
	count := search.CountOccurrences(WIKI_40B, data.Query)
	fmt.Println("\tCount:", count)
	fmt.Println("\tTime taken:", time.Since(t).Seconds())

	// Send result back
	response := ResponseData{Occurrences: count}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type RequestData struct {
	Length int64  `json:"length"`
	Query  string `json:"query"`
}

type ResponseData struct {
	Occurrences int64 `json:"occurrences"`
}
