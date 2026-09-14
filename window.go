package main

import (
	"fmt"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type FlyWindow struct {
	Window     *glfw.Window
	Brain      *Brain
	Fly        *Fly
	MouseX     float64
	MouseY     float64
	Running    bool
	FlyProgram uint32
	FlyVBO     uint32
	VAO        uint32
}

func NewFlyWindow(brain *Brain) (*FlyWindow, error) {
	// assume glfw.Init() and gl.Init() were already called by caller
	// use a small window which will be moved to follow the fly
	// make the small fly window half linear size (1/4 area)
	width := 80
	height := 80
	window, err := glfw.CreateWindow(width, height, "Fly", nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		return nil, err
	}

	// The fly's world size will be set to a large virtual canvas; the small
	// window will follow the fly by moving on-screen.
	virtualW := 2000.0
	virtualH := 1200.0
	fw := &FlyWindow{
		Window:  window,
		Brain:   brain,
		Fly:     NewFly(virtualW, virtualH),
		MouseX:  virtualW / 2,
		MouseY:  virtualH / 2,
		Running: true,
	}

	window.SetCloseCallback(func(w *glfw.Window) {
		fw.Running = false
	})

	return fw, nil
}

func (fw *FlyWindow) Run() {
	go fw.brainLoop()

	for fw.Running && !fw.Window.ShouldClose() {
		fw.RenderFly()
		fw.Window.SwapBuffers()
		glfw.PollEvents()
		time.Sleep(16 * time.Millisecond)
	}

	fw.Window.Destroy()
	glfw.Terminate()
}

func (fw *FlyWindow) InitGL() error {
	// compile simple shader program for fly rendering
	vert := `#version 410
	in vec2 pos;
	uniform mat4 mvp;
	void main() { gl_Position = mvp * vec4(pos, 0.0, 1.0); }
	` + "\x00"
	frag := `#version 410
	uniform vec4 color;
	out vec4 fragColor;
	void main() { fragColor = color; }
	` + "\x00"
	vs, err := compileShader(vert, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fs, err := compileShader(frag, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)
	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		return fmt.Errorf("fly program link failed")
	}
	fw.FlyProgram = prog

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	fw.FlyVBO = vbo
	// create VAO for core profile
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	fw.VAO = vao
	// enable alpha blending for transparent wings/background
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	return nil
}

func (fw *FlyWindow) RenderFly() {
	fw.Window.MakeContextCurrent()
	w, h := fw.Window.GetSize()
	gl.Viewport(0, 0, int32(w), int32(h))
	// transparent background
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// simple ortho projection
	proj := mgl32.Ortho2D(0, float32(w), 0, float32(h))

	// build model to place fly at center of small window
	cx := float32(w) / 2.0
	cy := float32(h) / 2.0

	// draw ellipsoidal body and wings and eyes
	angle := float32(fw.Fly.Angle)
	bodyRx := float32(14)
	bodyRy := float32(8)

	// prepare shader and mvp
	gl.UseProgram(fw.FlyProgram)
	loc := gl.GetUniformLocation(fw.FlyProgram, gl.Str("mvp\x00"))
	gl.UniformMatrix4fv(loc, 1, false, &proj[0])
	col := gl.GetUniformLocation(fw.FlyProgram, gl.Str("color\x00"))

	// bind VAO and VBO
	gl.BindVertexArray(fw.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, fw.FlyVBO)
	pos := uint32(gl.GetAttribLocation(fw.FlyProgram, gl.Str("pos\x00")))
	gl.EnableVertexAttribArray(pos)
	gl.VertexAttribPointer(pos, 2, gl.FLOAT, false, 0, gl.Ptr(nil))

	// body
	drawEllipse := func(cx_, cy_, rx, ry, ang float32, r, g, b, a float32) {
		verts := makeEllipseVerts(cx_, cy_, rx, ry, ang, 32)
		gl.BufferData(gl.ARRAY_BUFFER, len(verts)*4, gl.Ptr(&verts[0]), gl.DYNAMIC_DRAW)
		gl.Uniform4f(col, r, g, b, a)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(verts)/2))
	}

	drawEllipse(cx, cy, bodyRx, bodyRy, angle, 0.08, 0.08, 0.12, 1.0)

	// wings: left and right, slightly offset and flapping
	phase := float32(time.Now().UnixNano()%1000000000) / 1e9 * 2.0 * 3.14159
	flap := float32(math.Sin(float64(phase*6.0))) * 0.6
	wingRx := float32(18)
	wingRy := float32(9)
	// wing centers in local body coordinates
	leftOffsetX := -bodyRx * 0.4
	rightOffsetX := bodyRx * 0.4
	leftCx := cx + (leftOffsetX*float32(math.Cos(float64(angle))) - 0*float32(math.Sin(float64(angle))))
	leftCy := cy + (leftOffsetX*float32(math.Sin(float64(angle))) + 0*float32(math.Cos(float64(angle))))
	rightCx := cx + (rightOffsetX*float32(math.Cos(float64(angle))) - 0*float32(math.Sin(float64(angle))))
	rightCy := cy + (rightOffsetX*float32(math.Sin(float64(angle))) + 0*float32(math.Cos(float64(angle))))
	// draw wings with rotated angle + flap
	drawEllipse(leftCx, leftCy, wingRx, wingRy, angle+flap, 0.9, 0.95, 1.0, 0.35)
	drawEllipse(rightCx, rightCy, wingRx, wingRy, angle-flap, 0.9, 0.95, 1.0, 0.35)

	// eyes: two small red ellipses near front of body
	eyeOffset := bodyRx * 0.6
	eyeRx := float32(3.0)
	eyeRy := float32(2.0)
	frontX := cx + float32(math.Cos(float64(angle)))*eyeOffset
	frontY := cy + float32(math.Sin(float64(angle)))*eyeOffset
	// left/right eye positions perpendicular to angle
	px := float32(-math.Sin(float64(angle)))
	py := float32(math.Cos(float64(angle)))
	leftEyeX := frontX + px*4.0
	leftEyeY := frontY + py*4.0
	rightEyeX := frontX - px*4.0
	rightEyeY := frontY - py*4.0
	drawEllipse(leftEyeX, leftEyeY, eyeRx, eyeRy, 0, 1.0, 0.2, 0.2, 1.0)
	drawEllipse(rightEyeX, rightEyeY, eyeRx, eyeRy, 0, 1.0, 0.2, 0.2, 1.0)

	gl.DisableVertexAttribArray(pos)
}

