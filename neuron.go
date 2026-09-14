package main

type Neuron struct {
	BodyID    int64
	Name      string
	Potential float64
	Activity  float64
	ROI       string
}

func NewNeuron(id int64, name string) *Neuron {
	return &Neuron{BodyID: id, Name: name}
}
