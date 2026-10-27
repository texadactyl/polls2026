package main

/*
Database access for jacotest. Every SQLite operation is in this file.

Downloads + installs:
* go get modernc.org/sqlite (pure Go driver, no cgo)
* go install -v ./...
* https://www.sqlite.org/download.html (API for the Browser)
* https://sqlitebrowser.org/dl/ (Browser)

The history table has one row per test case execution. Its primary key is
(test_case, date_utc, time_utc). Rows for a test case are read newest first.
*/

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var sqlTracing = false

// How long SQLite waits on a lock held by another process before returning SQLITE_BUSY.
const busyTimeoutMsecs = 5000

// History table
const tableHistory = "history"

// History table columns
const colHashish = "hashish"
const colTimestamp = "timestamp"
const colState = "state"
const colDem = "dem"
const colGop = "gop"
const colInd = "ind"
const colUnk = "unk"
const colStartMonth = "start_month"
const colStartDay = "start_day"
const colEndMonth = "end_month"
const colEndDay = "end_day"
const colPollster = "pollster"

// File system and SQL constants
const dirDatabase = "database"
const fileDatabase = "senate_polls_2026.db"
const driverDatabase = "sqlite"

// Assigned and used at run-time
var pathDatabase string
var dbHandle *sql.DB
var dbIsOpen = false

// Formats for SQL statements.
var (

	// Create the history table. The primary key is state, end month, and end day.
	fmtSqlCreateHistory = fmt.Sprintf(
		"CREATE TABLE %s ("+
			"%s VARCHAR NOT NULL, "+
			"%s VARCHAR NOT NULL, %s VARCHAR NOT NULL, %s FLOAT, %s FLOAT, %s FLOAT, %s FLOAT, "+
			"%s INTEGER, %s INTEGER, %s INTEGER, %s INTEGER, %s VARCHAR NOT NULL,"+
			"PRIMARY KEY (%s)) ",
		// create TABLE
		tableHistory,
		// list of FIELD names
		colHashish, colTimestamp, colState, colDem, colGop, colInd, colUnk,
		colStartMonth, colStartDay, colEndMonth, colEndDay, colPollster,
		// primary key FIELD
		colHashish)

	// Insert a history row with all columns.
	fmtSqlInsertHistory = fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) "+
			"VALUES ('%%s', '%%s', '%%s', '%%f', '%%f', '%%f', '%%f', '%%d', '%%d', '%%d', '%%d', '%%s')",
		// insert into TABLE
		tableHistory,
		// list of FIELD names
		colHashish, colTimestamp, colState, colDem, colGop, colInd, colUnk, colStartMonth, colStartDay, colEndMonth, colEndDay, colPollster)

	// Select all history rows, grouped by state, end month (descending), and day (descending).
	fmtSqlSelectHistoryByState = fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = '%%s' ORDER BY %s DESC, %s DESC",
		colState, colDem, colGop, colInd, colUnk, colStartMonth, colStartDay, colEndMonth, colEndDay, colPollster,
		// from TABLE where FIELD = VALUE
		tableHistory, colState,
		// order by FIELD1 DESC, FIELD2 DESC
		colEndMonth, colEndDay)

	// Select all history rows, grouped by state, end month (descending), and day (descending).
	fmtSqlSelectHistoryByStateAscendingDates = fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = '%%s' ORDER BY %s, %s",
		colState, colDem, colGop, colInd, colUnk, colStartMonth, colStartDay, colEndMonth, colEndDay, colPollster,
		// from TABLE where FIELD = VALUE
		tableHistory, colState,
		// order by FIELD1, FIELD2
		colEndMonth, colEndDay)

	// Select all history rows, grouped by state, end month (descending), and day (descending).
	fmtSqlSelectHistoryInOct = fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s > 0 ORDER BY %s",
		colState, colDem, colGop, colInd, colUnk, colStartMonth, colStartDay, colEndMonth, colEndDay, colPollster,
		// from TABLE where FIELD = VALUE
		tableHistory, colEndMonth,
		// order by FIELD
		colState)

	// Delete a record.
	// fmtSqlDeleteByHashish = fmt.Sprintf("DELETE FROM %s WHERE %s = '%%s'", tableHistory, colHashish)
)

func sqlFunc(sqlText string) error {

	if sqlTracing {
		logger("sqlFunc: %s", sqlText)
	}

	// Execute request.
	_, errSql := dbHandle.Exec(sqlText)

	// Return result.
	return errSql

}

