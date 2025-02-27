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
	"html/template"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Bearing struct {
	Degree float64
	Time   string
}

var (
	refreshInterval int
	maxBearings     int
	paddedTimestamp bool
	re              *regexp.Regexp
	bearings        []Bearing
	mu              sync.Mutex
)

func main() {
	flag.IntVar(&refreshInterval, "refresh", 5, "Refresh interval in seconds")
	flag.IntVar(&maxBearings, "bearings", 20, "Max bearings to cache")
	flag.BoolVar(&paddedTimestamp, "paddedTimestamp", false, "Pad timestamps to 15 digits")
	flag.Parse()

	go readInput()

	http.HandleFunc("/", serveCompass)
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

		degree, err1 := strconv.ParseFloat(matches[1], 64)
		timestamp, err2 := strconv.ParseInt(matches[4], 10, 64)

		if err1 != nil || err2 != nil {
			continue
		}

		mu.Lock()
		if len(bearings) >= maxBearings {
			bearings = bearings[:maxBearings-1]
		}
		bearings = append([]Bearing{{
			Degree: degree / 10.0,
			Time:   time.UnixMilli(timestamp).Format("15:04:05.000"),
		}}, bearings...)
		mu.Unlock()
	}
}

func serveCompass(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	const tmpl = `<!DOCTYPE html>
<html>
<head>
    <title>Radial Compass</title>
    <meta http-equiv="refresh" content="{{.Refresh}}">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; }
        svg { margin: 20px auto; display: block; }
        table { margin: auto; border-collapse: collapse; }
        td, th { padding: 8px; border: 1px solid #ddd; }
    </style>
</head>
<body>
    <h1>Radial Bearing Display</h1>
    <svg width="400" height="400" viewBox="-200 -200 400 400">
        <circle cx="0" cy="0" r="190" fill="none" stroke="#ccc" stroke-width="1"/>
        {{range .Bearings}}
        <circle cx="{{.X}}" cy="{{.Y}}" r="5" fill="{{.Color}}"/>
        {{end}}
        <line x1="0" y1="-200" x2="0" y2="-180" stroke="black"/>
        <text x="0" y="-160" text-anchor="middle">N</text>
    </svg>

    <h2>Recent Bearings</h2>
    <table>
        <tr><th>Degree</th><th>Time</th></tr>
        {{range .Bearings}}
        <tr>
            <td>{{printf "%.1f" .Degree}}°</td>
            <td>{{.Time}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>`

	data := struct {
		Bearings []struct {
			Degree float64
			Time   string
			X      float64
			Y      float64
			Color  string
			Index  int
		}
		Refresh int
	}{
		Refresh: refreshInterval,
	}

	if len(bearings) > 0 {
		data.Bearings = make([]struct {
			Degree float64
			Time   string
			X      float64
			Y      float64
			Color  string
			Index  int
		}, len(bearings))

		for i, b := range bearings {
			rad := b.Degree * math.Pi / 180
			// Calculate radius progression
			radius := 190 * float64(maxBearings-i-1) / float64(maxBearings-1)

			data.Bearings[i] = struct {
				Degree float64
				Time   string
				X      float64
				Y      float64
				Color  string
				Index  int
			}{
				Degree: b.Degree,
				Time:   b.Time,
				X:      radius * math.Sin(rad),
				Y:      -radius * math.Cos(rad),
				Color:  fmt.Sprintf("hsl(%d, 100%%, 50%%)", (len(bearings)-i-1)*30),
				Index:  i,
			}
		}

	}

	t := template.Must(template.New("compass").Parse(tmpl))
	_ = t.Execute(w, data)
}
