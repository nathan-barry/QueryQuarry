package search

import (
	"encoding/binary"
	"log"
	"os"
)

// NOTE: This function allows overlapping sequences to count as different duplicates.
// So if our string is `aaaa` and we count how many times `aa` occurs, it will return 3,
// not 2. This is different from python's "aaaa".count("aa") which will say 2.
func CountOccurrences(textFile, saFile *os.File, query string) (int64, int64) {
	pointerSize, saSize := getSAInfo(textFile, saFile)

	// Binary search until match
	low, high := int64(0), saSize-1
	for low <= high {
		mid := (low + high) / 2

		textIndex := readSuffixArray(saFile, pointerSize, mid)

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			// TODO: This is currently truncating the query, not correct
			// Is setting `high = mid - 1` correct?
			querySize = saSize - textIndex
		}

		substr := readText(textFile, textIndex, querySize)

		if substr == query {
			// Match, perform binary search twice to find count
			firstSAIndex := findFirstOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			LastSAIndex := findLastOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			return firstSAIndex, LastSAIndex
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return -1, -1
}

func findFirstOccurrence(textFile, saFile *os.File, pointerSize, saSize, mid int64, query string) int64 {
	firstOccurrence := mid

	// Binary search
	low, high := int64(0), mid
	for low <= high {
		mid := (low + high) / 2

		textIndex := readSuffixArray(saFile, pointerSize, mid)

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			querySize = saSize - textIndex
		}

		substr := readText(textFile, textIndex, querySize)

		if substr == query {
			firstOccurrence = mid
			high = mid - 1 // continue searching left half
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return firstOccurrence
}

func findLastOccurrence(textFile, saFile *os.File, pointerSize, saSize, mid int64, query string) int64 {
	lastOccurrence := mid

	// Binary search
	low, high := mid, saSize
	for low <= high {
		mid := (low + high) / 2

		textIndex := readSuffixArray(saFile, pointerSize, mid)

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			querySize = saSize - textIndex
		}

		substr := readText(textFile, textIndex, querySize)

		if substr == query {
			lastOccurrence = mid
			low = mid + 1 // continue searching right half
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return lastOccurrence
}

func getSAInfo(textFile, saFile *os.File) (int64, int64) {
	textFileInfo, err := textFile.Stat()
	if err != nil {
		log.Fatalf("failed to get file info: %v", err)
	}

	saFileInfo, err := saFile.Stat()
	if err != nil {
		log.Fatalf("failed to get file info: %v", err)
	}
	return saFileInfo.Size() / textFileInfo.Size(), textFileInfo.Size()
}

func readSuffixArray(saFile *os.File, pointerSize, index int64) int64 {
	offset := index * pointerSize

	_, err := saFile.Seek(offset, 0)
	if err != nil {
		log.Fatalf("failed to seek saFile: %v", err)
	}

	buf := make([]byte, pointerSize) // TODO, move this out and reuse
	_, err = saFile.Read(buf)
	if err != nil {
		log.Fatalf("failed to read bytes from saFile: %v", err)
	}

	fullBuf := make([]byte, 8) // TODO, move this out and reuse
	copy(fullBuf, buf)

	return int64(binary.LittleEndian.Uint64(fullBuf))
}

func readText(textFile *os.File, start, length int64) string {
	_, err := textFile.Seek(start, 0)
	if err != nil {
		log.Fatalf("failed to seek textFile: %v", err)
	}

	buf := make([]byte, length) // TODO, move this out and reuse
	_, err = textFile.Read(buf)
	if err != nil {
		log.Fatalf("failed to read bytes from textFile: %v", err)
	}

	return string(buf)
}
