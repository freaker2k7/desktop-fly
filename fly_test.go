package main

import (
	"math"
	"testing"
)

func TestFlySensorsAndUpdate(t *testing.T) {
	f := NewFly(800, 600)
	// fixed position for deterministic sensors
	f.X = 400
	f.Y = 300
	f.Angle = 0

	s := f.Sensors(500, 300)
	if _, ok := s["mouse"]; !ok {
		t.Fatal("missing mouse sensor")
	}
	if s["left"] < 0 || s["right"] < 0 || s["brightness"] < 0 {
		t.Fatal("unexpected negative sensor values")
	}

	// apply thrust and update; ensure movement and bounding
	f.Thrust = 1.0
	f.Turn = 0.5
	prevX, prevY := f.X, f.Y
	f.Update(0.016, 10.0)
	if math.IsNaN(f.X) || math.IsNaN(f.Y) {
		t.Fatalf("invalid position after update: %v,%v", f.X, f.Y)
	}
	if f.X == prevX && f.Y == prevY {
		t.Fatalf("expected fly to move; pos unchanged")
	}
}
