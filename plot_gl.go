package main

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type PlotWindow struct {
	Window  *glfw.Window
	VBO     uint32
	VAO     uint32
	Count   int32
	Program uint32
	Scale   float32
	OffsetX float32
	OffsetY float32
	// cached attribute/uniform locations
	PosLoc   int32
	ColorLoc int32
	// Node centroids and active-edge overlay
	NodePositions     []float32
	NodeIDs           []int64
	ActiveEdgeVBO     uint32
	InactiveEdgeVBO   uint32
	ActiveEdgeCount   int32
	InactiveEdgeCount int32
	IdToIdx           map[int64]int
}

func NewPlotWindow(width, height int, title string) (*PlotWindow, error) {
	win, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return nil, err
	}
	win.MakeContextCurrent()

	return &PlotWindow{Window: win, Program: 0, Scale: 1.0, OffsetX: 0.0, OffsetY: 0.0}, nil
}

func (pw *PlotWindow) InitProgram() error {
	prog, err := newSimpleProgram()
	if err != nil {
		return err
	}
	pw.Program = prog
	// cache attribute/uniform locations
	pw.PosLoc = gl.GetAttribLocation(pw.Program, gl.Str("pos\x00"))
	pw.ColorLoc = gl.GetUniformLocation(pw.Program, gl.Str("color\x00"))
	// no separate node program; nodes rendered with simple program
	// create and bind a VAO for core-profile compatibility
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	pw.VAO = vao
	// allow shaders to set point size via gl_PointSize
	gl.Enable(gl.PROGRAM_POINT_SIZE)
	return nil
}

func (pw *PlotWindow) UploadSegments(segments []float32) error {
	if len(segments) == 0 {
		return nil
	}
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	size := len(segments) * 4
	gl.BufferData(gl.ARRAY_BUFFER, size, gl.Ptr(&segments[0]), gl.STATIC_DRAW)
	pw.VBO = vbo
	pw.Count = int32(len(segments) / 2) // each vertex has 2 floats
	return nil
}

// UploadNodePositions stores per-node centroids and builds an id->index map.
func (pw *PlotWindow) UploadNodePositions(nodePositions []float32, nodeIDs []int64) error {
	if len(nodePositions) == 0 || len(nodePositions)%2 != 0 {
		return nil
	}
	pw.NodePositions = make([]float32, len(nodePositions))
	copy(pw.NodePositions, nodePositions)
	pw.NodeIDs = make([]int64, len(nodeIDs))
	copy(pw.NodeIDs, nodeIDs)
	// create active/inactive-edge VBOs
	var v uint32
	gl.GenBuffers(1, &v)
	pw.ActiveEdgeVBO = v
	var v2 uint32
	gl.GenBuffers(1, &v2)
	pw.InactiveEdgeVBO = v2
	// build id->index map
	pw.IdToIdx = make(map[int64]int, len(nodeIDs))
	for i, id := range nodeIDs {
		pw.IdToIdx[id] = i
	}
	return nil
}

func (pw *PlotWindow) Render(centerX, centerY float32) {
	pw.Window.MakeContextCurrent()
	gl.ClearColor(0.05, 0.06, 0.08, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// compute ortho projection
	w, h := pw.Window.GetSize()
	proj := mgl32.Ortho2D(0, float32(w), 0, float32(h))

	// camera: center on (centerX, centerY) by translating world coords
	// world coords are in original neuPrint units; we map them to screen
	// using scale and offset.
	model := mgl32.Scale3D(pw.Scale, pw.Scale, 1.0)
	view := mgl32.Translate3D(-pw.OffsetX, -pw.OffsetY, 0)
	mvp := proj.Mul4(model.Mul4(view))

	gl.UseProgram(pw.Program)
	loc := gl.GetUniformLocation(pw.Program, gl.Str("mvp\x00"))
	gl.UniformMatrix4fv(loc, 1, false, &mvp[0])
	// ensure VAO is bound for attribute setups
	gl.BindVertexArray(pw.VAO)

	// draw skeleton if available; otherwise continue to nodes/edges
	if pw.VBO != 0 && pw.Count != 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, pw.VBO)
		if pw.PosLoc >= 0 {
			pos := uint32(pw.PosLoc)
			gl.EnableVertexAttribArray(pos)
			gl.VertexAttribPointer(pos, 2, gl.FLOAT, false, 0, gl.Ptr(nil))
		}
		// set skeleton color (warm yellow)
		if pw.ColorLoc >= 0 {
			gl.Uniform4f(pw.ColorLoc, 1.0, 0.9, 0.6, 1.0)
		} else {
			colLoc := gl.GetUniformLocation(pw.Program, gl.Str("color\x00"))
			gl.Uniform4f(colLoc, 1.0, 0.9, 0.6, 1.0)
		}
		gl.LineWidth(1.0)
		gl.DrawArrays(gl.LINES, 0, pw.Count)
		if pw.PosLoc >= 0 {
			gl.DisableVertexAttribArray(uint32(pw.PosLoc))
		}
	}

	// draw inactive synapse lines (warm yellow)
	if pw.InactiveEdgeVBO != 0 && pw.InactiveEdgeCount > 0 {
		if pw.ColorLoc >= 0 {
			gl.Uniform4f(pw.ColorLoc, 1.0, 0.9, 0.6, 1.0)
		} else {
			colLoc := gl.GetUniformLocation(pw.Program, gl.Str("color\x00"))
			gl.Uniform4f(colLoc, 1.0, 0.9, 0.6, 1.0)
		}
		gl.BindBuffer(gl.ARRAY_BUFFER, pw.InactiveEdgeVBO)
		if pw.PosLoc >= 0 {
			pos := uint32(pw.PosLoc)
			gl.EnableVertexAttribArray(pos)
			gl.VertexAttribPointer(pos, 2, gl.FLOAT, false, 0, gl.Ptr(nil))
		}
		gl.LineWidth(2.0)
		gl.DrawArrays(gl.LINES, 0, pw.InactiveEdgeCount)
		if pw.PosLoc >= 0 {
			gl.DisableVertexAttribArray(uint32(pw.PosLoc))
		}
	}

	// draw active synapse lines (red, much bolder)
	if pw.ActiveEdgeVBO != 0 && pw.ActiveEdgeCount > 0 {
		// fmt.Println("Drawing active edges; count =", pw.ActiveEdgeCount)
		if pw.ColorLoc >= 0 {
			gl.Uniform4f(pw.ColorLoc, 1.0, 0.0, 0.0, 1.0)
		} else {
			colLoc := gl.GetUniformLocation(pw.Program, gl.Str("color\x00"))
			gl.Uniform4f(colLoc, 1.0, 0.0, 0.0, 1.0)
		}
		gl.BindBuffer(gl.ARRAY_BUFFER, pw.ActiveEdgeVBO)
		if pw.PosLoc >= 0 {
			pos := uint32(pw.PosLoc)
			gl.EnableVertexAttribArray(pos)
			gl.VertexAttribPointer(pos, 2, gl.FLOAT, false, 0, gl.Ptr(nil))
		}
		gl.LineWidth(10.0)
		gl.DrawArrays(gl.LINES, 0, pw.ActiveEdgeCount)
		if pw.PosLoc >= 0 {
			gl.DisableVertexAttribArray(uint32(pw.PosLoc))
		}
	}

	pw.Window.SwapBuffers()
}

