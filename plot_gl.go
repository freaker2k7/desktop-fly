package main

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type PlotWindow struct {
	Window      *glfw.Window
	VBO         uint32
	VAO         uint32
	Count       int32
	Program     uint32
	NodeVBO     uint32
	ColorVBO    uint32
	SizeVBO     uint32
	NodeCount   int32
	NodeIDs     []int64
	NodeProgram uint32
	Scale       float32
	OffsetX     float32
	OffsetY     float32
	// store node positions locally to build active-edge lines
	NodePositions   []float32
	ActiveEdgeVBO   uint32
	ActiveEdgeCount int32
	FrameCounter    int
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
	// compile node program too
	np, err := newNodeProgram()
	if err != nil {
		return err
	}
	pw.NodeProgram = np
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

func (pw *PlotWindow) UploadNodes(nodePositions []float32, nodeIDs []int64) error {
	if len(nodePositions) == 0 || len(nodePositions)%2 != 0 {
		return nil
	}
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(nodePositions)*4, gl.Ptr(&nodePositions[0]), gl.STATIC_DRAW)
	pw.NodeVBO = vbo
	pw.NodeCount = int32(len(nodePositions) / 2)
	pw.NodeIDs = nodeIDs
	// keep a local copy of node positions for edge drawing
	pw.NodePositions = make([]float32, len(nodePositions))
	copy(pw.NodePositions, nodePositions)

	// create color buffer initialized to zeros
	colors := make([]float32, pw.NodeCount*3)
	var cbo uint32
	gl.GenBuffers(1, &cbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(colors)*4, gl.Ptr(&colors[0]), gl.DYNAMIC_DRAW)
	pw.ColorVBO = cbo
	// create size buffer (point sizes) initialized to default
	sizes := make([]float32, pw.NodeCount)
	for i := range sizes {
		sizes[i] = 6.0
	}
	var sbo uint32
	gl.GenBuffers(1, &sbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, sbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(sizes)*4, gl.Ptr(&sizes[0]), gl.DYNAMIC_DRAW)
	pw.SizeVBO = sbo
	// ensure active edge VBO is initialized (empty)
	var aeb uint32
	gl.GenBuffers(1, &aeb)
	pw.ActiveEdgeVBO = aeb
	return nil
}

// UpdateActiveEdges builds line segments for synapses where both neurons are active
func (pw *PlotWindow) UpdateActiveEdges(br *Brain) {
	if pw.NodeCount == 0 || pw.ActiveEdgeVBO == 0 {
		return
	}
	// map node id to position index
	idToIdx := make(map[int64]int)
	for i, id := range pw.NodeIDs {
		idToIdx[id] = i
	}
	verts := make([]float32, 0)
	// threshold for considering a neuron active
	const actThreshold = 0.5
	for _, e := range br.Edges {
		srcN, ok1 := br.Neurons[e.Src]
		dstN, ok2 := br.Neurons[e.Dst]
		if !ok1 || !ok2 {
			continue
		}
		if srcN.Activity >= actThreshold && dstN.Activity >= actThreshold {
			si, sok := idToIdx[e.Src]
			di, dok := idToIdx[e.Dst]
			if !sok || !dok {
				continue
			}
			sx := pw.NodePositions[si*2]
			sy := pw.NodePositions[si*2+1]
			dx := pw.NodePositions[di*2]
			dy := pw.NodePositions[di*2+1]
			verts = append(verts, sx, sy, dx, dy)
		}
	}
	if len(verts) == 0 {
		pw.ActiveEdgeCount = 0
		return
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, pw.ActiveEdgeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(verts)*4, gl.Ptr(&verts[0]), gl.DYNAMIC_DRAW)
	pw.ActiveEdgeCount = int32(len(verts) / 2)
}