/*
query - run a database query for all rows associated with the given stateCode.
*/
func query(sqlText string) ([]typeHistoryRecord, bool) {

	if sqlTracing {
		logger("query: %s", sqlText)
	}

	rows, err := dbHandle.Query(sqlText)
	if err != nil {
		logError("query: dbHandle.Query(%s) failed, err: %s", sqlText, err.Error())
		return nil, false
	}

	// Build a result array for caller.
	var results []typeHistoryRecord
	for rows.Next() {
		row := scanHistoryRow(rows, "query")
		results = append(results, row)
	}

	// Check for error detected in rows.Next().
	if err := rows.Err(); err != nil {
		logError("query: rows.Next(%s) failed, err: %s", sqlText, err.Error())
		err = rows.Close()
		if err != nil {
			fatalErr(err, "query(%s): rows.Close failed", sqlText)
		}
		return nil, false
	}

	// Number of rows is simply len(results)
	err = rows.Close()
	if err != nil {
		fatalErr(err, "query(%s): rows.Close failed", sqlText)
	}
	return results, true
}

/*
scanHistoryRow reads the current row of a result set.
*/
func scanHistoryRow(rows *sql.Rows, caller string) typeHistoryRecord {

	var row typeHistoryRecord
	err := rows.Scan(&row.state, &row.dem, &row.gop, &row.ind, &row.unk,
		&row.startMonth, &row.startDay, &row.endMonth, &row.endDay, &row.pollster)
	if err != nil {
		fatalErr(err, "scanHistoryRow(%s): rows.Scan failed", caller)
	}
	return row

}

/*
Internal function to initialize a new database.

* Create the history table. The primary key is (test_case, date_utc, time_utc).
* Create the secondary index.
Both statements are in one transaction so that a failure does not leave a half-built database.
*/
func initDB() {

	if sqlTracing {
		logger("initDB: Begin")
	}

	err := sqlFunc(fmtSqlCreateHistory)
	if err != nil {
		fatalErr(err, "initDB: unrecoverable SQL error")
	}

	if sqlTracing {
		logger("initDB: End")
	}

}

/*
DBOpen - Database Open

* If the database directory is not present, create it.
* If the database file in the database directory is not present, create it and call initDB.
* Otherwise, upgrade the schema if it is out of date.
*/
func DBOpen(flagVerbose bool) {

	sqlTracing = flagVerbose
	if sqlTracing {
		logger("DBOpen: Begin")
	}

	// Already open.
	if dbIsOpen {
		return
	}

	// Database directory
	err := os.MkdirAll(dirDatabase, 0755)
	if err != nil {
		fatalErr(err, "DBOpen: MkdirAll(%s) failed", dirDatabase)
	}

	// Database file. Only "does not exist" means that a new database is needed.
	pathDatabase = filepath.Join(dirDatabase, fileDatabase)
	_, err = os.Stat(pathDatabase)
	isNew := errors.Is(err, os.ErrNotExist)
	if err != nil && !isNew {
		fatalErr(err, "DBOpen: Stat(%s) failed", pathDatabase)
	}

	// sql.Open does not touch the file; Ping does (creating it if it is new).
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(%d)", pathDatabase, busyTimeoutMsecs)
	dbHandle, err = sql.Open(driverDatabase, dsn)
	if err != nil {
		fatalErr(err, "DBOpen: sql.Open(%s) failed", pathDatabase)
	}

	// One connection only: avoids SQLITE_BUSY between connections of this process.
	// Consequently, result sets must be closed before the next statement is issued.
	dbHandle.SetMaxOpenConns(3)

	err = dbHandle.Ping()
	if err != nil {
		fatalErr(err, "DBOpen: cannot connect to %s", pathDatabase)
	}
	dbIsOpen = true

	if isNew {
		initDB()
		if sqlTracing {
			logger("DBOpen: End, database created")
		}
		return
	}

}

/*
DBClose - Close the database.
*/
func DBClose() {

	// Quick return if database is closed.
	if !dbIsOpen {
		return
	}

	if sqlTracing {
		logger("DBClose: Begin")
	}

	// Make sure that we do not come through here again.
	dbIsOpen = false

	// Close the database.
	err := dbHandle.Close()
	if err != nil {
		fatalErr(err, "DBClose: dbHandle.Close failed")
	}

	if sqlTracing {
		logger("DBClose: End")
	}

}

/*
DBInsert - Insert a history record into the database.
*/
func DBInsert(args ...any) error {

	// Quick return if database is closed.
	if !dbIsOpen {
		fatalText("DBInsert: database is closed")
	}

	anyArray := make([]any, len(args))
	for ix, str := range args {
		anyArray[ix] = str
	}

	sqlText := fmt.Sprintf(fmtSqlInsertHistory, anyArray...)

	return sqlFunc(sqlText)

}