// UpdateActiveEdges creates line vertices for active connections and uploads them.
func (pw *PlotWindow) UpdateActiveEdges(br *Brain) {
	if pw.ActiveEdgeVBO == 0 || pw.InactiveEdgeVBO == 0 || len(pw.NodePositions) == 0 || pw.IdToIdx == nil {
		return
	}
	activeVerts := make([]float32, 0, 256)
	inactiveVerts := make([]float32, 0, 256)
	const threshold = 0.5
	for _, e := range br.Edges {
		srcN, ok1 := br.Neurons[e.Src]
		dstN, ok2 := br.Neurons[e.Dst]
		if !ok1 || !ok2 {
			continue
		}
		if srcN.Activity >= threshold && dstN.Activity >= threshold {
			si, sok := pw.IdToIdx[e.Src]
			di, dok := pw.IdToIdx[e.Dst]
			if !sok || !dok {
				continue
			}
			sx := pw.NodePositions[si*2]
			sy := pw.NodePositions[si*2+1]
			dx := pw.NodePositions[di*2]
			dy := pw.NodePositions[di*2+1]
			if e.Active {
				activeVerts = append(activeVerts, sx, sy, dx, dy)
			} else {
				inactiveVerts = append(inactiveVerts, sx, sy, dx, dy)
			}
		}
	}

	if len(inactiveVerts) == 0 {
		pw.InactiveEdgeCount = 0
	} else {
		gl.BindBuffer(gl.ARRAY_BUFFER, pw.InactiveEdgeVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(inactiveVerts)*4, gl.Ptr(&inactiveVerts[0]), gl.DYNAMIC_DRAW)
		pw.InactiveEdgeCount = int32(len(inactiveVerts) / 2)
	}

	if len(activeVerts) == 0 {
		pw.ActiveEdgeCount = 0
	} else {
		gl.BindBuffer(gl.ARRAY_BUFFER, pw.ActiveEdgeVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(activeVerts)*4, gl.Ptr(&activeVerts[0]), gl.DYNAMIC_DRAW)
		pw.ActiveEdgeCount = int32(len(activeVerts) / 2)
	}
}

func (pw *PlotWindow) Close() {
	if pw.Window != nil {
		pw.Window.Destroy()
	}
}

func newSimpleProgram() (uint32, error) {
	vertSrc := `#version 410
    in vec2 pos;
    uniform mat4 mvp;
    void main() {
        gl_Position = mvp * vec4(pos, 0.0, 1.0);
    }
    ` + "\x00"

	fragSrc := `#version 410
	out vec4 fragColor;
	uniform vec4 color;
	void main() {
		fragColor = color;
	}
    ` + "\x00"

	vs, err := compileShader(vertSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fs, err := compileShader(fragSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)
	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(prog, logLength, nil, &log[0])
		return 0, fmt.Errorf("program link failed: %s", string(log))
	}
	return prog, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])
		return 0, fmt.Errorf("shader compile failed: %s", string(log))
	}
	return shader, nil
}

func clamp01(v float32) float32 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}
