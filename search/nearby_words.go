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

	contexts := make([]string, count)

	buf := make([]byte, CONTEXT_SIZE)

	pointerSize, _ := getSAInfo(textFile, saFile)

	j := 0
	// Loop through each occurrence
	for i := firstSAIndex; i <= lastSAIndex; i++ { // TODO: Off by one possibly?
		// Backwards scan until find ID \ff \ff + 4 bytes
		// Question: This should be unique? Since it won't be valid unicode?
		textIndex := readSuffixArray(saFile, pointerSize, i)

		// TODO Handle start edge case
		textFile.Seek(textIndex-BEFORE_SIZE, 0)
		textFile.Read(buf)
		contexts[j] = string(buf)

		j++
	}

	return contexts
}
