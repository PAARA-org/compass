/* Copyright (c) 2025, kn6yuh@gmail.com, PAARA.org
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package main

const pageHeader = `<!DOCTYPE html>
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
    <title>Radio Direction Finder</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; }
        svg { margin: 15px auto; display: block; }
        table { margin: auto; border-collapse: collapse; }
        td, th { padding: 8px; border: 1px solid #ddd; }
    </style>
</head>
<body>
    <h1>Radio Direction Finder</h1>
	<div id="svgContainer">
`

const svgTemplate = `
		<svg width="400" height="400" viewBox="-200 -200 400 400">
			<circle cx="0" cy="0" r="190" fill="#F2F2F2" stroke="#ccc" stroke-width="1.5"/>

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
					<circle cx="{{.X}}" cy="{{.Y}}" r="5" fill="none" stroke="grey" stroke-width="0.7"/>
				{{ else }}
					<circle cx="{{.X}}" cy="{{.Y}}" r="5" fill="{{.Color}}" stroke="black" stroke-width="0.08"/>
				{{ end }}
			{{end}}
		</svg>
	</div>`

// Here we're printing the legend which needs to start at 75 pixels offset
// in order to be properly aligned with the compass above which is 400 pixels wide
const midTemplate = `
		<svg width="400" height="20" xmlns="http://www.w3.org/2000/svg">
		{{- range $i := seq 0 1000 }}
			{{ if eq (mod $i 200) 0 }}
				<text x="{{ add (divInt $i 4) 70 }}" y="15" text-anchor="middle" dominant-baseline="middle" font-size="12">{{ $i }}</text>
			{{ end }}
		{{- end }}
		</svg>
		<svg width="400" height="15" xmlns="http://www.w3.org/2000/svg">
		{{- range $i := seq 0 250 }}
			<line x1="{{ add $i 75 }}" y1="0" x2="{{ add $i 75 }}" y2="10" stroke="{{ valueToColor (mulInt $i 4) }}" stroke-width="1" />
		{{- end }}
		</svg>

	<h4>Recent Bearings</h4>
`

const tableTemplate = `
	<div id="tableContainer">
		<table>
			<tr><th>Degree</th><th>Magnitude</th><th>Time</th></tr>
			{{$maxrows := .MaxRows}}
			{{range $index, $element := .Bearings}}
				{{if lt $index $maxrows}}
					<tr>
						<td>{{printf "%.1f" .Degree}}°</td>
						<td>{{printf "%d" .Magnitude}}</td>
						<td>{{.Time}}</td>
					</tr>
				{{else}}
					{{break}}
				{{end}}
			{{end}}
		</table>`

const footerTemplate = `
	</div>
</body>
</html>`
