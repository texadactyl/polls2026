package main

import "fmt"

/*
reportByState - Report all history records for the given state.
*/
func reportByState(stateCode string) {

	// Error if database is closed.
	if !dbIsOpen {
		fatalText("reportByState: database is closed")
	}

	// Execute query.
	histories, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, stateCode))
	if !ok {
		sqlText := fmt.Sprintf(fmtSqlSelectHistoryByState, stateCode)
		fatalText("reportByState: query(%s) failed", sqlText)
	}

	if len(histories) < 1 {
		fmt.Printf("\tNo history records were found for state %s\n", stateCode)
		return
	}

	// For each row retrieved, report the column values.
	fmt.Printf("\n\tStart  End    ST  D%%   R%%   I%%   U%%  Pollster\n")
	fmt.Printf("\t-----  ----   --  --   --   --   --  --------\n")

	for _, historyRcd := range histories {
		fmt.Printf("\t%02d/%02d  %02d/%02d  %s  %02.0f   %02.0f   %02.0f   %02.0f  %s\n",
			historyRcd.startMonth,
			historyRcd.startDay,
			historyRcd.endMonth,
			historyRcd.endDay,
			historyRcd.state,
			historyRcd.dem,
			historyRcd.gop,
			historyRcd.ind,
			historyRcd.unk,
			historyRcd.pollster)
	}

	fmt.Println()
}
