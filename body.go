package main

// Body aggregates multiple neural modules: brain, eyes and wings.
type Body struct {
	Brain *Brain
	Eyes  *Brain
	Wings *Brain
}

// NewBody loads three neural modules from CSVs using configured names.
func NewBody(dataDir string) (*Body, error) {
	br, err := LoadBrain(RootNeuron, dataDir, 0, 0)
	if err != nil {
		return nil, err
	}
	eyes, err := LoadBrain(EyesNeuron, dataDir, 0, 0)
	if err != nil {
		return nil, err
	}
	wings, err := LoadBrain(WingsNeuron, dataDir, 0, 0)
	if err != nil {
		return nil, err
	}
	return &Body{Brain: br, Eyes: eyes, Wings: wings}, nil
}

// Merged returns a view combining neurons and edges from all modules.
// It does not deep-copy neurons; it creates a new Brain wrapper that
// references the existing neuron pointers so activity updates are visible.
func (b *Body) Merged() *Brain {
	merged := NewBrain()
	// copy pointers from each module
	for _, n := range b.Brain.Neurons {
		merged.Neurons[n.BodyID] = n
	}
	for _, n := range b.Eyes.Neurons {
		merged.Neurons[n.BodyID] = n
	}
	for _, n := range b.Wings.Neurons {
		merged.Neurons[n.BodyID] = n
	}
	// combine edges (keep as-is)
	merged.Edges = append(merged.Edges, b.Brain.Edges...)
	merged.Edges = append(merged.Edges, b.Eyes.Edges...)
	merged.Edges = append(merged.Edges, b.Wings.Edges...)
	return merged
}

// Step advances each module and returns aggregated (turn, thrust).
// We call Step on each module and average their outputs.
func (b *Body) Step(sensors map[string]float64) (float64, float64) {
	t1, th1 := 0.0, 0.0
	if b.Brain != nil {
		t1, th1 = b.Brain.Step(sensors)
	}
	t2, th2 := 0.0, 0.0
	if b.Eyes != nil {
		t2, th2 = b.Eyes.Step(sensors)
	}
	t3, th3 := 0.0, 0.0
	if b.Wings != nil {
		t3, th3 = b.Wings.Step(sensors)
	}
	// simple average across modules
	turn := (t1 + t2 + t3) / 3.0
	thrust := (th1 + th2 + th3) / 3.0
	return turn, thrust
}
