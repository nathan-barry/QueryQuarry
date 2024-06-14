package search

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func CountOccurrences(filename, query string) int64 {
	fmt.Println("filename:", filename)
	fmt.Println("query:", query)

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
	fmt.Println("SA pointer size:", pointerSize)
	fmt.Println("SA Size:", saSize)

	low, high := int64(0), saSize-1

	for low <= high {
		mid := (low + high) / 2

		textIndex, err := readSuffixArray(saFile, pointerSize, mid)
		if err != nil {
			log.Fatalf("failed to read suffix array: %v", err)
		}

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			querySize = saSize - textIndex
		}

		substr, err := readText(textFile, textIndex, querySize)
		if err != nil {
			log.Fatalf("failed to read text: %v", err)
		}

		fmt.Printf("mid: %v, substr: %s\n", mid, substr)

		if substr == query {
			fo := findFirstOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			lo := findLastOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			fmt.Printf("First Occurrence: %v, Last Occurrence: %v\n", fo, lo)
			return lo - fo
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return int64(-1)
}

func findFirstOccurrence(textFile, saFile *os.File, pointerSize, saSize, mid int64, query string) int64 {
	low, high := int64(0), mid

	firstOccurrence := int64(-1)

	var substr string

	for low <= high {
		mid := (low + high) / 2

		textIndex, err := readSuffixArray(saFile, pointerSize, mid)
		if err != nil {
			log.Fatalf("failed to read suffix array: %v", err)
		}

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			querySize = saSize - textIndex
		}

		substr, err = readText(textFile, textIndex, querySize)
		if err != nil {
			log.Fatalf("failed to read text: %v", err)
		}

		fmt.Printf("mid: %v, substr: %s\n", mid, substr)

		if substr == query {
			firstOccurrence = mid
			high = mid - 1 // continue searching left half
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	fmt.Println("First Query:", substr)
	return firstOccurrence
}

func findLastOccurrence(textFile, saFile *os.File, pointerSize, saSize, mid int64, query string) int64 {
	low, high := mid, saSize

	lastOccurrence := int64(-1)
	var substr string

	for low <= high {
		mid := (low + high) / 2

		textIndex, err := readSuffixArray(saFile, pointerSize, mid)
		if err != nil {
			log.Fatalf("failed to read suffix array: %v", err)
		}

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			querySize = saSize - textIndex
		}

		substr, err = readText(textFile, textIndex, querySize)
		if err != nil {
			log.Fatalf("failed to read text: %v", err)
		}

		fmt.Printf("mid: %v, substr: %s\n", mid, substr)

		if substr == query {
			lastOccurrence = mid
			low = mid + 1 // continue searching right half
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	fmt.Println("Last Query:", substr)
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

func readSuffixArray(saFile *os.File, pointerSize, index int64) (int64, error) {
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

	return int64(binary.LittleEndian.Uint64(fullBuf)), nil
}

func readText(textFile *os.File, start, length int64) (string, error) {
	_, err := textFile.Seek(start, 0)
	if err != nil {
		log.Fatalf("failed to seek textFile: %v", err)
	}

	buf := make([]byte, length)
	_, err = textFile.Read(buf)
	if err != nil {
		log.Fatalf("failed to read bytes from saFile: %v", err)
	}

	return string(buf), nil
}
