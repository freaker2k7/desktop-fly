package main

import (
	"testing"
)

func TestNewBrainAddNeuronAndStep(t *testing.T) {
	b := NewBrain()
	b.AddNeuron(1, "n1")
	b.AddNeuron(2, "n2")
	if len(b.Neurons) != 2 {
		t.Fatalf("expected 2 neurons, got %d", len(b.Neurons))
	}

	// set potentials high so they spike
	b.Neurons[1].Potential = 2.0
	b.Neurons[2].Potential = 2.0

	// add an edge 1->2 with weight 1.0
	b.Edges = append(b.Edges, Edge{Src: 1, Dst: 2, Weight: 1.0, ROI: ""})

	// run Step with empty sensory
	tval, th := b.Step(map[string]float64{})
	if tval < -1.0 || tval > 1.0 {
		t.Fatalf("turn out of bounds: %v", tval)
	}
	if th < 0.0 || th > 1.0 {
		t.Fatalf("thrust out of bounds: %v", th)
	}

	// After Step, neuron 2 potential should have increased (received from spike)
	// Note: since spikes reset source potential, target potential should be >0
	if b.Neurons[2].Potential <= 0.0 {
		t.Fatalf("expected neuron 2 potential > 0 after propagation, got %v", b.Neurons[2].Potential)
	}
}
