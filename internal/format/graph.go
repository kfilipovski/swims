package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Point struct {
	Label string  // date or label for X axis
	Value float64 // time in seconds
	Display string // formatted swim time
}

// TimeToSeconds converts "1:05.43" or "59.83" to seconds.
func TimeToSeconds(t string) float64 {
	t = strings.TrimSuffix(t, "r") // relay leadoff marker
	parts := strings.SplitN(t, ":", 2)
	if len(parts) == 2 {
		mins, _ := strconv.ParseFloat(parts[0], 64)
		secs, _ := strconv.ParseFloat(parts[1], 64)
		return mins*60 + secs
	}
	secs, _ := strconv.ParseFloat(t, 64)
	return secs
}

// Graph renders an ASCII chart. If invertY is true, higher values are at the top
// (good for times where slower = higher). If false, higher values are at the top
// (good for points where more = better).
func Graph(points []Point, width, height int, invertY bool) string {
	if len(points) == 0 {
		return ""
	}

	// Find min/max values
	minVal, maxVal := points[0].Value, points[0].Value
	for _, p := range points {
		if p.Value < minVal {
			minVal = p.Value
		}
		if p.Value > maxVal {
			maxVal = p.Value
		}
	}

	// Add some padding to range
	valRange := maxVal - minVal
	if valRange == 0 {
		valRange = 1
	}
	minVal -= valRange * 0.05
	maxVal += valRange * 0.05
	valRange = maxVal - minVal

	// Y-axis label width
	yLabelWidth := 8

	// Build the grid (y=0 is top, which is max time / slowest)
	grid := make([][]byte, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]byte, width)
		for x := range grid[y] {
			grid[y][x] = ' '
		}
	}

	// Plot points
	plotWidth := width - yLabelWidth - 1
	for i, p := range points {
		x := yLabelWidth + 1 + int(math.Round(float64(i)*float64(plotWidth-1)/float64(max(len(points)-1, 1))))
		// Higher value at top of chart
		y := int(math.Round(float64(height-1) * (maxVal - p.Value) / valRange))
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}
		if x < len(grid[y]) {
			grid[y][x] = '*'
		}
	}

	var sb strings.Builder

	// Render with Y-axis labels
	fmtLabel := formatSeconds
	if !invertY {
		fmtLabel = formatPlain
	}
	for y := 0; y < height; y++ {
		val := maxVal - float64(y)*valRange/float64(height-1)
		if y == 0 || y == height-1 || y == height/2 {
			sb.WriteString(fmt.Sprintf("%7s ", fmtLabel(val)))
		} else {
			sb.WriteString(strings.Repeat(" ", yLabelWidth) + " ")
		}
		// Y axis line
		sb.WriteByte('|')
		sb.Write(grid[y][yLabelWidth+1:])
		sb.WriteByte('\n')
	}

	// X axis
	sb.WriteString(strings.Repeat(" ", yLabelWidth) + " +")
	sb.WriteString(strings.Repeat("-", plotWidth))
	sb.WriteByte('\n')

	// X axis labels (first and last date)
	if len(points) > 0 {
		first := points[0].Label
		last := points[len(points)-1].Label
		gap := plotWidth - len(first) - len(last)
		if gap < 1 {
			gap = 1
		}
		sb.WriteString(strings.Repeat(" ", yLabelWidth+2))
		sb.WriteString(first)
		sb.WriteString(strings.Repeat(" ", gap))
		sb.WriteString(last)
		sb.WriteByte('\n')
	}

	return sb.String()
}

func formatSeconds(s float64) string {
	if s >= 60 {
		mins := int(s) / 60
		secs := s - float64(mins*60)
		return fmt.Sprintf("%d:%05.2f", mins, secs)
	}
	return fmt.Sprintf("%.2f", s)
}

func formatPlain(v float64) string {
	return fmt.Sprintf("%.0f", v)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
