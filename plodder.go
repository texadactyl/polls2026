package main

import (
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// VariableConfig pairs our metrics with their metadata for clean looping
type VariableConfig struct {
	Name   string
	Color  color.RGBA
	Values []float64
}

func plodder() {
	pngPath := "plodder2.png"
	p := plot.New()

	p.Title.Text = "Multi-Variable Performance Matrix"
	p.X.Label.Text = "Timeline"
	p.Y.Label.Text = "Performance Metrics (Same Scale)"

	// Establish the 15 chronological dates (Jan 1, 2026 to Oct 22, 2026)
	startDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC)

	totalDuration := endDate.Sub(startDate)
	stepInterval := totalDuration / 14 // Splits time into exactly 15 sequential steps

	timestamps := make([]float64, 15)
	for i := range 15 {
		currentDate := startDate.Add(stepInterval * time.Duration(i))
		timestamps[i] = float64(currentDate.Unix())
	}

	// Define your 4 dependent variables with distinct, professional colors
	variables := []VariableConfig{
		{
			Name:   "Variable Alpha",
			Color:  color.RGBA{R: 0, G: 122, B: 255, A: 255}, // Electric Blue
			Values: []float64{20, 24, 25, 31, 38, 42, 45, 51, 55, 60, 68, 71, 78, 82, 89},
		},
		{
			Name:   "Variable Beta",
			Color:  color.RGBA{R: 255, G: 45, B: 85, A: 255}, // System Pink/Red
			Values: []float64{45, 31, 62, 55, 41, 78, 52, 69, 88, 61, 74, 92, 65, 81, 95},
		},
		{
			Name:   "Variable Gamma",
			Color:  color.RGBA{R: 40, G: 167, B: 69, A: 255}, // Forest Green
			Values: []float64{60, 58, 55, 59, 64, 62, 70, 68, 73, 71, 80, 78, 85, 83, 90},
		},
		{
			Name:   "Variable Delta",
			Color:  color.RGBA{R: 111, G: 66, B: 193, A: 255}, // Deep Purple
			Values: []float64{10, 15, 12, 18, 22, 28, 24, 35, 30, 42, 39, 48, 44, 55, 51},
		},
	}

	// Standalone legend, drawn only in the external legend area.
	// p.Legend stays empty, so p.Draw() paints no legend inside the plot.
	extLegend := plot.NewLegend()
	extLegend.TextStyle = p.Legend.TextStyle // keep the plot's default text styling

	// Loop through your configuration slice to build and inject the lines
	for _, v := range variables {
		pts := make(plotter.XYs, 15)
		for i := range 15 {
			pts[i] = plotter.XY{
				X: timestamps[i],
				Y: v.Values[i],
			}
		}

		line, err := plotter.NewLine(pts)
		if err != nil {
			log.Panic(err)
		}
		line.Color = v.Color
		line.Width = vg.Points(2)

		// Add the line to the plot, and its entry to the external legend only
		p.Add(line)
		extLegend.Add(v.Name, line)
	}

	// Clean formatting to ensure labels are not cut off on bottom or left side
	p.X.Tick.Marker = plot.TimeTicks{
		Format: "2006-01-02",
	}

	p.X.Tick.Label.Rotation = math.Pi / 4
	p.X.Tick.Label.XAlign = draw.XRight
	p.X.Tick.Label.YAlign = draw.YCenter

	p.X.Padding = vg.Points(55) // Keeps the tail edges of dates safely inside the box
	p.Y.Padding = vg.Points(45) // Shifts chart right so the first date text doesn't clip on the left frame

	// =========================================================================
	// PUSH LEGEND COMPLETELY OUTSIDE (Right Side Whitespace Margin)
	// =========================================================================

	// Define the overall image dimensions (Make it wide to hold both pieces)
	width := 11.5 * vg.Inch
	height := 5.5 * vg.Inch

	// Allocate space for your external legend box
	legendWidth := 2.0 * vg.Inch

	// 2. Initialize a blank image backing canvas
	canvas := vgimg.New(width, height)
	baseDrawArea := draw.New(canvas)

	// CROP THE CANVAS INTO TWO SEPARATE ROOMS:
	// Room A: The plot graphic area (Takes the full size minus the legend width on the right)
	plotArea := draw.Crop(baseDrawArea, 0, -legendWidth, 0, 0)

	// Room B: The legend destination box (Occupies just that rightmost slice)
	legendArea := draw.Crop(baseDrawArea, width-legendWidth, 0, 0, 0)

	// 4. Reset the legend anchors so they align cleanly inside Room B
	extLegend.Left = true
	extLegend.Top = true
	extLegend.XOffs = vg.Points(10) // Small margin away from the chart boundary line
	extLegend.YOffs = vg.Points(0)

	// Draw each component explicitly into its own space
	p.Draw(plotArea)           // Paints your 4 lines and rotated axes (no legend)
	extLegend.Draw(legendArea) // Paints the legend key completely separate

	// Write the final visual canvas out to a PNG file
	err := writePNG(pngPath, canvas)
	if err != nil {
		log.Panic(err)
	}

}

// WritePNG - write the PNG file from the canvas.
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
