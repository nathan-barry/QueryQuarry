package search

import (
	"encoding/binary"
	"log"
	"os"
)

func CountOccurrences(filename, query string) int64 {
	// Open files
	textFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	saFile, err := os.Open(filename + ".table.bin")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer textFile.Close()
	defer saFile.Close()

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
			fo := findFirstOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			lo := findLastOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			return lo - fo + 1
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return 0
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

	buf := make([]byte, pointerSize)
	_, err = saFile.Read(buf)
	if err != nil {
		log.Fatalf("failed to read bytes from saFile: %v", err)
	}

	fullBuf := make([]byte, 8)
	copy(fullBuf, buf)

	return int64(binary.LittleEndian.Uint64(fullBuf))
}

func readText(textFile *os.File, start, length int64) string {
	_, err := textFile.Seek(start, 0)
	if err != nil {
		log.Fatalf("failed to seek textFile: %v", err)
	}

	buf := make([]byte, length)
	_, err = textFile.Read(buf)
	if err != nil {
		log.Fatalf("failed to read bytes from textFile: %v", err)
	}

	return string(buf)
}
