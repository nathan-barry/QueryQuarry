package handlers

import (
	"encoding/json"
	"fmt"
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
	fmt.Println("Query:", data.Query)

	// Opened in handler since some will use reuse files in multiple functions
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
	count := lastSAIndex - firstSAIndex + 1
	if firstSAIndex < 0 || lastSAIndex < 0 { // Both -1 if no occurrences
		count = 0
	}
	fmt.Println("\tCount:", count)
	fmt.Println("\tTime taken:", time.Since(t).Seconds())

	// Send result back
	response := ResponseData{Occurrences: count}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type RequestData struct {
	Length      int64  `json:"length"`
	Query       string `json:"query"`
	DocData     bool   `json:"doc-data"`    // return document IDs and documents
	Surrounding bool   `json:"surrounding"` // return before and after 50 words of query
}

type ResponseData struct {
	Occurrences int64    `json:"occurrences"`
	DocIDs      []uint32 `json:"doc-ids"`
	Documenets  []string `json:"documents"`
	Surrounding []string `json:"surrounding"`
}
