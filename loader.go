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

	// If no edges were found using the standard loader, attempt a fallback
	// that pairs per-synapse rows by (x,y,z) coordinates (some exports
	// provide one row per synapse with a `type` column indicating pre/post).
	if len(b.Edges) == 0 {
		if err := pairConnectionsByCoords(b, root, dataDir); err != nil {
			return nil, err
		}
	}

	if len(b.Neurons) == 0 {
		return nil, fmt.Errorf("no neurons loaded; ensure you ran scripts/get_neurons.py")
	}

	fmt.Printf("LoadBrain: neurons=%d edges=%d\n", len(b.Neurons), len(b.Edges))

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

// pairConnectionsByCoords scans *_connections.csv for files matching root
// and groups rows by their x,y,z coordinates. For each coordinate that has
// both pre and post entries we create edges from each pre body to each
// post body.
func pairConnectionsByCoords(b *Brain, root string, dataDir string) error {
	pattern := filepath.Join(dataDir, fmt.Sprintf("%s_connections.csv", root))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	type row struct {
		body int64
		typ  string
		xstr string
		ystr string
		zstr string
		x    float64
		y    float64
		z    float64
		conf float64
	}
	var allRows []row
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		r := csv.NewReader(fh)
		headers, err := r.Read()
		if err != nil {
			fh.Close()
			continue
		}
		idx := map[string]int{}
		for i, h := range headers {
			idx[h] = i
		}
		for {
			rec, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			var body int64
			if i, ok := idx["bodyId"]; ok {
				body, _ = strconv.ParseInt(rec[i], 10, 64)
			}
			if body == 0 {
				continue
			}
			typ := "post"
			if i, ok := idx["type"]; ok {
				typ = rec[i]
			}
			xstr, ystr, zstr := "", "", ""
			if i, ok := idx["x"]; ok {
				xstr = rec[i]
			}
			if i, ok := idx["y"]; ok {
				ystr = rec[i]
			}
			if i, ok := idx["z"]; ok {
				zstr = rec[i]
			}
			conf := 1.0
			if i, ok := idx["confidence"]; ok {
				if v, err := strconv.ParseFloat(rec[i], 64); err == nil {
					conf = v
				}
			}
			xr, _ := strconv.ParseFloat(xstr, 64)
			yr, _ := strconv.ParseFloat(ystr, 64)
			zr, _ := strconv.ParseFloat(zstr, 64)
			allRows = append(allRows, row{body: body, typ: typ, xstr: xstr, ystr: ystr, zstr: zstr, x: xr, y: yr, z: zr, conf: conf})
		}
		fh.Close()
	}
	// separate pre and post rows
	var pres []row
	var posts []row
	for _, r := range allRows {
		if strings.ToLower(r.typ) == "pre" {
			pres = append(pres, r)
		} else {
			posts = append(posts, r)
		}
	}
	if len(pres) == 0 || len(posts) == 0 {
		return nil
	}
	// for each pre find nearest post within threshold
	usedPost := make([]bool, len(posts))
	seen := make(map[string]bool)
	const threshold = 200.0
	for _, p := range pres {
		bestIdx := -1
		bestDist := 1e12
		for j, q := range posts {
			if usedPost[j] {
				continue
			}
			dx := p.x - q.x
			dy := p.y - q.y
			dz := p.z - q.z
			d := math.Sqrt(dx*dx + dy*dy + dz*dz)
			if d < bestDist {
				bestDist = d
				bestIdx = j
			}
		}
		if bestIdx >= 0 && bestDist <= threshold {
			q := posts[bestIdx]
			// only add edges where both neurons are known
			if _, ok := b.Neurons[p.body]; !ok {
				continue
			}
			if _, ok := b.Neurons[q.body]; !ok {
				continue
			}
			key := fmt.Sprintf("%d-%d", p.body, q.body)
			if seen[key] {
				continue
			}
			seen[key] = true
			usedPost[bestIdx] = true
			w := (p.conf + q.conf) / 2.0
			norm := mathLog1p(w*100.0) / 20.0
			if norm > 0.25 {
				norm = 0.25
			}
			b.Edges = append(b.Edges, Edge{Src: p.body, Dst: q.body, Weight: norm, ROI: ""})
		}
	}
	return nil
}

// simple helpers (isolate dependencies)
func mathLog1p(x float64) float64 {
	return math.Log1p(x)
}

// wrapper for strings.ContainsFold to avoid Go version issues
func stringsContainsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
