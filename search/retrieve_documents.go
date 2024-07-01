package search

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
)

const INT64_SIZE = 8

func RetrieveDocuments(csvWriter *csv.Writer, textFile, sizeFile *os.File, docIDs []uint32) error {
	// Init buffers
	startBuf := make([]byte, 8)
	endBuf := make([]byte, 8)
	startReader := bytes.NewReader(startBuf)
	endReader := bytes.NewReader(endBuf)

	// Loop
	for i := 0; i < len(docIDs); i++ {
		// Get Start and End positions
		_, err := sizeFile.Seek(int64(docIDs[i]-1)*INT64_SIZE, 0) // IDs start at 1, need to subtract 1 to index at 0
		if err != nil {
			return err
		}

		_, err = sizeFile.Read(startBuf)
		if err != nil {
			return err
		}
		_, err = sizeFile.Read(endBuf)
		if err != nil {
			return err
		}

		// Read bytes into int64
		_, err = startReader.Seek(0, 0)
		if err != nil {
			return err
		}
		_, err = endReader.Seek(0, 0)
		if err != nil {
			return err
		}

		var startPos int64
		var endPos int64
		binary.Read(startReader, binary.LittleEndian, &startPos)
		binary.Read(endReader, binary.LittleEndian, &endPos)
		docSize := endPos - startPos

		// Get Document
		docBuf := make([]byte, docSize)
		_, err = textFile.Seek(startPos, 0)
		if err != nil {
			return err
		}

		_, err = textFile.Read(docBuf)
		if err != nil {
			return err
		}

		if len(docBuf) < 6 {
			return errors.New("search.NearbyWords: DocBuf shorter than tag")
		}

		if err := csvWriter.Write([]string{fmt.Sprint(docIDs[i]), string(docBuf[6:])}); err != nil {
			return errors.New("search.NearbyWords: Issue writing to csv writer")
		}
	}

	return nil
}
