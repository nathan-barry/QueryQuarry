package handlers

import (
	"encoding/json"
	"io"
	"net/http"
)

func getReqData(reqData *RequestData, w http.ResponseWriter, r *http.Request) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
	}
	if len(body) == 0 {
		http.Error(w, "Body is empty", http.StatusBadRequest)
	}

	// Unmarshal request data
	err = json.Unmarshal(body, reqData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}
}
