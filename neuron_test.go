package main

import "testing"

func TestNewNeuron(t *testing.T) {
	n := NewNeuron(42, "foo")
	if n.BodyID != 42 {
		t.Fatalf("expected BodyID 42, got %d", n.BodyID)
	}
	if n.Name != "foo" {
		t.Fatalf("expected Name foo, got %s", n.Name)
	}
}
