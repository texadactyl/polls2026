package main

import (
	"fmt"
	"os"
	"strings"
)

const minEndMonth = 10

// pollOfPolls structure for state totals.
type typeStateTotals struct {
	count float64
	dem   float64
	gop   float64
	ind   float64
	unk   float64
}

// pollOfPolls structure for winners.
type typeWinner struct {
	demString string
	demCount  int64
	gopString string
	gopCount  int64
	indString string
	indCount  int64
	unkString string
	unkCount  int64
}

// Current Senate records.
type typeSenateRecord struct {
	stateCode string // state code (E.g. AK for Alaska)
	holderA   string // dem or gop
	holderB   string // dem or gop
	upA       int64  // Senator A up for re-election?
	upB       int64  // Senator B up for re-election?
}

var senateMap map[string]typeSenateRecord

const pathFileCurrentSenate = "current_senate.txt"

// pollOfPolls - run a mock poll of polls on all the states that we have poll data for.
// If stateCode is empty, run on all states; else run for only that state.
func pollOfPolls(stateCode string, gopPctIncrement float64) {

	var totals typeStateTotals
	totals.count = 0.0
	totals.dem = 0.0
	totals.gop = 0.0
	totals.ind = 0.0
	totals.unk = 0.0

	var winner typeWinner
	winner.demString = ""
	winner.gopString = ""
	winner.indString = ""
	winner.unkString = ""
	winner.demCount = 0
	winner.gopCount = 0
	winner.indCount = 0
	winner.unkCount = 0

	// Quick return if database is closed.
	if !dbIsOpen {
		fatalText("reportByState: database is closed")
	}

	// Execute query.
	histories, ok := query(fmtSqlSelectHistoryInOct)
	if !ok {
		fatalText("reportByState: query(%s) failed", fmtSqlSelectHistoryInOct)
	}

	// Retrieved any rows?
	if len(histories) < 1 {
		fmt.Println("\tNo history records were found for October")
		return
	}

	fmt.Println()
	logger("\tGOP increment: %05.2f", gopPctIncrement)
	logger("\tST    D%%      R%%      I%%      U%%")
	logger("\t--  -----   -----   -----   -----")
	currentStateCode := "" // force changing states for the first encountered history record.

	for _, historyRcd := range histories {
		// Skip current state?
		if stateCode != "" && stateCode != historyRcd.state {
			continue // Yes, skip.
		}
		// Before October?
		if historyRcd.endMonth < minEndMonth {
			continue // Yes, skip.
		}

		// Newly encountered state code?
		if currentStateCode != historyRcd.state {
			// Yes, emit data accumulated.
			pollStateResultEmit(currentStateCode, &totals, gopPctIncrement)
			// Record the winner status for the state.
			appendWinner(&winner, currentStateCode, totals)
			// Change states.
			currentStateCode = historyRcd.state
			// Reset state totals.
			totals.count = 0.0
			totals.dem = 0.0
			totals.gop = 0.0
			totals.ind = 0.0
			totals.unk = 0.0
		}

		// Accumulate totals for the current state.
		totals.count += 1.0
		totals.dem += historyRcd.dem
		totals.gop += historyRcd.gop
		totals.ind += historyRcd.ind
		totals.unk += historyRcd.unk
	}

	pollStateResultEmit(currentStateCode, &totals, gopPctIncrement)
	appendWinner(&winner, currentStateCode, totals)

	fmt.Println()
	logger("Dem winners (%d): %s", winner.demCount, winner.demString)
	logger("GOP winners (%d): %s", winner.gopCount, winner.gopString)
	logger("Ind winners (%d): %s", winner.indCount, winner.indString)
	logger("Unk winners (%d): %s", winner.unkCount, winner.unkString)
	fmt.Println()

}

// pollStateResultEmit - Emit results for one state. Also, each total is replaced by the average.
func pollStateResultEmit(stateCode string, totals *typeStateTotals, gopPctIncrement float64) {
	if totals.count < 1.0 {
		return
	}

	totals.dem /= totals.count
	totals.gop /= totals.count
	totals.ind /= totals.count
	totals.unk /= totals.count

	if totals.unk < gopPctIncrement {
		totals.gop += totals.unk
		totals.unk -= 0.0
	} else {
		totals.gop += gopPctIncrement
		totals.unk -= gopPctIncrement
	}

	logger("\t%s  %05.2f   %05.2f   %05.2f   %05.2f",
		stateCode,
		totals.dem,
		totals.gop,
		totals.ind,
		totals.unk)
}

