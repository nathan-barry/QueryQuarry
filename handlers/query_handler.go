package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nathan-barry/QueryQuarry/search"
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

	// Open files
	textFile, err := os.Open(WIKI_40B)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	saFile, err := os.Open(WIKI_40B + ".table.bin")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer textFile.Close()
	defer saFile.Close()

	// Count occurrences
	firstSAIndex, lastSAIndex := search.CountOccurrences(textFile, saFile, data.Query)

	var count int64
	var sentences []string
	if firstSAIndex < 0 || lastSAIndex < 0 { // Both -1 if no occurrences
		count = 0
	} else {
		count = lastSAIndex - firstSAIndex + 1
		sentences = search.NearbyWords(textFile, saFile, firstSAIndex, lastSAIndex)
	}
	log.Printf("QUERY: \"%v\", COUNT: %v, TIME_TAKEN: %v", data.Query, count, time.Since(t).Seconds())

	// Send result back
	response := ResponseData{Occurrences: count, Sentences: sentences}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type RequestData struct {
	Length int64  `json:"length"`
	Query  string `json:"query"`
}

type ResponseData struct {
	Occurrences int64    `json:"occurrences"`
	Sentences   []string `json:"sentences"`
}
