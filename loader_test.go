package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBrainFromCSV(t *testing.T) {
	dir := t.TempDir()
	root := "testroot"
	// write neurons CSV
	nfile := filepath.Join(dir, root+"_neurons.csv")
	os.WriteFile(nfile, []byte("bodyId,instance\n123,neuronA\n456,neuronB\n"), 0644)
	// write connections CSV
	cfile := filepath.Join(dir, root+"_connections.csv")
	os.WriteFile(cfile, []byte("bodyId_pre,bodyId_post,weight,roi\n123,456,2.0,ROI1\n"), 0644)

	b, err := LoadBrain(root, dir, 0, 0)
	if err != nil {
		t.Fatalf("LoadBrain failed: %v", err)
	}
	if len(b.Neurons) != 2 {
		t.Fatalf("expected 2 neurons, got %d", len(b.Neurons))
	}
	if len(b.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(b.Edges))
	}
}
