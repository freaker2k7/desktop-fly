package main

import "testing"

func TestNewPlotAndUpdate(t *testing.T) {
	b := NewBrain()
	p := NewPlot(b)
	if p.Brain == nil {
		t.Fatalf("expected brain assigned")
	}
	p.Show(false)
	p.Update([]int64{1, 2})
}
