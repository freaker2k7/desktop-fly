package main

import (
	"testing"
	"time"
)

func TestFlyWindowBrainLoopRuns(t *testing.T) {
	fw := &FlyWindow{Brain: NewBrain(), Fly: NewFly(200, 200), Running: true}
	// start brainLoop in goroutine and stop shortly after
	go fw.brainLoop()
	time.Sleep(50 * time.Millisecond)
	fw.Running = false
	// allow loop to exit
	time.Sleep(20 * time.Millisecond)
}
