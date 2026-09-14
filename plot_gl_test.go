package main

import "testing"

// func TestPlotWindowUpdateNodeColorsNoVBO(t *testing.T) {
// 	pw := &PlotWindow{}
// 	// NodeVBO == 0 should cause UpdateNodeColors to return quickly
// 	br := NewBrain()
// 	pw.UpdateNodeColors(br)
// }

func TestPlotWindowCloseNil(t *testing.T) {
	pw := &PlotWindow{}
	pw.Close() // should not panic when Window is nil
}
