package search

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
)

const INT64_SIZE = 8

func RetrieveDocuments(csvWriter *csv.Writer, textFile, sizeFile *os.File, docIDs []uint32) error {
	// Init buffers
	s := make([]byte, 8)
	e := make([]byte, 8)
	startBuf := bytes.NewReader(s)
	endBuf := bytes.NewReader(e)

	// Loop
	for i := 0; i < len(docIDs); i++ {
		// Get Start and End positions
		_, err := sizeFile.Seek(int64(docIDs[i]-1)*INT64_SIZE, 0) // IDs start at 1, need to subtract 1 to index at 0
		if err != nil {
			log.Fatalf("failed to seek textFile: %v", err)
		}

		_, err = sizeFile.Read(s)
		if err != nil {
			log.Fatalf("failed to read bytes from textFile: %v", err)
		}
		_, err = sizeFile.Read(e)
		if err != nil {
			log.Fatalf("failed to read bytes from textFile: %v", err)
		}

		// Read bytes into int64
		var startPos int64
		var endPos int64
		binary.Read(startBuf, binary.LittleEndian, &startPos)
		binary.Read(endBuf, binary.LittleEndian, &endPos)
		docSize := endPos - startPos

		// Get Document
		docBuf := make([]byte, docSize)
		_, err = textFile.Seek(startPos, 0)
		if err != nil {
			log.Fatalf("failed to seek textFile: %v", err)
		}

		_, err = textFile.Read(docBuf)
		if err != nil {
			log.Fatalf("failed to read bytes from textFile: %v", err)
		}

		if err := csvWriter.Write([]string{fmt.Sprint(docIDs[i]), string(docBuf)}); err != nil {
			return errors.New("Issue writing to csv writer") // TODO: Add more errors like this
		}
	}

	return nil
}
