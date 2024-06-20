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

	"github.com/nathan-barry/QueryQuarry/handlers"
)

const LOCALHOST = "http://localhost:8080/"

const (
	COUNT = "count"
	CSV   = "csv"
)

func main() {
	// Read filename from command line
	var action string
	var filename string
	flag.StringVar(&action, "action", "count", "Choose 'count' or 'csv'")
	flag.StringVar(&filename, "file", "", "Path to file with queries")
	flag.Parse()

	if action != COUNT && action != CSV {
		log.Fatal("invalid action")
	}

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
		fmt.Printf("%s: ", scanner.Text())

		// Create new request
		requestData := handlers.RequestData{
			Length: int64(len(scanner.Text())),
			Query:  scanner.Text(),
		}
		jsonData, err := json.Marshal(requestData)
		if err != nil {
			log.Fatal("Error marshalling json")
		}

		req, err := http.NewRequest("POST", LOCALHOST+action, bytes.NewBuffer(jsonData))
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
			switch action {
			case COUNT:
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Fatal("Error reading from response body")
				}

				var responseData handlers.ResponseData
				json.Unmarshal(body, &responseData)

				fmt.Println(responseData.Occurrences)
			case CSV:
				outFile, err := os.Create(filename + ".output.csv")
				if err != nil {
					log.Fatal("Error creating file")
				}
				defer outFile.Close()

				_, err = io.Copy(outFile, resp.Body)
				if err != nil {
					log.Fatal("Error copying csv to out file")
				}
				fmt.Println("File downloaded successfully")
			}
		} else {
			log.Fatal("Bad status code:", resp.StatusCode)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading the file: %s", err)
	}
}
