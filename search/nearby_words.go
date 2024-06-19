package search

import (
	"log"
	"os"
)

const CONTEXT_SIZE = 128
const BEFORE_SIZE = 32

// Returns 128 bytes of text where the query appears
func NearbyWords(textFile, saFile *os.File, firstSAIndex, lastSAIndex int64) []string {
	if firstSAIndex < 0 || lastSAIndex < 0 {
		log.Fatal("Negative suffix array index, no occurrences")
	}

	count := lastSAIndex - firstSAIndex + 1
	pointerSize, _ := getSAInfo(textFile, saFile)

	contexts := make([]string, count)
	buf := make([]byte, CONTEXT_SIZE)

	// Loop through each occurrence
	j := 0
	for i := firstSAIndex; i <= lastSAIndex; i++ {
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
		contexts[j] = string(buf[:n])

		j++
	}

	return contexts
}

// TODO: Add helper that truncates document if bleeds into another document
