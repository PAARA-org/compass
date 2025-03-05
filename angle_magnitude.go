/* Copyright (c) 2025, kn6yuh@gmail.com, PAARA.org
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package main

import (
	"fmt"
	"math"
)

// This contains the color pallette definition
type colorStop struct {
	position float64 // Normalized position (0.0 to 1.0)
	r, g, b  int     // RGB values
}

// This function picks the right color depending on whether or not accessible
// colors are being used
func getColor(value int, colors []string, a11y bool) (color string) {
	if a11y {
		return magnitudeToColor(value, colors)
	} else {
		return magnitudeToColorGradient(value)
	}
}

// This function returns the string representation of the hexadecimal
// value of the color for each of the maginitude values
func magnitudeToColor(value int, colors []string) string {
	if value < 200 {
		return colors[0]
	} else if value < 400 {
		return colors[1]
	} else if value < 600 {
		return colors[2]
	} else if value < 800 {
		return colors[3]
	} else {
		return colors[4]
	}
}

// This function takes the angle vector average magnitude
// and converts the values to colors in a gradient from white to blue
// Andy (KR6DD) suggested this color schema as seen in waterfalls
// of SDR# and Spectran.exe (those run only on PC's).
func magnitudeToColorGradient(value int) string {
	if value < 0 {
		value = 0
	} else if value > 999 {
		value = 999
	}

	stops := []colorStop{
		{0.0, 0, 0, 139},      // Dark Blue (min strength)
		{0.16, 173, 216, 230}, // Light Blue
		{0.32, 255, 255, 255}, // White
		{0.48, 255, 255, 0},   // Yellow
		{0.64, 155, 165, 0},   // Orange
		{0.84, 255, 0, 0},     // Red
		{1.0, 139, 69, 19},    // Brown (max strength)
	}

	r, g, b := getRGB(value, stops)

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// Function to get RGB values for a given value between 0 and 999
func getRGB(value int, stops []colorStop) (int, int, int) {
	// Normalize the input value to a range of 0.0 to 1.0
	normalizedValue := float64(value) / 999.0

	// Find the two stops between which the normalized value falls
	for i := 0; i < len(stops)-1; i++ {
		start := stops[i]
		end := stops[i+1]

		if normalizedValue >= start.position && normalizedValue <= end.position {
			// Calculate the interpolation factor (t)
			t := (normalizedValue - start.position) / (end.position - start.position)

			// Linearly interpolate each color channel
			r := int(math.Round(float64(start.r) + t*float64(end.r-start.r)))
			g := int(math.Round(float64(start.g) + t*float64(end.g-start.g)))
			b := int(math.Round(float64(start.b) + t*float64(end.b-start.b)))

			return r, g, b
		}
	}

	// If value is out of range, return the closest stop's color
	if normalizedValue < stops[0].position {
		return stops[0].r, stops[0].g, stops[0].b
	}
	last := stops[len(stops)-1]
	return last.r, last.g, last.b
}
