package main

import "testing"

func TestMathMin(t *testing.T) {
	if mathMin(1.0, 2.0) != 1.0 {
		t.Fatalf("mathMin returned wrong value")
	}
}
