package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LoadBrain attempts to read CSV exports created by the Python script
// in `DataDir`. It looks for files named `<root>_neurons.csv` and
// `<root>_connections.csv`. This is a best-effort loader and will skip
// malformed rows.
func LoadBrain(root string, dataDir string, maxNeurons int, maxConnections int) (*Brain, error) {
	b := NewBrain()
	neuronsFile := filepath.Join(dataDir, fmt.Sprintf("%s_neurons.csv", root))
	connsFile := filepath.Join(dataDir, fmt.Sprintf("%s_connections.csv", root))

	if _, err := os.Stat(neuronsFile); err == nil {
		if err := loadNeuronsCSV(b, neuronsFile, maxNeurons); err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(connsFile); err == nil {
		if err := loadConnsCSV(b, connsFile, maxConnections); err != nil {
			return nil, err
		}
	}

	if len(b.Neurons) == 0 {
		return nil, fmt.Errorf("no neurons loaded; ensure you ran scripts/get_neurons.py")
	}

	return b, nil
}

func loadNeuronsCSV(b *Brain, path string, maxNeurons int) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	r := csv.NewReader(fh)
	headers, err := r.Read()
	if err != nil {
		return err
	}
	idx := map[string]int{}
	for i, h := range headers {
		idx[h] = i
	}
	var count int
	for {
		if maxNeurons > 0 && count >= maxNeurons {
			break
		}
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		var bodyID int64
		if i, ok := idx["bodyId"]; ok {
			bodyID, _ = strconv.ParseInt(row[i], 10, 64)
		} else if i, ok := idx["bodyId_pre"]; ok {
			bodyID, _ = strconv.ParseInt(row[i], 10, 64)
		} else {
			// try to find any column name that contains 'body'
			for k, i := range idx {
				if len(k) >= 4 && (k == "body" || containsIgnoreCase(k, "body")) {
					bodyID, _ = strconv.ParseInt(row[i], 10, 64)
					break
				}
			}
		}
		if bodyID == 0 {
			continue
		}
		name := ""
		if i, ok := idx["instance"]; ok {
			name = row[i]
		} else if i, ok := idx["type"]; ok {
			name = row[i]
		} else {
			name = fmt.Sprintf("%d", bodyID)
		}
		b.AddNeuron(bodyID, name)
		count++
	}
	return nil
}

func loadConnsCSV(b *Brain, path string, maxConnections int) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	r := csv.NewReader(fh)
	headers, err := r.Read()
	if err != nil {
		return err
	}
	idx := map[string]int{}
	for i, h := range headers {
		idx[h] = i
	}
	var count int
	for {
		if maxConnections > 0 && count >= maxConnections {
			break
		}
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		var src, dst int64
		if i, ok := idx["bodyId_pre"]; ok {
			src, _ = strconv.ParseInt(row[i], 10, 64)
		}
		if i, ok := idx["bodyId_post"]; ok {
			dst, _ = strconv.ParseInt(row[i], 10, 64)
		}
		if src == 0 || dst == 0 {
			// try to find two 'body' columns
			var found []int
			for k, i := range idx {
				if containsIgnoreCase(k, "body") {
					found = append(found, i)
				}
			}
			if len(found) >= 2 {
				src, _ = strconv.ParseInt(row[found[0]], 10, 64)
				dst, _ = strconv.ParseInt(row[found[1]], 10, 64)
			} else {
				continue
			}
		}
		weight := 1.0
		if i, ok := idx["weight"]; ok {
			if w, err := strconv.ParseFloat(row[i], 64); err == nil {
				weight = w
			}
		}
		// normalize similar to Python loader
		w := mathLog1p(weight) / 20.0
		if w > 0.25 {
			w = 0.25
		}
		roi := ""
		if i, ok := idx["roi"]; ok {
			roi = row[i]
		}
		if _, ok := b.Neurons[src]; !ok {
			continue
		}
		if _, ok := b.Neurons[dst]; !ok {
			continue
		}
		b.Edges = append(b.Edges, Edge{Src: src, Dst: dst, Weight: w, ROI: roi})
		count++
	}
	return nil
}

func containsIgnoreCase(s, sub string) bool {
	return stringsContainsFold(s, sub)
}

// simple helpers (isolate dependencies)
func mathLog1p(x float64) float64 {
	return math.Log1p(x)
}

// wrapper for strings.ContainsFold to avoid Go version issues
func stringsContainsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
