/* Copyright (c) 2025, kn6yuh@gmail.com, PAARA.org
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

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
