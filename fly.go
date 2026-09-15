package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Fly struct {
	Width  float64
	Height float64
	X, Y   float64
	Vx, Vy float64
	Angle  float64
	Turn   float64
	Thrust float64
	Wander float64
	// cache last view capture to avoid frequent screencaptures
	lastViewTime time.Time
	lastViewVal  float64
	// visual body radius used for drawing and capture scaling
	BodyRadius float64
}

func NewFly(width, height float64) *Fly {
	f := &Fly{Width: width, Height: height}
	f.X = width * 0.5
	f.Y = height * 0.5
	f.Vx = rand.Float64()*2 - 1
	f.Vy = rand.Float64()*2 - 1
	f.Angle = rand.Float64() * 2 * math.Pi
	f.BodyRadius = 14.0
	return f
}

func (f *Fly) Sensors(mouseX, mouseY float64) map[string]float64 {
	dx := mouseX - f.X
	dy := mouseY - f.Y
	distance := math.Hypot(dx, dy)
	mouseSignal := math.Max(0.0, 1.0-distance/300.0)

	c := math.Cos(-f.Angle)
	s := math.Sin(-f.Angle)
	localX := dx*c - dy*s

	left := math.Max(0.0, -localX/400.0)
	right := math.Max(0.0, localX/400.0)

	brightness := (math.Abs(math.Sin(f.X*0.01)) + math.Abs(math.Cos(f.Y*0.01))) * 0.5

	// Capture a small square region around the fly on screen and
	// compute a vibrancy "view" metric for a 270° FOV centered on
	// the fly's angle. This uses the macOS `screencapture` tool; if
	// it fails we gracefully return view=0.
	cx := int(f.X)
	cy := int(f.Y)
	// make capture radius relative to fly visual size (10x body radius)
	radius := int(f.BodyRadius * 10.0)
	viewVal := 0.0
	now := time.Now()
	if !f.lastViewTime.IsZero() && now.Sub(f.lastViewTime) < time.Millisecond*500 {
		// reuse cached value
		viewVal = f.lastViewVal
	} else {
		if v, err := captureView(cx, cy, radius, f.Angle); err == nil {
			viewVal = v
			f.lastViewVal = v
		}
		// update last attempt time to avoid repeated captures
		f.lastViewTime = now
	}

	return map[string]float64{
		"mouse":      mouseSignal,
		"left":       math.Min(1.0, left),
		"right":      math.Min(1.0, right),
		"brightness": brightness,
		"view":       math.Min(1.0, viewVal),
	}
}

// captureView captures a square region centered at (cx,cy) with the
// given radius and computes a vibrancy metric (0..1) restricted to a
// 270° field of view around `angle`. Uses `screencapture` to save a
// temporary PNG then analyzes pixels. Sampling is sparse for speed.
func captureView(cx, cy, radius int, angle float64) (float64, error) {
	// ensure temp path exists
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("desktopfly_view_%d.png", time.Now().UnixNano()))
	// screencapture expects region as x,y,w,h
	x := cx - radius
	y := cy - radius
	w := radius * 2
	h := radius * 2
	cmd := exec.Command("screencapture", "-x", "-R", fmt.Sprintf("%d,%d,%d,%d", x, y, w, h), tmp)
	if err := cmd.Run(); err != nil {
		return 0.0, err
	}
	defer os.Remove(tmp)

	fh, err := os.Open(tmp)
	if err != nil {
		return 0.0, err
	}
	defer fh.Close()
	img, _, err := image.Decode(fh)
	if err != nil {
		return 0.0, err
	}

	bounds := img.Bounds()
	cxF := float64(bounds.Dx()) / 2.0
	cyF := float64(bounds.Dy()) / 2.0
	// 270° sector -> +/-135°
	halfFOV := 135.0 * math.Pi / 180.0
	var sum float64
	var count int
	step := 4 // sample every 4 pixels
	for py := bounds.Min.Y; py < bounds.Max.Y; py += step {
		for px := bounds.Min.X; px < bounds.Max.X; px += step {
			dx := float64(px) - cxF
			dy := float64(py) - cyF
			theta := math.Atan2(dy, dx)
			// normalize relative angle to [-pi,pi]
			rel := theta - angle
			for rel <= -math.Pi {
				rel += 2 * math.Pi
			}
			for rel > math.Pi {
				rel -= 2 * math.Pi
			}
			if math.Abs(rel) > halfFOV {
				continue
			}
			r, g, b, _ := img.At(px, py).RGBA()
			// convert to 0..1 floats
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0
			mx := math.Max(rf, math.Max(gf, bf))
			mn := math.Min(rf, math.Min(gf, bf))
			// vibrancy: saturation * value approx
			sat := 0.0
			if mx > 0 {
				sat = (mx - mn) / mx
			}
			vibr := sat * mx
			sum += vibr
			count++
		}
	}
	if count == 0 {
		return 0.0, nil
	}
	return sum / float64(count), nil
}

func (f *Fly) Update(dt float64, maxSpeed float64) {
	f.Wander += rand.Float64()*0.3 - 0.15
	f.Angle += f.Turn*0.08 + f.Wander*0.002
	acceleration := f.Thrust*0.35 + 0.04
	f.Vx += math.Cos(f.Angle) * acceleration
	f.Vy += math.Sin(f.Angle) * acceleration
	f.Vx *= 0.985
	f.Vy *= 0.985
	speed := math.Hypot(f.Vx, f.Vy)
	if speed > maxSpeed {
		factor := maxSpeed / speed
		f.Vx *= factor
		f.Vy *= factor
	}
	f.X += f.Vx
	f.Y += f.Vy

	margin := 20.0
	if f.X < margin {
		f.X = margin
		f.Vx = math.Abs(f.Vx)
	} else if f.X > f.Width-margin {
		f.X = f.Width - margin
		f.Vx = -math.Abs(f.Vx)
	}
	if f.Y < margin {
		f.Y = margin
		f.Vy = math.Abs(f.Vy)
	} else if f.Y > f.Height-margin {
		f.Y = f.Height - margin
		f.Vy = -math.Abs(f.Vy)
	}
}
