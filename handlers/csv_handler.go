package handlers

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nathan-barry/QueryQuarry/search"
)

func CSVHandler(w http.ResponseWriter, r *http.Request) {
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

	sizeFile, err := os.Open(reqData.Dataset + ".size")
	if err != nil {
		http.Error(w, "Error opening dataset size file", http.StatusInternalServerError)
		return
	}
	defer sizeFile.Close()

	numDocs, err := search.GetNumDocs(sizeFile)
	if err != nil {
		http.Error(w, "Error when getting number of documents", http.StatusInternalServerError)
		return
	}

	query := []byte(reqData.Query)

	// Count occurrences
	firstSAIndex, lastSAIndex, err := search.CountOccurrences(textFile, saFile, query)
	if err != nil {
		http.Error(w, "Error when counting occurrences", http.StatusInternalServerError)
		return
	}
	countTime := time.Since(t).Seconds()
	t = time.Now()

	// Get CSV Data
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	count := int64(0)
	if firstSAIndex >= 0 && lastSAIndex >= 0 { // Both -1 if no occurrences
		count = lastSAIndex - firstSAIndex + 1
		docIDs, err := search.FindDocuments(textFile, saFile, firstSAIndex, lastSAIndex, numDocs)
		if err != nil {
			http.Error(w, "Error when finding documents", http.StatusInternalServerError)
			return
		}

		if err := search.RetrieveDocuments(csvWriter, textFile, sizeFile, docIDs); err != nil {
			http.Error(w, "Failed to write CSV data", http.StatusInternalServerError)
			return
		}
	}

	// Log information
	log.Printf("QUERY: \"%v\", COUNT: %v, COUNT_TIME: %v, CSV_TIME: %v",
		reqData.Query, count, countTime, time.Since(t).Seconds())

	// Send result back
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=data.csv")
}
