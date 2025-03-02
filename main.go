/* Copyright (c) 2025, kn6yuh@gmail.com, PAARA.org
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	expireInterval  int
	refreshInterval int
	maxBearings     int
	maxTableRows    int
	paddedTimestamp bool
	re              *regexp.Regexp
	bearings        []Bearing
	mu              sync.Mutex
)

func main() {
	flag.IntVar(&refreshInterval, "refresh", 200, "Refresh interval in milliseconds")
	flag.IntVar(&expireInterval, "expire", 2000, "Bearing expire interval in milliseconds")
	flag.IntVar(&maxBearings, "bearings", 20, "Max bearings to cache")
	flag.IntVar(&maxTableRows, "rows", 5, "Max table rows to display")
	flag.BoolVar(&paddedTimestamp, "paddedTimestamp", false, "Pad timestamps to 15 digits")
	flag.Parse()

	go readInput()

	http.HandleFunc("/", serveCompass)
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		// Generate only the SVG and table HTML
		svgAndTableHTML := generateSVGAndTableHTML()
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(svgAndTableHTML))
	})
	fmt.Println("Server running on http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}

func readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	// C0090 - degrees with decimal
	//      000 - angle vector average magnitude
	//         000 - FIR filtered Doppler tone peak value
	//              1739170710000 - unix epoch time in milliseconds OR
	//            001739170710000 - zero padded unix epoch time in milliseconds (15 total)
	if paddedTimestamp {
		re = regexp.MustCompile(`^C(\d{4})(\d{3})(\d{3})(\d{15})$`)
	} else {
		re = regexp.MustCompile(`^C(\d{4})(\d{3})(\d{3})(\d{13})$`)
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		degree, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			continue
		}
		magnitude, err := strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			continue
		}
		timestamp, err := strconv.ParseInt(matches[4], 10, 64)
		if err != nil {
			continue
		}

		mu.Lock()
		if len(bearings) >= maxBearings {
			bearings = bearings[:maxBearings-1]
		}
		bearings = append([]Bearing{{
			Degree:    degree / 10.0,
			Time:      time.UnixMilli(timestamp).Format("15:04:05.000"),
			Timestamp: time.UnixMilli(timestamp),
			Magnitude: int(magnitude),
		}}, bearings...)
		mu.Unlock()
	}
}
