package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkeletonSegments(t *testing.T) {
	dir := t.TempDir()
	root := "r"
	fname := filepath.Join(dir, root+"_skeleton_99.csv")
	// header: id,x,z,link
	data := "id,x,z,link\n1,10,20,\n2,15,25,1\n"
	os.WriteFile(fname, []byte(data), 0644)

	segs, nodes, ids, minX, _, maxX, _, err := LoadSkeletonSegments(root, dir)
	if err != nil {
		t.Fatalf("LoadSkeletonSegments error: %v", err)
	}
	if len(segs) == 0 {
		t.Fatalf("expected segments, got none")
	}
	if len(nodes) == 0 || len(ids) == 0 {
		t.Fatalf("expected node centroids and ids")
	}
	if minX == 0 && maxX == 0 {
		t.Fatalf("bbox seems invalid: %v %v", minX, maxX)
	}
}
