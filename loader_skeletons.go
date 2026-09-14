package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// LoadSkeletonSegments scans the data directory for skeleton CSVs named
// "<root>_skeleton_*.csv" and builds a list of 2D segments (x0,y0,x1,y1)
// projected from x,z coordinates. It also computes per-neuron centroids
// and returns node positions and node IDs so the plot can render nodes
// and color them by activity.
func LoadSkeletonSegments(root string, dataDir string) ([]float32, []float32, []int64, float32, float32, float32, float32, error) {
	pattern := filepath.Join(dataDir, fmt.Sprintf("%s_skeleton_*.csv", root))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil, nil, 0, 0, 0, 0, err
	}
	fmt.Printf("LoadSkeletonSegments: pattern=%s files=%d\n", pattern, len(files))
	var segments []float32
	var minX, minY, maxX, maxY float32
	first := true
	var nodePositions []float32
	var nodeIDs []int64

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
		// read rows into map rowId -> (x,z)
		coords := map[int64][2]float64{}
		links := map[int64]int64{}
		var sumX, sumY float64
		var countPts int64
		for {
			row, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			var rowId int64
			if i, ok := idx["rowId"]; ok {
				rowId, _ = strconv.ParseInt(row[i], 10, 64)
			} else if i, ok := idx["id"]; ok {
				rowId, _ = strconv.ParseInt(row[i], 10, 64)
			} else {
				continue
			}
			var x, z float64
			if i, ok := idx["x"]; ok {
				x, _ = strconv.ParseFloat(row[i], 64)
			}
			if i, ok := idx["z"]; ok {
				z, _ = strconv.ParseFloat(row[i], 64)
			}
			coords[rowId] = [2]float64{x, z}
			sumX += x
			sumY += z
			countPts++
			if i, ok := idx["link"]; ok {
				if row[i] != "" {
					v, _ := strconv.ParseInt(row[i], 10, 64)
					links[rowId] = v
				}
			}
		}
		fh.Close()

		// build segments by linking child->parent
		for child, parent := range links {
			pcoords, okp := coords[parent]
			ccoords, okc := coords[child]
			if okp && okc {
				x0 := float32(ccoords[0])
				y0 := float32(ccoords[1])
				x1 := float32(pcoords[0])
				y1 := float32(pcoords[1])
				segments = append(segments, x0, y0, x1, y1)
				if first {
					minX, maxX = x0, x0
					minY, maxY = y0, y0
					first = false
				}
				if x0 < minX {
					minX = x0
				}
				if y0 < minY {
					minY = y0
				}
				if x0 > maxX {
					maxX = x0
				}
				if y0 > maxY {
					maxY = y0
				}
				if x1 < minX {
					minX = x1
				}
				if y1 < minY {
					minY = y1
				}
				if x1 > maxX {
					maxX = x1
				}
				if y1 > maxY {
					maxY = y1
				}
			}
		}

		// centroid for this skeleton file (if any points found)
		if countPts > 0 {
			cx := float32(sumX / float64(countPts))
			cy := float32(sumY / float64(countPts))
			nodePositions = append(nodePositions, cx, cy)
			// try to parse body id from filename suffix
			base := filepath.Base(f)
			// filename like <root>_skeleton_<bodyid>.csv
			var bid int64
			fmt.Sscanf(base, fmt.Sprintf("%s_skeleton_%%d.csv", root), &bid)
			nodeIDs = append(nodeIDs, bid)
		}
	}

	fmt.Printf("LoadSkeletonSegments: nodes=%d segments=%d bbox=(%f,%f)-(%f,%f)\n", len(nodeIDs), len(segments)/4, minX, minY, maxX, maxY)
	return segments, nodePositions, nodeIDs, minX, minY, maxX, maxY, nil
}