func (pw *PlotWindow) UpdateNodeColors(br *Brain) {
	if pw.NodeVBO == 0 || pw.ColorVBO == 0 || pw.NodeCount == 0 {
		return
	}
	colors := make([]float32, pw.NodeCount*3)
	sizes := make([]float32, pw.NodeCount)
	for i, id := range pw.NodeIDs {
		act := 0.0
		if n, ok := br.Neurons[id]; ok {
			act = n.Activity
		}
		// inactive nodes: white; active nodes (spikes) red
		if act >= 0.5 {
			colors[i*3+0] = 1.0
			colors[i*3+1] = 0.0
			colors[i*3+2] = 0.0
			sizes[i] = 12.0
		} else {
			colors[i*3+0] = 1.0
			colors[i*3+1] = 1.0
			colors[i*3+2] = 1.0
			sizes[i] = 6.0
		}
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, pw.ColorVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(colors)*4, gl.Ptr(&colors[0]))
	if pw.SizeVBO != 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, pw.SizeVBO)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(sizes)*4, gl.Ptr(&sizes[0]))
	}
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

	if pw.VBO == 0 || pw.Count == 0 {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, pw.VBO)
	pos := uint32(gl.GetAttribLocation(pw.Program, gl.Str("pos\x00")))
	gl.EnableVertexAttribArray(pos)
	gl.VertexAttribPointer(pos, 2, gl.FLOAT, false, 0, gl.Ptr(nil))

	// set skeleton color (warm yellow)
	colLoc := gl.GetUniformLocation(pw.Program, gl.Str("color\x00"))
	gl.Uniform4f(colLoc, 1.0, 0.9, 0.6, 1.0)

	gl.LineWidth(1.0)
	gl.DrawArrays(gl.LINES, 0, pw.Count)

	gl.DisableVertexAttribArray(pos)
	// draw nodes
	if pw.NodeVBO != 0 && pw.NodeCount > 0 && pw.NodeProgram != 0 {
		// enable alpha blending for node highlight glow
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

		gl.UseProgram(pw.NodeProgram)
		loc2 := gl.GetUniformLocation(pw.NodeProgram, gl.Str("mvp\x00"))
		gl.UniformMatrix4fv(loc2, 1, false, &mvp[0])

		gl.BindBuffer(gl.ARRAY_BUFFER, pw.NodeVBO)
		posAttr := uint32(gl.GetAttribLocation(pw.NodeProgram, gl.Str("pos\x00")))
		gl.EnableVertexAttribArray(posAttr)
		gl.VertexAttribPointer(posAttr, 2, gl.FLOAT, false, 0, gl.Ptr(nil))

		gl.BindBuffer(gl.ARRAY_BUFFER, pw.ColorVBO)
		colAttr := uint32(gl.GetAttribLocation(pw.NodeProgram, gl.Str("color\x00")))
		gl.EnableVertexAttribArray(colAttr)
		gl.VertexAttribPointer(colAttr, 3, gl.FLOAT, false, 0, gl.Ptr(nil))

		// size attribute
		if pw.SizeVBO != 0 {
			gl.BindBuffer(gl.ARRAY_BUFFER, pw.SizeVBO)
			sizeAttr := uint32(gl.GetAttribLocation(pw.NodeProgram, gl.Str("sizeAttr\x00")))
			if int(sizeAttr) >= 0 {
				gl.EnableVertexAttribArray(sizeAttr)
				gl.VertexAttribPointer(sizeAttr, 1, gl.FLOAT, false, 0, gl.Ptr(nil))
			}
		}

		gl.PointSize(6.0)
		gl.DrawArrays(gl.POINTS, 0, pw.NodeCount)

		// restore blend state (keep it enabled for other draws if needed)
		// caller may change later; leave enabled.

		gl.DisableVertexAttribArray(posAttr)
		gl.DisableVertexAttribArray(colAttr)
		if pw.SizeVBO != 0 {
			sizeAttr := uint32(gl.GetAttribLocation(pw.NodeProgram, gl.Str("sizeAttr\x00")))
			if int(sizeAttr) >= 0 {
				gl.DisableVertexAttribArray(sizeAttr)
			}
		}
	}

	// draw active synapse lines (red)
	if pw.ActiveEdgeVBO != 0 && pw.ActiveEdgeCount > 0 {
		gl.UseProgram(pw.Program)
		// set color uniform to red
		colLoc := gl.GetUniformLocation(pw.Program, gl.Str("color\x00"))
		gl.Uniform4f(colLoc, 1.0, 0.15, 0.15, 1.0)
		gl.BindBuffer(gl.ARRAY_BUFFER, pw.ActiveEdgeVBO)
		pos2 := uint32(gl.GetAttribLocation(pw.Program, gl.Str("pos\x00")))
		gl.EnableVertexAttribArray(pos2)
		gl.VertexAttribPointer(pos2, 2, gl.FLOAT, false, 0, gl.Ptr(nil))
		gl.LineWidth(2.0)
		gl.DrawArrays(gl.LINES, 0, pw.ActiveEdgeCount)
		gl.DisableVertexAttribArray(pos2)
	}

	pw.Window.SwapBuffers()
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

func newNodeProgram() (uint32, error) {
	vertSrc := `#version 410
	in vec2 pos;
	in vec3 color;
	in float sizeAttr;
	out vec3 vcolor;
	uniform mat4 mvp;
	void main() {
		vcolor = color;
		gl_Position = mvp * vec4(pos, 0.0, 1.0);
		gl_PointSize = sizeAttr;
	}
	` + "\x00"

	fragSrc := `#version 410
	in vec3 vcolor;
	out vec4 fragColor;
	void main() {
		// make circular point sprites by discarding fragments outside radius
		vec2 coord = gl_PointCoord - vec2(0.5);
		float dist = length(coord);
		if (dist > 0.5) {
			discard;
		}
		fragColor = vec4(vcolor, 1.0);
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
		return 0, fmt.Errorf("node program link failed: %s", string(log))
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
