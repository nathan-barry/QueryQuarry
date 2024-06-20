package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nathan-barry/QueryQuarry/search"
)

const WIKI_40B = "data/wiki40b.test"

func CountHandler(w http.ResponseWriter, r *http.Request) {
	t := time.Now()

	// Read body
	var reqData RequestData
	getReqData(&reqData, w, r)

	// Open files
	textFile, err := os.Open(WIKI_40B)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer textFile.Close()
	saFile, err := os.Open(WIKI_40B + ".table.bin")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer saFile.Close()

	// Count occurrences
	firstSAIndex, lastSAIndex := search.CountOccurrences(textFile, saFile, reqData.Query)
	countTime := time.Since(t).Seconds()
	t = time.Now()

	// Get Nearby Words for Each Occurance
	var sentences []string
	count := int64(0)
	if firstSAIndex >= 0 && lastSAIndex >= 0 { // Both -1 if no occurrences
		count = lastSAIndex - firstSAIndex + 1
		sentences = search.NearbyWords(textFile, saFile, firstSAIndex, lastSAIndex)
	}

	// Log information
	log.Printf("QUERY: \"%v\", COUNT: %v, COUNT_TIME: %v, SENTENCE_TIME: %v",
		reqData.Query, count, countTime, time.Since(t).Seconds())

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
