package search

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

const CHUNK_SIZE = 8192 // From a little testing, seems like a good size
var StartTokenPrefix = []byte{0xff, 0xff}

// This will return two arrays. []int DocIDs, []int textStarts
// First is useful for figuring out size of document from the dataset.split.size file
// Second is the start of the document in the text file.
// Can do text[textStart[i]:textStart[i]+docSize[i]] to grab entire document
func FindDocuments(textFile, saFile *os.File, firstSAIndex, lastSAIndex int64, numDocs uint32) ([]uint32, error) {
	if firstSAIndex < 0 || lastSAIndex < 0 {
		return nil, errors.New("search.FindDocuments: invalid first or last SA Index, one is -1")
	}
	count := lastSAIndex - firstSAIndex + 1

	docIDs := make([]uint32, count)

	pointerSize, _, err := getSAInfo(textFile, saFile)
	if err != nil {
		return docIDs, err
	}

	// Loop through each occurrence
	i := 0
	for pos := firstSAIndex; pos <= lastSAIndex; pos++ { // TODO: Off by one possibly?
		// Backwards scan until find ID \ff \ff + 4 bytes
		textIndex, err := readSuffixArray(saFile, pointerSize, pos)
		if err != nil {
			return docIDs, err
		}

		docID, err := findDocID(textFile, textIndex, CHUNK_SIZE)
		if err != nil {
			return docIDs, err
		}

		if docID > numDocs {
			if err != nil {
				return docIDs, errors.New("search.FindDocuments: docID greater than numDocs")
			}
		}
		docIDs[i] = docID
		i++
	}

	return docIDs, nil
}

func GetNumDocs(sizeFile *os.File) (uint32, error) {
	sizeMetadata, err := sizeFile.Stat()
	if err != nil {
		return 0, err
	}
	return uint32((sizeMetadata.Size() - 1) / 8), nil
}

func findDocID(textFile *os.File, seekPos, chunkSize int64) (uint32, error) {
	buf := make([]byte, chunkSize)

	i := 0
	for {
		i++
		if seekPos < chunkSize {
			chunkSize = seekPos
			seekPos = 0
		} else {
			seekPos -= (chunkSize - 6) // slight overlap
			// TODO: edgecase initial position is last element?
		}

		_, err := textFile.Seek(seekPos, 0)
		if err != nil {
			return 0, err
		}

		n, err := textFile.Read(buf)
		if err != nil {
			return 0, err
		}

		index := bytes.LastIndex(buf[:n], StartTokenPrefix)
		if index != -1 && index+6 <= n && index-4 >= 0 {
			// handle edge case where docID has 0xffff. Move index to first
			index += bytes.Index(buf[index-4:index+2], StartTokenPrefix) - 4
			docID := binary.LittleEndian.Uint32(buf[index+2 : index+6])
			return docID, nil
		}

		if seekPos == 0 {
			return 0, errors.New("search.findDocID: Error finding Start Token, reached start, found none")
		}
	}
}
