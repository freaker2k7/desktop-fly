package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	// On macOS GLFW and OpenGL must be used from the main OS thread.
	runtime.LockOSThread()

	if !strings.HasPrefix(DataDir, "/") {
		// Use the running executable location (works when binary is distributed)
		exePath, err := os.Executable()
		if err != nil {
			panic(fmt.Sprintf("Unable to get executable path: %v", err))
		}
		// resolve symlinks if any
		if realExe, err := filepath.EvalSymlinks(exePath); err == nil {
			exePath = realExe
		}
		dir := filepath.Dir(exePath)
		DataDir = filepath.Join(dir, DataDir)

		fmt.Printf("DataDir set to: %s ; %v\n", DataDir, DataDir)
	}

	// Load Body (brain + eyes + wings)
	body, err := NewBody(DataDir)
	if err != nil {
		panic(fmt.Sprintf("failed to load body modules: %v", err))
	}

	// Load skeleton segments for plotting from all configured roots
	roots := []string{RootNeuron, BrainNeuron, EyesNeuron, WingsNeuron}
	segments, nodePositions, nodeIDs, minX, minY, maxX, maxY, err := LoadSkeletons(roots, DataDir)
	if err != nil {
		panic(fmt.Sprintf("failed to load skeletons: %v", err))
	}

	if err := glfw.Init(); err != nil {
		panic(fmt.Sprintf("failed to init glfw: %v", err))
	}
	defer glfw.Terminate()

	// Set macOS-friendly GL window hints before creating windows
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create plot window first and make its context current to initialize GL.
	plotWin, err := NewPlotWindow(1024, 768, "Neuron Plot")
	if err != nil {
		panic(fmt.Sprintf("failed to create plot window: %v", err))
	}

	plotWin.Window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		panic(fmt.Sprintf("failed to init gl: %v", err))
	}

	if err := plotWin.InitProgram(); err != nil {
		panic(fmt.Sprintf("failed to init GL program: %v", err))
	}

	// Upload skeleton segments (if any)
	if len(segments) > 0 {
		if err := plotWin.UploadSegments(segments); err != nil {
			fmt.Printf("warning: failed to upload segments: %v\n", err)
		}
		// Ensure node positions are uploaded (no-op change to keep consistent)
		if len(nodePositions) > 0 {
			if err := plotWin.UploadNodePositions(nodePositions, nodeIDs); err != nil {
				fmt.Printf("warning: failed to upload node positions: %v\n", err)
			}
		}
		// nodes rendering removed; only skeleton segments are uploaded
		// compute a reasonable scale and offset to center content
		ww, wh := plotWin.Window.GetSize()
		// scale such that bbox fits into window
		sx := float32(ww) / (maxX - minX + 1)
		sy := float32(wh) / (maxY - minY + 1)
		plotWin.Scale = 0.9 * float32(mathMin(float64(sx), float64(sy)))
		// offset so center maps to center of window
		cx := (minX + maxX) / 2
		cy := (minY + maxY) / 2
		plotWin.OffsetX = cx - float32(ww)/(float32(2.0)*plotWin.Scale)
		plotWin.OffsetY = cy - float32(wh)/(float32(2.0)*plotWin.Scale)
	}

	// Create the small fly window which will be moved to follow the fly.
	// Make the fly window undecorated and use a transparent framebuffer
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	fw, err := NewFlyWindow(body, plotWin.Window)
	if err != nil {
		panic(fmt.Sprintf("failed to create fly window: %v", err))
	}

	// Initialize GL resources for the fly window (make its context current)
	fw.Window.MakeContextCurrent()
	if err := fw.InitGL(); err != nil {
		panic(fmt.Sprintf("failed to init fly GL: %v", err))
	}

	// Main loop: render both windows and step brain at BRAIN_HZ
	last := time.Now()
	tickDur := time.Second / time.Duration(BrainHz)
	for !plotWin.Window.ShouldClose() && !fw.Window.ShouldClose() {
		now := time.Now()
		dt := now.Sub(last).Seconds()

		// Step brain and update fly at brain tick rate
		if dt >= tickDur.Seconds() {
			// advance the last tick time only when we actually step
			last = now
			sensors := fw.Fly.Sensors(fw.MouseX, fw.MouseY)
			turn, thrust := body.Step(sensors)
			fw.Fly.Turn = turn
			fw.Fly.Thrust = thrust
			fw.Fly.Update(dt, 7.0)

			// move the small fly window so the fly is centered within it
			ww, wh := fw.Window.GetSize()
			// GLFW SetPos expects ints (screen coordinates)
			fx := int(fw.Fly.X) - ww/2
			fy := int(fw.Fly.Y) - wh/2
			fw.Window.SetPos(fx, fy)
		}

		// render plot (center camera on fly position)
		// update active edges from brain state then render
		merged := body.Merged()
		plotWin.UpdateActiveEdges(merged)
		plotWin.Render(float32(fw.Fly.X), float32(fw.Fly.Y))

		// Render fly window (simple blank window; fly visuals handled by OS window content)
		fw.Window.MakeContextCurrent()
		// draw fly
		fw.RenderFly()
		fw.Window.SwapBuffers()

		glfw.PollEvents()
		// small sleep to avoid busy spin
		time.Sleep(10 * time.Millisecond)
	}

	plotWin.Close()
	fw.Window.Destroy()
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
