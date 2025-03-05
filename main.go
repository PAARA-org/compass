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
	accessibleMode   bool     // When true, it turns on accessible color mode
	accessibleColors string   // Comma separated values of accessible colors
	colors           []string // Slice of colors to use for accessibility
	darkMode         bool     // Turns on dark mode
	expireInterval   int      // How quickly the bearing dots expire
	refreshInterval  int      // How quickly the browser should pull fresh data
	maxBearings      int      // How many bearing dots should we cache
	maxTableRows     int      // How many recent bearings to be displayed in the table
	paddedTimestamp  bool     // When true, we use 15 character timestamp
	re               *regexp.Regexp
	bearings         []Bearing
	mu               sync.Mutex
)

func main() {
	flag.BoolVar(&accessibleMode, "accessible", false, "Enable accessible color mode")
	flag.StringVar(&accessibleColors, "colors", "#2c7bb6,#abd9e9,#ffffbf,#fdae61,#d7191c", "5 colors to use for displaying magnitude, low to high")
	flag.BoolVar(&darkMode, "darkmode", false, "Enable dark mode")
	flag.IntVar(&expireInterval, "expire", 2000, "Bearing expire interval in milliseconds")
	flag.IntVar(&maxBearings, "bearings", 20, "Max bearings to cache")
	flag.IntVar(&maxTableRows, "rows", 5, "Max table rows to display")
	flag.BoolVar(&paddedTimestamp, "paddedTimestamp", false, "Pad timestamps to 15 digits")
	flag.IntVar(&refreshInterval, "refresh", 200, "Refresh interval in milliseconds")
	flag.Parse()

	go readInput()

	var err error
	colors, err = parseColors(accessibleColors)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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

func parseColors(csv string) (colors []string, err error) {
	hexPattern := regexp.MustCompile("^#[0-9A-Fa-f]{6}$")
	colors = strings.Split(csv, ",")
	if len(colors) < 5 {
		err = fmt.Errorf("Could not detect 5 comma separated colors.")
		return
	}

	for i, color := range colors {
		if !strings.HasPrefix(color, "#") || !hexPattern.MatchString(color) {
			err = fmt.Errorf("Color at index %d does not have the expected format (eg: #aa00b7)", i)
		}
	}
	return
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
