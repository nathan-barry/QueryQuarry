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
	"path"
	"strings"

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
	var dataset string
	var filename string
	flag.StringVar(&action, "action", "count", "Choose action: 'count' or 'csv'")
	flag.StringVar(&dataset, "data", "./data/wiki40b.test", "Enter path to dataset") // TODO: generalize this to any dataset in data dir
	flag.StringVar(&filename, "file", "", "Path to file with queries")
	flag.Parse()

	if action != COUNT && action != CSV {
		log.Fatal("invalid action")
	}

	// Initialize client
	client := &http.Client{}

	// Open file
	queryFile, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error opening the following file:", filename)
	}
	defer queryFile.Close()

	// Loop through each query, make request
	scanner := bufio.NewScanner(queryFile)

	switch action {
	case COUNT:
		cmdCount(client, scanner, dataset)
	case CSV:
		cmdCSV(client, scanner, dataset, filename)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading the file: %s", err)
	}
}

func cmdCount(client *http.Client, scanner *bufio.Scanner, dataset string) {
	for scanner.Scan() {
		fmt.Printf("%s: ", scanner.Text())

		// Send request to server
		req := createRequest(dataset, COUNT, scanner)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Error sending request")
		}
		defer resp.Body.Close()

		// Do something with the response
		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("Error reading from response body")
			}

			var responseData handlers.ResponseData
			json.Unmarshal(body, &responseData)

			fmt.Println(responseData.Occurrences)
		} else {
			log.Fatal("Bad status code:", resp.StatusCode)
		}
	}
}

func cmdCSV(client *http.Client, scanner *bufio.Scanner, dataset, filename string) {
	ext := path.Ext(filename)
	outFile, err := os.Create(strings.TrimSuffix(filename, ext) + ".out.csv")
	if err != nil {
		log.Fatal("Error creating file")
	}
	defer outFile.Close()

	for scanner.Scan() {
		fmt.Printf("%s: ", scanner.Text())

		req := createRequest(dataset, CSV, scanner)

		// Send request to server
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Error sending request")
		}
		defer resp.Body.Close()

		// Do something with the response
		if resp.StatusCode == http.StatusOK {
			_, err = io.Copy(outFile, resp.Body)
			if err != nil {
				log.Fatal("Error copying csv to out file")
			}
			fmt.Println("File downloaded successfully")
		} else {
			log.Fatal("Bad status code:", resp.StatusCode)
		}
	}
}

func createRequest(dataset, action string, scanner *bufio.Scanner) *http.Request {
	// Create new request
	requestData := handlers.RequestData{
		Dataset: dataset,
		Length:  int64(len(scanner.Text())),
		Query:   scanner.Text(),
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

	return req
}
