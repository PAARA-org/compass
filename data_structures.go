/* Copyright (c) 2025, kn6yuh@gmail.com, PAARA.org
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package main

import "time"

type Bearing struct {
	Degree    float64   // Degrees with one decimal
	Time      string    // Time as string
	Timestamp time.Time // Time as timestamp
	Magnitude int       // angle vector average magnitude
}

type BearingForTemplate struct {
	Degree    float64
	Time      string
	Magnitude int
	MsecAgo   int64
	X         float64
	Y         float64
	Color     string
	Index     int
}

type BT struct {
	Bearings []BearingForTemplate
	DarkMode bool // Controls page rendering in dark mode
	Expiry   int  // Expiry interval in milliseconds
	MaxRows  int  // Max table rows to display
	Refresh  int  // Refresh interval in seconds
}
