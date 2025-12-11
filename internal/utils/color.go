package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
)

// GenerateUserColor generates a consistent, visually pleasant color for a user
// based on their user ID. The color is deterministic - the same ID always produces
// the same color. Colors are generated in HSL space for better visual distribution.
func GenerateUserColor(userID string) string {
	// Create MD5 hash of the user ID for consistent color generation
	hash := md5.Sum([]byte(userID))
	hashHex := hex.EncodeToString(hash[:])

	// Use different parts of the hash for different color components
	// This ensures good distribution across the color spectrum

	// Hue: Use first 4 chars of hash (0-65535) and map to 0-360 degrees
	// We'll favor colors that are more distinguishable (avoid yellow-green range)
	hueHash := hexToInt(hashHex[0:4])
	hue := float64(hueHash%360) // Full color wheel

	// Saturation: 60-80% for vibrant but not oversaturated colors
	satHash := hexToInt(hashHex[4:6])
	saturation := 60.0 + float64(satHash%21) // 60-80%

	// Lightness: 45-65% for colors that work on both light and dark backgrounds
	lightHash := hexToInt(hashHex[6:8])
	lightness := 45.0 + float64(lightHash%21) // 45-65%

	// Convert HSL to RGB
	r, g, b := hslToRgb(hue, saturation, lightness)

	// Return as hex color
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// hexToInt converts a hex string to an integer
func hexToInt(hexStr string) int {
	var result int
	fmt.Sscanf(hexStr, "%x", &result)
	return result
}

// hslToRgb converts HSL color values to RGB
// h: 0-360, s: 0-100, l: 0-100
// returns r, g, b: 0-255
func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	// Normalize values
	h = h / 360.0
	s = s / 100.0
	l = l / 100.0

	var r, g, b float64

	if s == 0 {
		// Achromatic (grey)
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRgb(p, q, h+1.0/3.0)
		g = hueToRgb(p, q, h)
		b = hueToRgb(p, q, h-1.0/3.0)
	}

	return uint8(math.Round(r * 255)),
		uint8(math.Round(g * 255)),
		uint8(math.Round(b * 255))
}

// hueToRgb is a helper function for HSL to RGB conversion
func hueToRgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

