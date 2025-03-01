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
`

const svgTemplate = `
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
		</svg>`

const midTemplate = `
	</div>

	<h2>Recent Bearings</h2>
	<div id="tableContainer">`

const tableTemplate = `
		<table>
			<tr><th>Degree</th><th>Time</th></tr>
			{{range .Bearings}}
			<tr>
				<td>{{printf "%.1f" .Degree}}°</td>
				<td>{{.Time}}</td>
			</tr>
			{{end}}
		</table>`

const footerTemplate = `
	</div>
</body>
</html>`
