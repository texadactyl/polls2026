package main

import "fmt"

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
	logger("\tST    D%%      R%%      I%%      U%%")
	logger("\t--  -----   -----   -----   -----")
	currentStateCode := "" // force changing states for the first encountered history record.

	for _, historyRcd := range histories {
		// Skip current state?
		if stateCode != "" && stateCode != historyRcd.state {
			continue // Yes, skip.
		}
		// Before October?
		if historyRcd.endMonth < 10 {
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
