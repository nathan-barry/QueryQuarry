package search

import (
	"bytes"
	"log"
	"os"
)

const CONTEXT_SIZE = 128
const MAX_SENTENCES = 64

// Returns 128 bytes of text where the query appears
func NearbyWords(textFile, saFile *os.File, firstSAIndex, lastSAIndex int64, queryLength int) ([]string, []string) {
	if firstSAIndex < 0 || lastSAIndex < 0 {
		log.Fatal("Negative suffix array index, no occurrences")
	}

	pointerSize, _ := getSAInfo(textFile, saFile)

	if lastSAIndex-firstSAIndex+1 > MAX_SENTENCES {
		lastSAIndex = firstSAIndex + MAX_SENTENCES - 1
	}

	before := make([]string, lastSAIndex-firstSAIndex+1)
	after := make([]string, lastSAIndex-firstSAIndex+1)
	buf := make([]byte, CONTEXT_SIZE)

	// Loop through each occurrence
	j := 0
	for i := lastSAIndex; i >= firstSAIndex; i-- {
		textIndex := readSuffixArray(saFile, pointerSize, i)

		// Read before sentence
		startIndex := textIndex - CONTEXT_SIZE
		n := CONTEXT_SIZE
		if startIndex < 0 {
			n = int(startIndex)
			startIndex = 0
		}
		textFile.Seek(startIndex, 0)

		_, err := textFile.Read(buf)
		if err != nil {
			log.Fatal("Error reading nearby words")
		}
		// Cut off start document ID
		if idx := bytes.LastIndex(buf, StartTokenPrefix); idx != -1 {
			before[j] = string(buf[idx+6 : n])
		} else {
			before[j] = string(buf[:n])
		}

		// Read before sentence
		startIndex = textIndex + int64(queryLength)
		textFile.Seek(startIndex, 0)

		n, err = textFile.Read(buf)
		if err != nil {
			log.Fatal("Error reading nearby words")
		}
		// Cut off end document ID
		if idx := bytes.Index(buf, StartTokenPrefix); idx != -1 {
			n = idx
		}
		after[j] = string(buf[:n])

		j++
	}

	return before, after
}

// TODO: Add helper that truncates document if bleeds into another document
