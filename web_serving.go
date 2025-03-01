package main

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"text/template"
	"time"
)

var funcMap = template.FuncMap{
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

func serveCompass(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	const tmpl = pageHeader + svgTemplate + midTemplate + tableTemplate + footerTemplate

	data := &BT{
		Refresh: refreshInterval,
		Expiry:  expireInterval,
	}

	if len(bearings) > 0 {
		data.Bearings = make([]BearingForTemplate, len(bearings))

		for i, b := range bearings {
			rad := b.Degree * math.Pi / 180
			// Calculate radius progression
			radius := 190 * float64(maxBearings-i-1) / float64(maxBearings-1)

			data.Bearings[i] = BearingForTemplate{
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

	t := template.Must(template.New("compass").Funcs(funcMap).Parse(tmpl))
	err := t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func generateSVGAndTableHTML() string {
	mu.Lock()
	defer mu.Unlock()

	const partialTmpl = svgTemplate + tableTemplate

	data := &BT{
		Refresh: refreshInterval,
		Expiry:  expireInterval,
	}

	if len(bearings) > 0 {
		data.Bearings = make([]BearingForTemplate, len(bearings))

		for i, b := range bearings {
			rad := b.Degree * math.Pi / 180
			// Calculate radius progression
			radius := 190 * float64(maxBearings-i-1) / float64(maxBearings-1)

			data.Bearings[i] = BearingForTemplate{
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

	buf := new(bytes.Buffer)
	t := template.Must(template.New("compass").Funcs(funcMap).Parse(partialTmpl))
	err := t.Execute(buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