// appendWinner - Determine winner and add them to the appropriate slice.
func appendWinner(winner *typeWinner, stateCode string, totals typeStateTotals) {
	if totals.count < 1.0 {
		return
	}

	if totals.dem > totals.gop && totals.dem > totals.ind && totals.dem > totals.unk {
		winner.demString += stateCode + " "
		winner.demCount += 1
		return
	}
	if totals.gop > totals.dem && totals.gop > totals.ind && totals.gop > totals.unk {
		winner.gopString += stateCode + " "
		winner.gopCount += 1
		return
	}
	if totals.ind > totals.gop && totals.ind > totals.dem && totals.ind > totals.unk {
		winner.indString += stateCode + " "
		winner.indCount += 1
		return
	}

	// Nobody won.
	winner.unkString += stateCode + " "
	winner.unkCount += 1

}

// loadCurrentSenate loads the current senate map.
func loadCurrentSenate() {
	buffer := make([]byte, 2048) // current size in storage: 950 bytes
	senateMap := make(map[string]typeSenateRecord)
	var curSenateRecord typeSenateRecord
	var strArray []string
	var ix int
	var bStateCode string

	// Open the senate file.
	fileHandle, err := os.Open(pathFileCurrentSenate)
	if err != nil {
		fatalErr(err, "loadCurrentSenate: os.Open(%s) failed", pathFileCurrentSenate)
	}
	defer fileHandle.Close()

	// Read all the bytes into a buffer.
	num, err := fileHandle.Read(buffer)
	if err != nil {
		fatalErr(err, "loadCurrentSenate: fileHandle.Read(%s) failed", pathFileCurrentSenate)
	}

	// Trim then split all the records into a convenient string array.
	strArray = strings.Split(strings.Trim(string(buffer[:num]), " "), "\n")

	// Build the state-senator map.
	for ix = 0; ix < len(strArray); ix++ {
		// Process senator A.
		ix, curSenateRecord.stateCode, curSenateRecord.holderA, curSenateRecord.upA = getSenateRecord(strArray, ix, len(strArray))
		if ix == -1 {
			break
		}

		// Process senator B from the same state.
		ix, bStateCode, curSenateRecord.holderB, curSenateRecord.upB = getSenateRecord(strArray, ix+1, len(strArray))
		if ix == -1 {
			fatalText("loadCurrentSenate: state code (%s) has an entry for senator A but is missing an entry for senator B", curSenateRecord.stateCode)
		}
		if bStateCode != curSenateRecord.stateCode {
			fatalText("loadCurrentSenate: state code (%s) mismatch for senator B, expected: %s", bStateCode, curSenateRecord.stateCode)
		}

		// Store this state's senator A&B map entry.
		senateMap[curSenateRecord.stateCode] = curSenateRecord
	}

	logger("loadCurrentSenate: %d current senate map entries", len(senateMap))
}

// getSenateRecord parses a senate record from the string array.
// Start the scan from index jxStart.
// Index value jxMax is the upper bound.
// Skip over comment records.
// Recognize end of array record (-END-).
func getSenateRecord(strArray []string, jxStart, jxMax int) (int, string, string, int64) {
	var jx int
	var stateCode, holder string
	var up int64
	for jx = jxStart; jxStart < jxMax; jx++ {
		// Hit end of array?
		if strArray[jx] == "-END-" {
			return -1, "", "", -1 // Hit end of array.
		}

		// Comment or nil entry?
		if strings.HasPrefix(strArray[jx], "#") || strings.TrimSpace(strArray[jx]) == "" {
			continue // YES
		}

		// Got a significant entry.
		_, err := fmt.Sscanf(strArray[jx], "%s %s %d", &stateCode, &holder, &up)
		if err != nil {
			fatalErr(err, "getSenateRecord: fmt.Sscanf(%s) failed", strArray[jxStart])
		}

		// Return hit-index + data.
		return jx, stateCode, holder, up
	}

	// Ran off the end of the array!
	fatalText("getSenateRecord: jx (%d) >= jxMax (%d); jxStart = %d", jx, jxMax, jxStart)
	return -1, "", "", -1 // Never executed, just making the GoLang compiler happy.
}
