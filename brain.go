package main

import (
	"math"
)

type Edge struct {
	Src    int64
	Dst    int64
	Weight float64
	ROI    string
	Active bool
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
	view := sensory["view"]

	// for i := 0; i < len(neurons) && i < 8; i++ {
	for i := range neurons {
		switch i % 4 {
		case 0:
			neurons[i].Potential += left * 0.35
		case 1:
			neurons[i].Potential += right * 0.35
		case 2:
			// prefer vibrant "view" signals to encourage resting on
			// colorful regions
			neurons[i].Potential += view * 0.6
		case 3:
			// mouse still provides urgency but is secondary to view
			neurons[i].Potential += mouse * 0.5
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

	for i := range b.Edges {
		e := &b.Edges[i]
		if _, ok := spikes[e.Src]; ok {
			if target, ok := b.Neurons[e.Dst]; ok {
				var mult float64
				if e.ROI != "" && (b.Neurons[e.Src].ROI == e.ROI || target.ROI == e.ROI) {
					mult = ROI_SAME_BOOST
				} else if b.Neurons[e.Src].ROI == "" && target.ROI == "" {
					mult = ROI_UNKNOWN
				} else {
					mult = ROI_DIFF_PENALTY
				}
				target.Potential += e.Weight * mult
				// fmt.Println("Activating edge from", e.Src, "to", e.Dst, "with weight", e.Weight)
				e.Active = true
			} else {
				e.Active = false
			}
		} else {
			e.Active = false
		}
	}

	turn := 0.0
	thrust := 0.0
	for i, neuron := range neurons {
		switch i % 3 {
		case 0:
			turn -= neuron.Activity
		case 1:
			turn += neuron.Activity
		case 2:
			thrust += neuron.Activity
		}
	}

	scale := math.Max(1, float64(len(neurons)/8))

	if scale == 0 {
		scale = 1
	}

	t := math.Max(-1.0, math.Min(1.0, turn/scale))
	th := math.Max(0.0, math.Min(1.0, thrust/scale))

	// Behavior overrides: if mouse is near, run away (strong thrust
	// and steer away from mouse). Otherwise, prefer to rest on vibrant
	// colors by reducing thrust proportional to view.
	if mouse > 0.6 {
		// steer away: use right-left difference as direction
		dir := right - left
		t = math.Max(-1.0, math.Min(1.0, dir*1.5))
		th = math.Max(th, math.Min(1.0, 0.9+0.1*mouse))
	} else {
		// reduce thrust when vibrant colors are present to encourage
		// resting on them (view in [0..1])
		th = th * (1.0 - 0.7*view)
	}

	return t, th
}
