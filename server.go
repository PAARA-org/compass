/* Copyright (c) 2025, kn6yuh@gmail.com, PAARA.org
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package main

import (
	"bufio"
	"bytes"
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
	Degree    float64
	Time      string
	Timestamp time.Time
}

var (
	expireInterval  int
	refreshInterval int
	maxBearings     int
	paddedTimestamp bool
	re              *regexp.Regexp
	bearings        []Bearing
	mu              sync.Mutex
)

func main() {
	flag.IntVar(&refreshInterval, "refresh", 200, "Refresh interval in milliseconds")
	flag.IntVar(&expireInterval, "expire", 2000, "Bearing expire interval in milliseconds")
	flag.IntVar(&maxBearings, "bearings", 20, "Max bearings to cache")
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
			Degree:    degree / 10.0,
			Time:      time.UnixMilli(timestamp).Format("15:04:05.000"),
			Timestamp: time.UnixMilli(timestamp),
		}}, bearings...)
		mu.Unlock()
	}
}

func serveCompass(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	const tmpl = `<!DOCTYPE html>
<html>
<script>
function updateContent() {
    fetch('/update')
        .then(response => response.text())
        .then(html => {
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, 'text/html');

            const svgContainer = document.getElementById('svgContainer');
            const tableContainer = document.getElementById('tableContainer');

            svgContainer.innerHTML = doc.querySelector('svg').outerHTML;
            tableContainer.innerHTML = doc.querySelector('table').outerHTML;
        });
}

setInterval(updateContent, {{.Refresh}});
</script>
<head>
    <title>Radial Compass</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; }
        svg { margin: 20px auto; display: block; }
        table { margin: auto; border-collapse: collapse; }
        td, th { padding: 8px; border: 1px solid #ddd; }
    </style>
</head>
<body>
    <h1>Direction Finder Compass</h1>
	<div id="svgContainer">
		<svg width="400" height="400" viewBox="-200 -200 400 400">
			<circle cx="0" cy="0" r="190" fill="none" stroke="#ccc" stroke-width="1"/>

			<!-- Degree ticks and labels -->
			{{range $i := seq 0 23}}
				{{$angle := mul (toFloat64 $i) 15}}
				{{$radians := div (mul (sub 90 $angle) 3.14159265) 180}}
				{{$x1 := mul (cos $radians) 190}}
				{{$y1 := mul (sin $radians) -190}}
				{{$x2 := mul (cos $radians) 180}}
				{{$y2 := mul (sin $radians) -180}}
				<line x1="{{$x1}}" y1="{{$y1}}" x2="{{$x2}}" y2="{{$y2}}" stroke="black" stroke-width="0.5"/>

				{{if eq (mod $i 6) 0}}
					{{$labelX := mul (cos $radians) 165}}
					{{$labelY := mul (sin $radians) -165}}
					<text x="{{$labelX}}" y="{{$labelY}}" text-anchor="middle" dominant-baseline="middle" font-size="12">{{$angle}}°</text>
				{{end}}
			{{end}}

			{{$expiry := .Expiry}}
			{{range .Bearings}}
				{{ if gt .MsecAgo $expiry }}
					<circle cx="{{.X}}" cy="{{.Y}}" r="5" fill="none" stroke="black" stroke-width="1"/>
				{{ else }}
					<circle cx="{{.X}}" cy="{{.Y}}" r="5" fill="{{.Color}}"/>
				{{ end }}
			{{end}}
		</svg>
	</div>

	<h2>Recent Bearings</h2>
	<div id="tableContainer">
		<table>
			<tr><th>Degree</th><th>Time</th></tr>
			{{range .Bearings}}
			<tr>
				<td>{{printf "%.1f" .Degree}}°</td>
				<td>{{.Time}}</td>
			</tr>
			{{end}}
		</table>
	</div>
</body>
</html>`

	data := struct {
		Bearings []struct {
			Degree  float64
			Time    string
			MsecAgo int64
			X       float64
			Y       float64
			Color   string
			Index   int
		}
		Refresh int // Refresh interval in seconds
		Expiry  int // Expiry interval in milliseconds
	}{
		Refresh: refreshInterval,
		Expiry:  expireInterval,
	}

	if len(bearings) > 0 {
		data.Bearings = make([]struct {
			Degree  float64
			Time    string
			MsecAgo int64
			X       float64
			Y       float64
			Color   string
			Index   int
		}, len(bearings))

		for i, b := range bearings {
			rad := b.Degree * math.Pi / 180
			// Calculate radius progression
			radius := 190 * float64(maxBearings-i-1) / float64(maxBearings-1)

			data.Bearings[i] = struct {
				Degree  float64
				Time    string
				MsecAgo int64
				X       float64
				Y       float64
				Color   string
				Index   int
			}{
				Degree:  b.Degree,
				Time:    b.Time,
				MsecAgo: time.Since(b.Timestamp).Milliseconds(),
				X:       radius * math.Sin(rad),
				Y:       -radius * math.Cos(rad),
				Color:   fmt.Sprintf("hsl(%d, 100%%, 50%%)", (len(bearings)-i-1)*30),
				Index:   i,
			}
		}

	}

	funcMap := template.FuncMap{
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"div": func(a, b float64) float64 {
			return a / b
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
		"cos": math.Cos,
		"sin": math.Sin,
		"mod": func(a, b int) int {
			return a % b
		},
		"toFloat64": func(i int) float64 {
			return float64(i)
		},
		"toInt": func(i float64) int {
			return int(i)
		},
	}

	t := template.Must(template.New("compass").Funcs(funcMap).Parse(tmpl))
	err := t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func generateSVGAndTableHTML() string {
	mu.Lock()
	defer mu.Unlock()

	const partialTmpl = `
		<svg width="400" height="400" viewBox="-200 -200 400 400">
			<circle cx="0" cy="0" r="190" fill="none" stroke="#ccc" stroke-width="1"/>

			<!-- Degree ticks and labels -->
			{{range $i := seq 0 23}}
				{{$angle := mul (toFloat64 $i) 15}}
				{{$radians := div (mul (sub 90 $angle) 3.14159265) 180}}
				{{$x1 := mul (cos $radians) 190}}
				{{$y1 := mul (sin $radians) -190}}
				{{$x2 := mul (cos $radians) 180}}
				{{$y2 := mul (sin $radians) -180}}
				<line x1="{{$x1}}" y1="{{$y1}}" x2="{{$x2}}" y2="{{$y2}}" stroke="black" stroke-width="0.5"/>

				{{if eq (mod $i 6) 0}}
					{{$labelX := mul (cos $radians) 165}}
					{{$labelY := mul (sin $radians) -165}}
					<text x="{{$labelX}}" y="{{$labelY}}" text-anchor="middle" dominant-baseline="middle" font-size="12">{{$angle}}°</text>
				{{end}}
			{{end}}

			{{$expiry := .Expiry}}
			{{range .Bearings}}
				{{ if gt .MsecAgo $expiry }}
					<circle cx="{{.X}}" cy="{{.Y}}" r="5" fill="none" stroke="black" stroke-width="1"/>
				{{ else }}
					<circle cx="{{.X}}" cy="{{.Y}}" r="5" fill="{{.Color}}"/>
				{{ end }}
			{{end}}
		</svg>

		<table>
			<tr><th>Degree</th><th>Time</th></tr>
			{{range .Bearings}}
			<tr>
				<td>{{printf "%.1f" .Degree}}°</td>
				<td>{{.Time}}</td>
			</tr>
			{{end}}
		</table>
`

	data := struct {
		Bearings []struct {
			Degree  float64
			Time    string
			MsecAgo int64
			X       float64
			Y       float64
			Color   string
			Index   int
		}
		Refresh int // Refresh interval in seconds
		Expiry  int // Expiry interval in milliseconds
	}{
		Refresh: refreshInterval,
		Expiry:  expireInterval,
	}

	if len(bearings) > 0 {
		data.Bearings = make([]struct {
			Degree  float64
			Time    string
			MsecAgo int64
			X       float64
			Y       float64
			Color   string
			Index   int
		}, len(bearings))

		for i, b := range bearings {
			rad := b.Degree * math.Pi / 180
			// Calculate radius progression
			radius := 190 * float64(maxBearings-i-1) / float64(maxBearings-1)

			data.Bearings[i] = struct {
				Degree  float64
				Time    string
				MsecAgo int64
				X       float64
				Y       float64
				Color   string
				Index   int
			}{
				Degree:  b.Degree,
				Time:    b.Time,
				MsecAgo: time.Since(b.Timestamp).Milliseconds(),
				X:       radius * math.Sin(rad),
				Y:       -radius * math.Cos(rad),
				Color:   fmt.Sprintf("hsl(%d, 100%%, 50%%)", (len(bearings)-i-1)*30),
				Index:   i,
			}
		}

	}

	funcMap := template.FuncMap{
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"div": func(a, b float64) float64 {
			return a / b
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
		"cos": math.Cos,
		"sin": math.Sin,
		"mod": func(a, b int) int {
			return a % b
		},
		"toFloat64": func(i int) float64 {
			return float64(i)
		},
		"toInt": func(i float64) int {
			return int(i)
		},
	}

	buf := new(bytes.Buffer)
	t := template.Must(template.New("compass").Funcs(funcMap).Parse(partialTmpl))
	err := t.Execute(buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
