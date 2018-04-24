package main

import (
	"fmt"
	"strings"

	"container/list"

	"time"

	"github.com/donutmonger/traffic/car"
	"github.com/donutmonger/traffic/color"
	"github.com/donutmonger/traffic/vector"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sirupsen/logrus"
)

var vertexShader = `
#version 330

in vec2 position;
uniform mat3 transformation;

void main() {
    vec3 position_temp = vec3(position, 1.0) * transformation;

    gl_Position = vec4(position_temp.xy, 0.0, 1.0);
}
` + "\x00"

var fragmentShader = `
#version 330

uniform vec4 color;

out vec4 outColor;

void main() {
    outColor = color;
}
` + "\x00"

var planeVertices = []float32{
	-1.0, 1.0,
	1.0, 1.0,
	1.0, -1.0,

	1.0, -1.0,
	-1.0, -1.0,
	-1.0, 1.0,
}

func main() {
	if err := glfw.Init(); err != nil {
		logrus.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(800, 600, "Cube", nil, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		logrus.Fatal(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	logrus.Infof("OpenGL version: %v", version)

	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		logrus.Fatal(err)
	}
	gl.UseProgram(program)

	transform := mgl32.Ident3()
	transformUniformLoc := gl.GetUniformLocation(program, gl.Str("transformation\x00"))
	gl.UniformMatrix3fv(transformUniformLoc, 1, false, &transform[0])

	colorUniformLoc := gl.GetUniformLocation(program, gl.Str("color\x00"))

	gl.BindFragDataLocation(program, 0, gl.Str("outColor\x00"))

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(planeVertices)*4, gl.Ptr(planeVertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.15, 0.15, 0.15, 1.0)

	cars := list.New()
	cars.PushBack(car.New(vector.Vector2{X: -0.9, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.8, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.7, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.6, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.5, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.4, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.3, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.2, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: -0.1, Y: 0.0}, vector.Vector2{X: 0.0, Y: 0.0}))
	cars.PushBack(car.New(vector.Vector2{X: 0.0, Y: 0.0}, vector.Vector2{X: 0.2, Y: 0.0}))

	previousTime := glfw.GetTime()

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		seconds := glfw.GetTime()
		elapsed := seconds - previousTime
		previousTime = seconds

		i := 0
		current := cars.Front()
		for current != nil {
			// TODO
			// Reaction Time (avg human reaction time is 0.25 seconds
			// Feedback loop with distance to next car
			// Don't rely on frame timer
			c := current.Value.(*car.Car)

			next := current.Next()
			if next != nil {
				n := next.Value.(*car.Car)
				c.TargetVelocity = n.Velocity
				if n.Velocity.X != 0.0 {
					c.AddTimeWaited(time.Duration(1000*elapsed) * time.Millisecond)
				}
			} else {
				c.AddTimeWaited(1 * time.Second)
			}

			if c.Velocity.X >= c.TargetVelocity.X {
				c.Acceleration.X = 0.0
			} else if c.HasReacted() {
				c.Acceleration.X = 0.2
			}

			c.Velocity.X += c.Acceleration.X * float32(elapsed)
			c.Velocity.Y += c.Acceleration.Y * float32(elapsed)

			c.Position.X += c.Velocity.X * float32(elapsed)
			c.Position.Y += c.Velocity.Y * float32(elapsed)

			current = next

			i++
		}

		// Render
		gl.UseProgram(program)
		gl.BindVertexArray(vao)

		drawCars(cars,
			func(c color.Color) {
				colorArray := []float32{c.R, c.G, c.B, c.A}
				gl.Uniform4fv(colorUniformLoc, 1, &colorArray[0])
			},
			func(t vector.Vector2) {
				transform := mgl32.Mat3FromCols(
					mgl32.Vec3([3]float32{0.03, 0.00, t.X}),
					mgl32.Vec3([3]float32{0.00, 0.02, t.Y}),
					mgl32.Vec3([3]float32{0.00, 0.00, 1.0}))
				gl.UniformMatrix3fv(transformUniformLoc, 1, false, &transform[0])
			},
			func() {
				gl.DrawArrays(gl.TRIANGLES, 0, 6)
			})

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func drawCars(cars *list.List, setColor func(c color.Color), applyTranslation func(t vector.Vector2), draw func()) {
	current := cars.Front()
	for current != nil {
		c := current.Value.(*car.Car)
		setColor(c.Color)
		applyTranslation(c.Position)
		draw()
		current = current.Next()
	}
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyQ && action == glfw.Press {
		logrus.Info("Exiting...")
		w.SetShouldClose(true)
	}
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
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

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