func (fw *FlyWindow) brainLoop() {
	ticker := time.NewTicker(time.Duration(1000/BrainHz) * time.Millisecond)
	defer ticker.Stop()
	for fw.Running {
		// wait for next tick
		<-ticker.C
		sensors := fw.Fly.Sensors(fw.MouseX, fw.MouseY)
		turn, thrust := fw.Brain.Step(sensors)
		fw.Fly.Turn = turn
		fw.Fly.Thrust = thrust
		fw.Fly.Update(1.0/60.0, 7.0)
	}
}

// makeEllipseVerts returns a triangle fan vertex slice (x,y) centered at cx,cy
// with radii rx,ry and rotated by angle. nSegments controls smoothness.
func makeEllipseVerts(cx, cy, rx, ry, angle float32, nSegments int) []float32 {
	verts := make([]float32, 0, (nSegments+2)*2)
	// center
	verts = append(verts, cx, cy)
	for i := 0; i <= nSegments; i++ {
		theta := float32(i) * 2.0 * 3.14159 / float32(nSegments)
		x := rx * float32(math.Cos(float64(theta)))
		y := ry * float32(math.Sin(float64(theta)))
		// rotate by angle
		rx_ := x*float32(math.Cos(float64(angle))) - y*float32(math.Sin(float64(angle)))
		ry_ := x*float32(math.Sin(float64(angle))) + y*float32(math.Cos(float64(angle)))
		verts = append(verts, cx+rx_, cy+ry_)
	}
	return verts
}
