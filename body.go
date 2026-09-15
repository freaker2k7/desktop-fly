package main

import "fmt"

// Body holds a single unified neural `Brain` that contains neurons and
// skeletons merged from multiple CSV exports (root, brain, eyes, wings).
type Body struct {
	Brain *Brain
}

// NewBody loads three neural modules from CSVs using configured names.
func NewBody(dataDir string) (*Body, error) {
	roots := []string{RootNeuron, BrainNeuron, EyesNeuron, WingsNeuron}
	mainBrain, err := LoadCombinedBrain(roots, dataDir, 0, 0)
	if err != nil {
		return nil, err
	}
	return &Body{Brain: mainBrain}, nil
}

// Merged returns a view combining neurons and edges from all modules.
// It does not deep-copy neurons; it creates a new Brain wrapper that
// references the existing neuron pointers so activity updates are visible.
func (b *Body) Merged() *Brain {
	// The body already holds a unified brain; return it directly.
	return b.Brain
}

// Step advances each module and returns aggregated (turn, thrust).
// We call Step on each module and average their outputs.
func (b *Body) Step(sensors map[string]float64) (float64, float64) {
	if b.Brain == nil {
		return 0.0, 0.0
	}
	return b.Brain.Step(sensors)
}

// mergeBrain copies neurons and edges from src into dst. Existing
// neurons in dst are preserved; duplicate edges (same src/dst) are
// deduplicated.
func mergeBrain(dst *Brain, src *Brain) {
	if dst == nil || src == nil {
		return
	}
	for id, n := range src.Neurons {
		if _, ok := dst.Neurons[id]; !ok {
			dst.Neurons[id] = n
		}
	}
	seen := make(map[string]bool)
	for _, e := range dst.Edges {
		key := fmt.Sprintf("%d-%d", e.Src, e.Dst)
		seen[key] = true
	}
	for _, e := range src.Edges {
		key := fmt.Sprintf("%d-%d", e.Src, e.Dst)
		if seen[key] {
			continue
		}
		// only add edges where both neurons are present in dst
		if _, ok := dst.Neurons[e.Src]; !ok {
			continue
		}
		if _, ok := dst.Neurons[e.Dst]; !ok {
			continue
		}
		dst.Edges = append(dst.Edges, e)
		seen[key] = true
	}
}
