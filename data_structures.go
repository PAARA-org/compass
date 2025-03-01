package main

import "time"

type Bearing struct {
	Degree    float64
	Time      string
	Timestamp time.Time
}

type BearingForTemplate struct {
	Degree  float64
	Time    string
	MsecAgo int64
	X       float64
	Y       float64
	Color   string
	Index   int
}

type BT struct {
	Bearings []BearingForTemplate
	Refresh  int // Refresh interval in seconds
	Expiry   int // Expiry interval in milliseconds
}
