package search

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

const CHUNK_SIZE = 4096

// This will return two arrays. []int DocIDs, []int textStarts
// First is useful for figuring out size of document from the dataset.split.size file
// Second is the start of the document in the text file.
// Can do text[textStart[i]:textStart[i]+docSize[i]] to grab entire document
func FindDocuments(textFile, saFile *os.File, firstSAIndex, lastSAIndex int64) ([]int64, []int64) {
	if firstSAIndex < 0 || lastSAIndex < 0 {
		log.Fatal("Negative suffix array index, no occurrences")
	}
	count := lastSAIndex - firstSAIndex + 1

	docIDs := make([]int64, count)
	textStarts := make([]int64, count)

	pointerSize, _ := getSAInfo(textFile, saFile)

	// Loop through each occurrence
	for i := firstSAIndex; i < lastSAIndex; i++ { // TODO: Off by one possibly?
		// Backwards scan until find ID \ff \ff + 4 bytes
		// Question: This should be unique? Since it won't be valid unicode?
		textIndex := readSuffixArray(saFile, pointerSize, i)

		findStartToken(textFile, textIndex, CHUNK_SIZE)
	}

	return docIDs, textStarts
}

func findStartToken(textFile *os.File, seekPos, chunkSize int64) (uint32, int64) {
	startTokenPrefix := []byte{0xff, 0xff}
	buf := make([]byte, chunkSize)

	for {
		if seekPos < chunkSize {
			chunkSize = seekPos
			seekPos = 0
		} else {
			seekPos -= chunkSize
		}

		_, err := textFile.Seek(seekPos, 0)
		if err != nil {
			log.Fatal("Error seeking while finding start token for document")
		}

		n, err := textFile.Read(buf)
		if err != nil {
			log.Fatal("Error reading chunk in document for findStartToken")
		}

		index := bytes.LastIndex(buf[:n], startTokenPrefix)

		if index != -1 && index+6 <= n {
			docID := binary.LittleEndian.Uint32(buf[index+2 : index+6])
			return docID, seekPos + int64(index)
		} else {
			// TODO: Edge case where docID not fully in chunk
			log.Fatal("DocID not fully in chunk: TODO COMPLETE THIS")
		}

		if seekPos == 0 {
			log.Fatal("Error finding Start Token, reached start, found none")
		}
	}
}
