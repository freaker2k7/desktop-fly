package main

import (
	"log"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	// On macOS GLFW and OpenGL must be used from the main OS thread.
	runtime.LockOSThread()

	// Load Body (brain + eyes + wings)
	body, err := NewBody(DataDir)
	if err != nil {
		log.Fatalln("failed to load body modules:", err)
	}

	// Load skeleton segments for plotting (start with brain root)
	segments, nodePositions, nodeIDs, minX, minY, maxX, maxY, err := LoadSkeletonSegments(RootNeuron, DataDir)
	if err != nil {
		log.Fatalln("failed to load skeletons:", err)
	}

	// Also append nodes for eyes and wings so they appear in the plot
	if eyesSegs, eyesNodes, eyesIDs, ex0, ey0, ex1, ey1, err := LoadSkeletonSegments(EyesNeuron, DataDir); err == nil {
		if len(eyesNodes) > 0 {
			nodePositions = append(nodePositions, eyesNodes...)
			nodeIDs = append(nodeIDs, eyesIDs...)
			if ex0 < minX {
				minX = ex0
			}
			if ey0 < minY {
				minY = ey0
			}
			if ex1 > maxX {
				maxX = ex1
			}
			if ey1 > maxY {
				maxY = ey1
			}
		}
		// also append any segments for eyes so skeletons show too
		if len(eyesSegs) > 0 {
			segments = append(segments, eyesSegs...)
		}
	}
	if wSegs, wNodes, wIDs, wx0, wy0, wx1, wy1, err := LoadSkeletonSegments(WingsNeuron, DataDir); err == nil {
		if len(wNodes) > 0 {
			nodePositions = append(nodePositions, wNodes...)
			nodeIDs = append(nodeIDs, wIDs...)
			if wx0 < minX {
				minX = wx0
			}
			if wy0 < minY {
				minY = wy0
			}
			if wx1 > maxX {
				maxX = wx1
			}
			if wy1 > maxY {
				maxY = wy1
			}
		}
		if len(wSegs) > 0 {
			segments = append(segments, wSegs...)
		}
	}

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to init glfw:", err)
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
		log.Fatalln("failed to create plot window:", err)
	}

	plotWin.Window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		log.Fatalln("failed to init gl:", err)
	}

	if err := plotWin.InitProgram(); err != nil {
		log.Fatalln("failed to init GL program:", err)
	}

	// Upload skeleton segments (if any)
	if len(segments) > 0 {
		if err := plotWin.UploadSegments(segments); err != nil {
			log.Println("warning: failed to upload segments:", err)
		}
		// Ensure node positions are uploaded (no-op change to keep consistent)
		if len(nodePositions) > 0 {
			if err := plotWin.UploadNodePositions(nodePositions, nodeIDs); err != nil {
				log.Println("warning: failed to upload node positions:", err)
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
		log.Fatalln("failed to create fly window:", err)
	}

	// Initialize GL resources for the fly window (make its context current)
	fw.Window.MakeContextCurrent()
	if err := fw.InitGL(); err != nil {
		log.Fatalln("failed to init fly GL:", err)
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
