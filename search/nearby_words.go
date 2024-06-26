package search

import (
	"log"
	"os"
)

const CONTEXT_SIZE = 128
const BEFORE_SIZE = 32
const MAX_SENTENCES = 64

// Returns 128 bytes of text where the query appears
func NearbyWords(textFile, saFile *os.File, firstSAIndex, lastSAIndex int64) []string {
	if firstSAIndex < 0 || lastSAIndex < 0 {
		log.Fatal("Negative suffix array index, no occurrences")
	}

	pointerSize, _ := getSAInfo(textFile, saFile)

	if lastSAIndex-firstSAIndex+1 > MAX_SENTENCES {
		lastSAIndex = firstSAIndex + MAX_SENTENCES - 1
	}

	sentences := make([]string, lastSAIndex-firstSAIndex+1)
	buf := make([]byte, CONTEXT_SIZE)

	// Loop through each occurrence
	j := 0
	for i := lastSAIndex; i >= firstSAIndex; i-- {
		textIndex := readSuffixArray(saFile, pointerSize, i)

		startIndex := textIndex - BEFORE_SIZE
		if startIndex < 0 {
			startIndex = 0
		}
		textFile.Seek(startIndex, 0)

		n, err := textFile.Read(buf)
		if err != nil {
			log.Fatal("Error reading nearby words")
		}
		sentences[j] = string(buf[:n])

		j++
	}

	return sentences
}

// TODO: Add helper that truncates document if bleeds into another document
