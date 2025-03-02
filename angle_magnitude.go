/* Copyright (c) 2025, kn6yuh@gmail.com, PAARA.org
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package main

import "fmt"

// This function takes the angle vector average magnitude
// and converts the values to colors in a gradient from white to blue
// Andy (KR6DD) suggested this color schema as seen in waterfalls
// of SDR# and Spectran.exe (those run only on PC's).
func magnitudeToColor(value int) string {
	if value < 0 {
		value = 0
	} else if value > 999 {
		value = 999
	}

	t := float64(999-value) / 999.0
	type colorStop struct {
		t       float64
		r, g, b uint8
	}

	stops := []colorStop{
		{0.0, 255, 255, 255}, // White (max strength)
		{0.15, 255, 0, 0},    // Red
		{0.35, 255, 255, 0},  // Yellow
		{0.55, 0, 255, 0},    // Green
		{0.75, 0, 255, 255},  // Cyan
		{1.0, 0, 0, 255},     // Blue (min strength)
	}

	var prev, next *colorStop
	for i := 0; i < len(stops)-1; i++ {
		if t >= stops[i].t && t <= stops[i+1].t {
			prev, next = &stops[i], &stops[i+1]
			break
		}
	}

	if prev == nil {
		if t <= stops[0].t {
			return fmt.Sprintf("#%02X%02X%02X", stops[0].r, stops[0].g, stops[0].b)
		}
		return fmt.Sprintf("#%02X%02X%02X", stops[len(stops)-1].r, stops[len(stops)-1].g, stops[len(stops)-1].b)
	}

	delta := next.t - prev.t
	frac := (t - prev.t) / delta
	r := prev.r + uint8(frac*float64(next.r-prev.r))
	g := prev.g + uint8(frac*float64(next.g-prev.g))
	b := prev.b + uint8(frac*float64(next.b-prev.b))

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
