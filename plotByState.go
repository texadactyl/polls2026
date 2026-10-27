package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// maxXLabels is the most X-axis labels drawn before labels are thinned out.
const maxXLabels = 20

// VariableConfig pairs our metrics with their metadata for clean looping
type VariableConfig struct {
	Name   string
	Color  color.RGBA
	Values []float64
}

// indexTicks labels integer X positions (0, 1, 2, ...) with strings such as "10/03".
// gonum/plot has no string-valued axis, so each point is plotted at its index.
type indexTicks []string

func plotByState(stateCode string, startStamp string) {

	// Error if database is closed.
	if !dbIsOpen {
		fatalText("plotByState: database is closed")
	}

	// Execute query.
	// Rows must arrive in chronological order (ORDER BY the end date in the SQL).
	sqlText := fmt.Sprintf(fmtSqlSelectHistoryByStateAscendingDates, stateCode)
	histories, ok := query(sqlText)
	if !ok {
		fatalText("plotByState: query(%s) failed", sqlText)
	}

	// No histories: print a message and return.
	if len(histories) < 1 {
		fmt.Printf("\tplotByState: No history records were found for state %s\n", stateCode)
		return
	}

	// One history record: print a message and return.
	if len(histories) < 2 {
		fmt.Printf("\tplotByState: Only 1 history record was found for state %s\n", stateCode)
		return
	}

	// Start a new plot.
	plt := plot.New()
	plt.Title.Text = "Trends for State " + stateCode
	plt.X.Label.Text = "Timeline"
	plt.Y.Label.Text = "Percentage"

	// Percentages: keep the scale fixed so it does not jump from state to state.
	plt.Y.Min = 0
	plt.Y.Max = 100

	numPoints := len(histories)
	dateSlice := make([]string, numPoints)
	demSlice := make([]float64, numPoints)
	gopSlice := make([]float64, numPoints)
	indSlice := make([]float64, numPoints)
	unkSlice := make([]float64, numPoints)
	for ix := range numPoints {
		dateSlice[ix] = fmt.Sprintf("%02d/%02d", histories[ix].endMonth, histories[ix].endDay)
		demSlice[ix] = histories[ix].dem
		gopSlice[ix] = histories[ix].gop
		indSlice[ix] = histories[ix].ind
		unkSlice[ix] = histories[ix].unk
	}

	// The 4 dependent variables, each with a distinct color.
	variables := []VariableConfig{
		{Name: "Dem", Color: color.RGBA{R: 0, G: 122, B: 255, A: 255}, Values: demSlice},  // Electric Blue
		{Name: "GOP", Color: color.RGBA{R: 255, G: 45, B: 85, A: 255}, Values: gopSlice},  // System Pink/Red
		{Name: "Ind", Color: color.RGBA{R: 40, G: 167, B: 69, A: 255}, Values: indSlice},  // Forest Green
		{Name: "Unk", Color: color.RGBA{R: 111, G: 66, B: 193, A: 255}, Values: unkSlice}, // Deep Purple
	}

	// Standalone legend, drawn only in the external legend area.
	// plt.Legend stays empty, so plt.Draw() paints no legend inside the plot.
	extLegend := plot.NewLegend()
	extLegend.TextStyle = plt.Legend.TextStyle // keep the plot's default text styling

	// Build one line per variable. X is the point index; the tick labels supply the dates.
	for _, variable := range variables {
		points := make(plotter.XYs, numPoints)
		for ix := range numPoints {
			points[ix] = plotter.XY{
				X: float64(ix),
				Y: variable.Values[ix],
			}
		}

		line, err := plotter.NewLine(points)
		if err != nil {
			fatalErr(err, "plotByState: plotter.NewLine(%s) failed", variable.Name)
		}
		line.Color = variable.Color
		line.Width = vg.Points(2)

		// Add the line to the plot, and its entry to the external legend only
		plt.Add(line)
		extLegend.Add(variable.Name, line)
	}

	// X-axis: string date labels at integer positions, rotated so they do not overlap.
	plt.X.Tick.Marker = indexTicks(dateSlice)
	plt.X.Tick.Label.Rotation = math.Pi / 4
	plt.X.Tick.Label.XAlign = draw.XRight
	plt.X.Tick.Label.YAlign = draw.YCenter

	// Padding is the gap between the data and the axes (not a label margin).
	plt.X.Padding = vg.Points(10)
	plt.Y.Padding = vg.Points(10)

	// =========================================================================
	// Put the legend outside the plot (right-side margin)
	// =========================================================================

	// Overall image size: wide enough to hold both the plot and the legend.
	width := 11.5 * vg.Inch
	height := 5.5 * vg.Inch

	// Space reserved for the external legend box.
	legendWidth := 2.0 * vg.Inch

	// Blank image-backed canvas.
	canvas := vgimg.New(width, height)
	baseDrawArea := draw.New(canvas)

	// Crop the canvas into two areas:
	// plotArea: everything except the legend strip on the right.
	plotArea := draw.Crop(baseDrawArea, 0, -legendWidth, 0, 0)

	// legendArea: just the rightmost strip.
	legendArea := draw.Crop(baseDrawArea, width-legendWidth, 0, 0, 0)

	// Anchor the legend inside the legend area.
	extLegend.Left = true
	extLegend.Top = true
	extLegend.XOffs = vg.Points(10) // Small margin away from the plot boundary
	extLegend.YOffs = vg.Points(0)

	// Draw each component into its own area.
	plt.Draw(plotArea)         // 4 lines and rotated axes (no legend)
	extLegend.Draw(legendArea) // legend key, separate from the plot

	// Make sure the output directory exists.
	const plotsDir = "plots"
	err := os.MkdirAll(plotsDir, 0o755)
	if err != nil {
		fatalErr(err, "plotByState: os.MkdirAll(%s) failed", plotsDir)
	}

	// Absolute-or-relative path for the output PNG file.
	pngPath := filepath.Join(plotsDir, fmt.Sprintf("%s_%s.png", stateCode, startStamp))

	// Write the final canvas out to a PNG file.
	err = writePNG(pngPath, canvas)
	if err != nil {
		fatalErr(err, "plotByState: writePNG(%s) failed", pngPath)
	}

}

// writePNG - write the PNG file from the canvas.
func writePNG(path string, canvas *vgimg.Canvas) (err error) {
	filePNG, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		if errClose := filePNG.Close(); err == nil {
			err = errClose
		}
	}()

	_, err = vgimg.PngCanvas{Canvas: canvas}.WriteTo(filePNG)
	return err
}

// Ticks implements plot.Ticker.
// See gonum.org/v1/plot library.
func (labels indexTicks) Ticks(minValue, maxValue float64) []plot.Tick {
	count := len(labels)
	step := 1
	if count > maxXLabels {
		step = (count + maxXLabels - 1) / maxXLabels
	}
	ticks := make([]plot.Tick, 0, count)
	for ix := 0; ix < count; ix += step {
		value := float64(ix)
		if value < minValue || value > maxValue {
			continue
		}
		ticks = append(ticks, plot.Tick{Value: value, Label: labels[ix]})
	}
	return ticks
}
