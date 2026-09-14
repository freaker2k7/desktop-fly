package main

import "fmt"

// Plot is a placeholder for the hvplot-based Python Plot. In Go we
// currently implement a no-op that could be extended to write a simple
// HTML/JS viewer or communicate with a separate visualization process.
type Plot struct {
	Brain *Brain
}

func NewPlot(brain *Brain) *Plot {
	return &Plot{Brain: brain}
}

func (p *Plot) Show(autoOpen bool) {
	// Placeholder; in future this could write a minimal HTML page
	// and open it in the browser.
	if autoOpen {
		fmt.Println("Plot: autoOpen requested but not implemented in Go yet")
	}
}

func (p *Plot) Update(triggered []int64) {
	// no-op for now
}
