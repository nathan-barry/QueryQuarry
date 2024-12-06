// This command returns various information about a dataset
// 1. Number of Bytes, Number of documents, average size and variance
// 2. Verifies unique IDs (should always be correct from load_dataset.py)
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"slices"

	"github.com/nathan-barry/QueryQuarry/search"
)

func main() {
	// Read filename from command line
	var dataset string
	flag.StringVar(&dataset, "data", "./data/wiki40b.test", "Enter path to dataset") // TODO: generalize this to any dataset in data dir
	flag.Parse()

	// Open files
	textFile, err := os.Open(dataset)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer textFile.Close()
	sizeFile, err := os.Open(dataset + ".size")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer sizeFile.Close()

	textMetadata, err := textFile.Stat()
	fmt.Println("File Name:", textMetadata.Name())
	fmt.Println("\tNumber of Bytes:", textMetadata.Size())

	sizeMetadata, err := sizeFile.Stat()
	fmt.Println("File Name:", sizeMetadata.Name())
	fmt.Println("\tNumber of Bytes:", sizeMetadata.Size())
	fmt.Println("Number of Documents:", (sizeMetadata.Size()-1)/8)

	checkDocuments(textFile, sizeFile, (sizeMetadata.Size()-1)/8)
}

func checkDocuments(textFile, sizeFile *os.File, numDocs int64) error {
	// Init buffers
	startBuf := make([]byte, 8)
	endBuf := make([]byte, 8)
	startReader := bytes.NewReader(startBuf)
	endReader := bytes.NewReader(endBuf)

	docLengths := make([]int, numDocs)

	// Loop through each document
	for i := int64(0); i < numDocs; i++ {
		// Get Start and End positions
		_, err := sizeFile.Seek(i*8, 0)
		if err != nil {
			log.Fatalf("failed to seek sizeFile: %v", err)
		}
		_, err = sizeFile.Read(startBuf)
		if err != nil {
			log.Fatalf("failed to read bytes from sizeFile 1: %v", err)
		}
		_, err = sizeFile.Read(endBuf)
		if err != nil {
			log.Fatalf("failed to read bytes from sizeFile 2: %v", err)
		}

		// Read bytes into int64
		_, err = startReader.Seek(0, 0)
		if err != nil {
			log.Fatalf("failed to seek to start of reader: %v", err)
		}
		_, err = endReader.Seek(0, 0)
		if err != nil {
			log.Fatalf("failed to seek start of reader: %v", err)
		}
		var startPos int64
		var endPos int64
		err = binary.Read(startReader, binary.LittleEndian, &startPos)
		if err != nil {
			log.Fatalf("failed to read number: %v", err)
		}
		err = binary.Read(endReader, binary.LittleEndian, &endPos)
		if err != nil {
			log.Fatalf("failed to read number: %v", err)
		}
		docSize := endPos - startPos

		// Load Document
		docBuf := make([]byte, docSize)
		_, err = textFile.Seek(startPos, 0)
		if err != nil {
			log.Fatalf("failed to seek textFile: %v", err)
		}

		_, err = textFile.Read(docBuf)
		if err != nil {
			log.Fatalf("failed to read bytes from textFile: %v", err)
		}

		// Ensure no document has invalid startTokenPrefix tags
		// NOTE: Should never panic, python preprocessing script ensures valid UTF8
		if bytes.Contains(docBuf[6:], search.StartTokenPrefix) {
			index := bytes.Index(docBuf[6:], search.StartTokenPrefix)
			log.Fatal("CONTAINS INVALID ID", i, numDocs, index)
		}

		docLengths[i] = int(docSize) - 6
	}
	fmt.Println("\tAll documents have valid IDs")

	// Calculate average length and std
	average := calcAverageLength(docLengths)
	std := calcLengthStd(docLengths, average)

	fmt.Printf("Doc Stats (in bytes):\n\tAvg. length: %v\n\tAvg. std: %v\n\tMax length: %v\n\tMin length: %v\n",
		average, std, slices.Max(docLengths), slices.Min(docLengths))

	return nil
}

func calcAverageLength(docLengths []int) float64 {
	total := 0.0
	for docLength := range docLengths {
		total += float64(docLength)
	}
	return total / float64(len(docLengths))
}

func calcLengthStd(docLengths []int, average float64) float64 {
	summation := 0.0
	for docLength := range docLengths {
		summation += math.Pow(float64(docLength)-average, 2)
	}
	return math.Sqrt(summation / float64(len(docLengths)))
}
