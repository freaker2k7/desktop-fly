package main

import (
	"math"
	"math/rand"
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
}

func NewFly(width, height float64) *Fly {
	f := &Fly{Width: width, Height: height}
	f.X = width * 0.5
	f.Y = height * 0.5
	f.Vx = rand.Float64()*2 - 1
	f.Vy = rand.Float64()*2 - 1
	f.Angle = rand.Float64() * 2 * math.Pi
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

	return map[string]float64{
		"mouse":      mouseSignal,
		"left":       math.Min(1.0, left),
		"right":      math.Min(1.0, right),
		"brightness": brightness,
	}
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
