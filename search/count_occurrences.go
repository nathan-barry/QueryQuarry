package search

import (
	"encoding/binary"
	"os"
)

// NOTE: This function allows overlapping sequences to count as different duplicates.
// So if our string is `aaaa` and we count how many times `aa` occurs, it will return 3,
// not 2. This is different from python's "aaaa".count("aa") which will say 2.
func CountOccurrences(textFile, saFile *os.File, query string) (int64, int64, error) {
	pointerSize, saSize, err := getSAInfo(textFile, saFile)
	if err != nil {
		return 0, 0, err
	}

	// Binary search until match
	low, high := int64(0), saSize-1
	for low <= high {
		mid := (low + high) / 2

		textIndex, err := readSuffixArray(saFile, pointerSize, mid)
		if err != nil {
			return 0, 0, err
		}

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			// TODO: This is currently truncating the query, not correct
			// Is setting `high = mid - 1` correct?
			querySize = saSize - textIndex
		}

		substr, err := readText(textFile, textIndex, querySize)
		if err != nil {
			return 0, 0, err
		}

		if substr == query {
			// Match, perform binary search twice to find count
			firstSAIndex, err := findFirstOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			if err != nil {
				return 0, 0, err
			}
			LastSAIndex, err := findLastOccurrence(textFile, saFile, pointerSize, saSize, mid, query)
			if err != nil {
				return 0, 0, err
			}
			return firstSAIndex, LastSAIndex, nil
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return -1, -1, nil
}

func findFirstOccurrence(textFile, saFile *os.File, pointerSize, saSize, mid int64, query string) (int64, error) {
	firstOccurrence := mid

	// Binary search
	low, high := int64(0), mid
	for low <= high {
		mid := (low + high) / 2

		textIndex, err := readSuffixArray(saFile, pointerSize, mid)
		if err != nil {
			return 0, err
		}

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			querySize = saSize - textIndex
		}

		substr, err := readText(textFile, textIndex, querySize)
		if err != nil {
			return 0, err
		}

		if substr == query {
			firstOccurrence = mid
			high = mid - 1 // continue searching left half
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return firstOccurrence, nil
}

func findLastOccurrence(textFile, saFile *os.File, pointerSize, saSize, mid int64, query string) (int64, error) {
	lastOccurrence := mid

	// Binary search
	low, high := mid, saSize
	for low <= high {
		mid := (low + high) / 2

		textIndex, err := readSuffixArray(saFile, pointerSize, mid)
		if err != nil {
			return 0, err
		}

		querySize := int64(len(query))
		if querySize > saSize-textIndex {
			querySize = saSize - textIndex
		}

		substr, err := readText(textFile, textIndex, querySize)
		if err != nil {
			return 0, err
		}

		if substr == query {
			lastOccurrence = mid
			low = mid + 1 // continue searching right half
		} else if substr < query {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return lastOccurrence, nil
}

func getSAInfo(textFile, saFile *os.File) (int64, int64, error) {
	textFileInfo, err := textFile.Stat()
	if err != nil {
		return 0, 0, err
	}

	saFileInfo, err := saFile.Stat()
	if err != nil {
		return 0, 0, err
	}
	return saFileInfo.Size() / textFileInfo.Size(), textFileInfo.Size(), nil
}

func readSuffixArray(saFile *os.File, pointerSize, index int64) (int64, error) {
	offset := index * pointerSize

	_, err := saFile.Seek(offset, 0)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, pointerSize) // TODO, move this out and reuse
	_, err = saFile.Read(buf)
	if err != nil {
		return 0, err
	}

	fullBuf := make([]byte, 8) // TODO, move this out and reuse
	copy(fullBuf, buf)

	return int64(binary.LittleEndian.Uint64(fullBuf)), nil
}

func readText(textFile *os.File, start, length int64) (string, error) {
	_, err := textFile.Seek(start, 0)
	if err != nil {
		return "", err
	}

	buf := make([]byte, length) // TODO, move this out and reuse
	_, err = textFile.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
