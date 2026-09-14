package main

import (
	"math"
)

type Edge struct {
	Src    int64
	Dst    int64
	Weight float64
	ROI    string
}

type Brain struct {
	Neurons map[int64]*Neuron
	Edges   []Edge
}

func NewBrain() *Brain {
	return &Brain{Neurons: make(map[int64]*Neuron)}
}

func (b *Brain) AddNeuron(id int64, name string) {
	if _, ok := b.Neurons[id]; !ok {
		b.Neurons[id] = NewNeuron(id, name)
	}
}

func (b *Brain) Step(sensory map[string]float64) (float64, float64) {
	// copy neurons to slice
	neurons := make([]*Neuron, 0, len(b.Neurons))
	for _, n := range b.Neurons {
		neurons = append(neurons, n)
	}
	if len(neurons) == 0 {
		return 0.0, 0.0
	}

	left := sensory["left"]
	right := sensory["right"]
	mouse := sensory["mouse"]
	brightness := sensory["brightness"]

	for i := 0; i < len(neurons) && i < 8; i++ {
		switch i % 4 {
		case 0:
			neurons[i].Potential += left * 0.35
		case 1:
			neurons[i].Potential += right * 0.35
		case 2:
			neurons[i].Potential += mouse * 0.5
		case 3:
			neurons[i].Potential += brightness * 0.15
		}
	}

	spikes := make(map[int64]struct{})
	for _, neuron := range neurons {
		neuron.Potential *= 0.92
		if neuron.Potential >= 1.0 {
			neuron.Potential = 0.0
			neuron.Activity = 1.0
			spikes[neuron.BodyID] = struct{}{}
		} else {
			neuron.Activity *= 0.75
		}
	}

	for _, e := range b.Edges {
		if _, ok := spikes[e.Src]; ok {
			if target, ok := b.Neurons[e.Dst]; ok {
				var mult float64
				// approximate ROI-based modifiers using presence
				if e.ROI != "" && (b.Neurons[e.Src].ROI == e.ROI || target.ROI == e.ROI) {
					mult = ROI_SAME_BOOST
				} else if b.Neurons[e.Src].ROI == "" && target.ROI == "" {
					mult = ROI_UNKNOWN
				} else {
					mult = ROI_DIFF_PENALTY
				}
				target.Potential += e.Weight * mult
			}
		}
	}

	turn := 0.0
	thrust := 0.0
	for i, neuron := range neurons {
		if i%3 == 0 {
			turn -= neuron.Activity
		} else if i%3 == 1 {
			turn += neuron.Activity
		} else {
			thrust += neuron.Activity
		}
	}

	scale := math.Max(1, float64(len(neurons)/8))

	if scale == 0 {
		scale = 1
	}

	t := math.Max(-1.0, math.Min(1.0, turn/scale))
	th := math.Max(0.0, math.Min(1.0, thrust/scale))
	return t, th
}
