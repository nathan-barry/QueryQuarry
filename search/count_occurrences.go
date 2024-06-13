package search

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func CountOccurrences(filename, query string) int {
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

	textIndex, err := readSuffixArray(saFile, pointerSize, saSize/2)
	fmt.Println("text index", textIndex)

	querySize := int64(len(query))
	stringBytes, err := readText(textFile, textIndex, querySize)
	fmt.Println("string", string(stringBytes))

	return 0
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

func readText(textFile *os.File, start, length int64) ([]byte, error) {
	_, err := textFile.Seek(start, 0)
	if err != nil {
		log.Fatalf("failed to seek textFile: %v", err)
	}

	buf := make([]byte, length)
	_, err = textFile.Read(buf)
	if err != nil {
		log.Fatalf("failed to read bytes from saFile: %v", err)
	}

	return buf, nil
}
