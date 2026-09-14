package main

import (
	"os"
	"strconv"
)

// Configuration with sensible defaults; override using environment vars.
var (
	RootNeuron  = getenv("ROOT_NEURON", "KC")
	EyesNeuron  = getenv("EYES_NEURON", "HBeyelet")
	WingsNeuron = getenv("WINGS_NEURON", "TTMn")
	DataDir     = getenv("DATA_DIR", "data")
	// Brain tick frequency
	BrainHz = getenvInt("BRAIN_HZ", 30)
)

// ROI weighting constants (match Python settings.py defaults)
var (
	ROI_SAME_BOOST   = 1.0
	ROI_DIFF_PENALTY = 0.7
	ROI_UNKNOWN      = 0.85
	// Maximum fly speed
	MAX_SPEED = 7.0
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
