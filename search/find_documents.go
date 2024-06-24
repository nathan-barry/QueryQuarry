package search

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

const CHUNK_SIZE = 8192 // From a little testing, seems like a good size
var StartTokenPrefix = []byte{0xff, 0xff}

func GetNumDocs(sizeFile *os.File) uint32 {
	sizeMetadata, err := sizeFile.Stat()
	if err != nil {
		log.Fatal("Error getting sizeFile.Stat()")
	}
	return uint32((sizeMetadata.Size() - 1) / 8)
}

// This will return two arrays. []int DocIDs, []int textStarts
// First is useful for figuring out size of document from the dataset.split.size file
// Second is the start of the document in the text file.
// Can do text[textStart[i]:textStart[i]+docSize[i]] to grab entire document
func FindDocuments(textFile, saFile *os.File, firstSAIndex, lastSAIndex int64, numDocs uint32) []uint32 {
	if firstSAIndex < 0 || lastSAIndex < 0 {
		log.Fatal("Negative suffix array index, no occurrences")
	}
	count := lastSAIndex - firstSAIndex + 1

	docIDs := make([]uint32, count)

	pointerSize, _ := getSAInfo(textFile, saFile)

	// Loop through each occurrence
	i := 0
	for pos := firstSAIndex; pos <= lastSAIndex; pos++ { // TODO: Off by one possibly?
		// Backwards scan until find ID \ff \ff + 4 bytes
		// Question: This should be unique? Since it won't be valid unicode?
		textIndex := readSuffixArray(saFile, pointerSize, pos)

		docID := findDocID(textFile, textIndex, CHUNK_SIZE)
		if docID > numDocs {
			log.Fatalf("docID > numDocs. DocID: %v, SA Index: %v, firstSAIndex: %v, lastSAIndex: %v", docID, pos, firstSAIndex, lastSAIndex)
		}
		docIDs[i] = docID
		i++
	}

	return docIDs
}

func findDocID(textFile *os.File, seekPos, chunkSize int64) uint32 {
	buf := make([]byte, chunkSize)

	i := 0
	for {
		i++
		if seekPos < chunkSize {
			chunkSize = seekPos
			seekPos = 0
		} else {
			seekPos -= (chunkSize - 6) // slight overlap
			// TODO: edgecase initial position is last element
		}

		_, err := textFile.Seek(seekPos, 0)
		if err != nil {
			log.Fatal("Error seeking while finding start token for document")
		}

		n, err := textFile.Read(buf)
		if err != nil {
			log.Fatal("Error reading chunk in document for findStartToken")
		}

		index := bytes.LastIndex(buf[:n], StartTokenPrefix)
		if index != -1 && index+6 <= n && index-4 >= 0 {
			// handle edge case where docID has 0xffff. Move index to first
			index += bytes.Index(buf[index-4:index+2], StartTokenPrefix) - 4
			docID := binary.LittleEndian.Uint32(buf[index+2 : index+6])
			return docID
		}

		if seekPos == 0 {
			log.Fatal("Error finding Start Token, reached start, found none")
		}
	}
}
