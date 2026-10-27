package main

import (
	"os"
	"path/filepath"
	"strings"
)

// loader - Download the file given by the url,
//
//	create a string array element for each downloaded string line,
//	and insert a new database history record for each string in the array.
func loader(startStamp, pathDirDownloads, url string) {
	// Download the file.
	downloadFileName := startStamp + ".txt"
	fullPath := filepath.Join(pathDirDownloads, downloadFileName)
	urlGet(url, fullPath)

	// Create a string array element for each downloaded string line.
	byteArray, err := os.ReadFile(fullPath)
	if err != nil {
		fatalErr(err, "os.ReadFile(%s) for reading returned an error", fullPath)
	}
	strArray := strings.SplitSeq(strings.TrimRight(string(byteArray), "\r\n"), "\n")

	// For each string in the array, insert a new database history record.
	countInserts := 0
	countDiscards := 0
	for line := range strArray {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			fatalText("expected at least 9 fields, got %d", len(fields))
		}
		hashish := hashStrArray(fields)
		state := fields[0]
		dem := stringToFloat64(fields[1], "dem")
		gop := stringToFloat64(fields[2], "gop")
		ind := stringToFloat64(fields[3], "ind")
		unk := computeUnknown(dem, gop, ind)
		startMonth := monthStringToInt64(fields[4], "startMonth")
		startDay := stringToInt64(fields[5], "startDay")
		endMonth := monthStringToInt64(fields[6], "endMonth")
		endDay := stringToInt64(fields[7], "endDay")
		pollster := strings.Join(fields[8:], " ")

		// Insert one record.
		err = DBInsert(hashish, startStamp, strings.ToUpper(state), dem, gop, ind, unk, startMonth, startDay, endMonth, endDay, pollster)
		if err == nil {
			countInserts++
			continue // no errors
		}
		// If duplicate key ("UNIQUE" in error message), then this insert-attempt was redundant.
		if !strings.Contains(err.Error(), "UNIQUE") {
			fatalErr(err, "main: DBInsert(%s) returned an unexpected error", hashish)
		}
		countDiscards++
	}

	logger("loader: %d inserts, %d discards", countInserts, countDiscards)

}
