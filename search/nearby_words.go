package search

import (
	"bytes"
	"errors"
	"os"
)

const CONTEXT_SIZE = 128
const MAX_SENTENCES = 64

// Returns 128 bytes of text where the query appears
func NearbyWords(textFile, saFile *os.File, firstSAIndex, lastSAIndex int64, queryLength int) ([]string, []string, error) {
	if firstSAIndex < 0 || lastSAIndex < 0 {
		return nil, nil, errors.New("search.NearbyWords: invalid first or last SA Index, one is -1")
	}

	pointerSize, _, err := getSAInfo(textFile, saFile)
	if err != nil {
		return nil, nil, err
	}

	if lastSAIndex-firstSAIndex+1 > MAX_SENTENCES {
		lastSAIndex = firstSAIndex + MAX_SENTENCES - 1
	}

	before := make([]string, lastSAIndex-firstSAIndex+1)
	after := make([]string, lastSAIndex-firstSAIndex+1)
	buf := make([]byte, CONTEXT_SIZE)

	// Loop through each occurrence
	j := 0
	for i := lastSAIndex; i >= firstSAIndex; i-- {
		textIndex, err := readSuffixArray(saFile, pointerSize, i)
		if err != nil {
			return nil, nil, err
		}

		// Read before sentence
		startIndex := textIndex - CONTEXT_SIZE
		readNum := CONTEXT_SIZE
		if startIndex < 0 {
			readNum = int(startIndex)
			startIndex = 0
		}
		textFile.Seek(startIndex, 0)

		_, err = textFile.Read(buf)
		if err != nil {
			return nil, nil, err
		}

		// Cut off start document ID
		if idx := bytes.LastIndex(buf, StartTokenPrefix); idx != -1 {
			if idx+6 > readNum {
				idx = readNum - 6
			}
			before[j] = string(buf[idx+6 : readNum])
		} else {
			before[j] = string(buf[:readNum])
		}

		// Read before sentence
		startIndex = textIndex + int64(queryLength)
		textFile.Seek(startIndex, 0)

		n, err := textFile.Read(buf)
		if err != nil {
			return nil, nil, err
		}
		// Cut off end document ID
		if idx := bytes.Index(buf, StartTokenPrefix); idx != -1 {
			n = idx
		}
		after[j] = string(buf[:n])

		j++
	}

	return before, after, nil
}
