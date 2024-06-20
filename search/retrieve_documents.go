package search

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

const WIKI_40B = "data/wiki40b.test"
const INT64_SIZE = 8

func RetrieveDocuments(docIDs []uint32, docPos []int64) [][]string {
	csvData := [][]string{{"DocID", "Document"}}

	// Find the size by indexing into the dataset .size file and
	// grabbing the DocID'th int, which is the number of bytes of the doc?

	// Open files
	textFile, err := os.Open(WIKI_40B)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer textFile.Close()
	sizeFile, err := os.Open(WIKI_40B + ".size")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer sizeFile.Close()

	// Init buffers
	s := make([]byte, 8)
	startBuf := bytes.NewReader(s)
	e := make([]byte, 8)
	endBuf := bytes.NewReader(e)

	// Loop
	for i := 0; i < len(docIDs); i++ {
		fmt.Println("docIDs:", docIDs)
		// Get Size
		_, err = sizeFile.Seek(int64(docIDs[i])*8-8, 0)
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

		// Read size
		var startPos int64
		var endPos int64
		fmt.Printf("start: %x, \tend: %x\n", s, e)
		binary.Read(startBuf, binary.LittleEndian, &startPos)
		binary.Read(endBuf, binary.LittleEndian, &endPos)
		docSize := endPos - startPos
		fmt.Println("start:", startPos, "\tend:", endPos, "\tsize:", docSize)

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

		fmt.Println("DocPos actual:", docPos[i], "\tStartPos calculated:", startPos, "\tfirst - second =", docPos[i]-startPos)

		csvData = append(csvData, []string{fmt.Sprint(docIDs[i]), string(docBuf)})
	}

	// Then disk seek to docPos and read size bytes into buffer

	fmt.Println(csvData)

	return csvData
}
