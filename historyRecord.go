package main

// typeHistoryRecord is one row of the history table, with the primary key and timestamp removed and the unknown field added.
type typeHistoryRecord struct {
	// hashish omitted
	startStamp string
	state      string
	dem        float64
	gop        float64
	ind        float64
	unk        float64
	startMonth int64
	startDay   int64
	endMonth   int64
	endDay     int64
	pollster   string
}
