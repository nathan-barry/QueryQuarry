package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/nathan-barry/query-quarry/handlers"
)

const LOCALHOST = "http://localhost:8080/query"

func main() {
	// Read filename from command line
	var filename string
	flag.StringVar(&filename, "file", "", "Path to file with queries")
	flag.Parse()

	// Open file
	queryFile, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error opening the following file:", filename)
	}
	defer queryFile.Close()

	// Initialize client
	client := &http.Client{}

	// Loop through each query, make request
	scanner := bufio.NewScanner(queryFile)
	for scanner.Scan() {
		fmt.Print(scanner.Text())

		// Create new request
		requestData := handlers.RequestData{
			Length: int64(len(scanner.Text())),
			Query:  scanner.Text(),
		}
		jsonData, err := json.Marshal(requestData)
		if err != nil {
			log.Fatal("Error marshalling json")
		}

		req, err := http.NewRequest("POST", LOCALHOST, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatal("Error with new request")
		}
		req.Header.Set("Content-Type", "application/json")

		// Send request to server
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Error sending request")
		}
		defer resp.Body.Close()

		// Print result
		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("Error reading from response body")
			}
			var responseData handlers.ResponseData
			json.Unmarshal(body, &responseData)

			fmt.Println(":", responseData.Occurrences)
		} else {
			log.Fatal("Bad status code:", resp.StatusCode)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading the file: %s", err)
	}
}
