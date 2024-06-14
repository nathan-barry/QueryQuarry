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

func QueryHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error reading body")
		http.Error(w, "Error reading body", http.StatusInternalServerError)
	}

	if len(body) == 0 {
		log.Fatal("Body is empty")
		http.Error(w, "Body is empty", http.StatusBadRequest)
	}

	var data RequestData
	err = json.Unmarshal(body, &data)
	fmt.Println("Body:", string(body))
	fmt.Println("JSON:", data)
	if err != nil {
		log.Fatal("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}
	fmt.Println("data.Length", data.Length)
	fmt.Println("data.Query", data.Query)

	filename := "data/wiki40b.test"
	t := time.Now()

	count := search.CountOccurrences(filename, data.Query)
	fmt.Println("count:", count)
	fmt.Println("Time taken:", time.Since(t).Seconds())

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
