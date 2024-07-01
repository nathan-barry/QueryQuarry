package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nathan-barry/QueryQuarry/search"
)

func CountHandler(w http.ResponseWriter, r *http.Request) {
	t := time.Now()

	// Read body
	var reqData RequestData
	getReqData(&reqData, w, r)

	// Open files
	textFile, err := os.Open(reqData.Dataset)
	if err != nil {
		http.Error(w, "Error opening dataset text file, does not exist", http.StatusNotFound)
		return
	}
	defer textFile.Close()

	saFile, err := os.Open(reqData.Dataset + ".table.bin")
	if err != nil {
		http.Error(w, "Error opening dataset SA file", http.StatusInternalServerError)
		return
	}
	defer saFile.Close()

	// Count occurrences
	firstSAIndex, lastSAIndex, err := search.CountOccurrences(textFile, saFile, reqData.Query)
	if err != nil {
		http.Error(w, "Error counting occurrences", http.StatusInternalServerError)
		return
	}
	countTime := time.Since(t).Seconds()
	t = time.Now()

	// Get Nearby Words for Each Occurance
	var before []string
	var after []string
	count := int64(0)
	if firstSAIndex >= 0 && lastSAIndex >= 0 { // Both -1 if no occurrences
		count = lastSAIndex - firstSAIndex + 1
		before, after, err = search.NearbyWords(textFile, saFile, firstSAIndex, lastSAIndex, len(reqData.Query))
		if err != nil {
			http.Error(w, "Error when finding nearby words", http.StatusInternalServerError)
			return
		}
	}

	// Log information
	log.Printf("QUERY: \"%v\", COUNT: %v, COUNT_TIME: %v, SENTENCE_TIME: %v",
		reqData.Query, count, countTime, time.Since(t).Seconds())

	// Send result back
	response := ResponseData{
		Occurrences: count,
		Before:      before,
		Query:       reqData.Query,
		After:       after,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type RequestData struct {
	Dataset string `json:"dataset"`
	Length  int64  `json:"length"`
	Query   string `json:"query"`
}

type ResponseData struct {
	Occurrences int64    `json:"occurrences"`
	Before      []string `json:"before"`
	Query       string   `json:"query"`
	After       []string `json:"after"`
}
