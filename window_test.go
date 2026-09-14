package main

import (
	"testing"
	"time"
)

func TestFlyWindowBrainLoopRuns(t *testing.T) {
	body := &Body{Brain: NewBrain(), Eyes: NewBrain(), Wings: NewBrain()}
	fw := &FlyWindow{Body: body, Fly: NewFly(200, 200), Running: true}
	// start brainLoop in goroutine and stop shortly after
	go fw.brainLoop()
	time.Sleep(50 * time.Millisecond)
	fw.Running = false
	// allow loop to exit
	time.Sleep(20 * time.Millisecond)
}
