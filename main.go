package main

import (
	"fmt"
	"strings"

	"container/list"

	"math"

	"time"

	"runtime"

	"github.com/donutmonger/traffic/car"
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
    vec3 modelPos = vec3(position, 1.0) * transformation;
	gl_Position = vec4(modelPos.xy, 0.0, 1.0);
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

func init() {
	runtime.LockOSThread()
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
	window, err := glfw.CreateWindow(1280, 720, "Cube", nil, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)
	window.SetScrollCallback(scrollCallback)

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
	colorArray := [4]float32{0, 0, 0, 0}

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
	for i := 120; i >= 0; i-- {
		var x float32 = 15.0*float32(i) + 300
		cars.PushBack(car.New(mgl32.Vec2{-x, 0}, mgl32.Vec2{0, 0}))
	}
	cars.PushBack(car.New(mgl32.Vec2{-285, 0.0}, mgl32.Vec2{30, 0.0}))

	dt := 10 * time.Millisecond
	physicsTicker := time.NewTicker(dt)
	go func() {
		for range physicsTicker.C {
			// Update cars
			i := 0
			current := cars.Front()
			for current != nil {
				// TODO
				// Don't rely on frame timer
				c := current.Value.(*car.Car)

				next := current.Next()
				if next != nil {
					n := next.Value.(*car.Car)
					netDistance := n.Position.X() - c.Position.X() - n.Length
					approachingRate := c.Velocity.X() - n.Velocity.X()

					var desiredVelocity float32 = 30
					var minimumSpacing float32 = 7.5
					var desiredTimeHeadwaySeconds float32 = 1.5
					var maximumAcceleration float32 = 4.0
					var comfortableBrakingDeceleration float32 = 6.0
					var accelerationExponent float32 = 4.0

					firstComponent := float32(math.Pow(float64(c.Velocity.X()/desiredVelocity), float64(accelerationExponent)))
					sStar := minimumSpacing + (c.Velocity.X() * desiredTimeHeadwaySeconds) + ((c.Velocity.X() * approachingRate) / (2 * float32(math.Sqrt(float64(maximumAcceleration*comfortableBrakingDeceleration)))))
					secondComponent := float32(math.Pow(float64(sStar/netDistance), 2.0))
					acceleration := maximumAcceleration * (1 - firstComponent - secondComponent)

					c.Acceleration = mgl32.Vec2{acceleration, 0}

				} else {
					// Lead Car
					if stopLeadCar {
						if c.Velocity.X() > 0 {
							c.Acceleration = mgl32.Vec2{-6, 0}
						} else {
							c.Acceleration = mgl32.Vec2{0, 0}
							c.Velocity = mgl32.Vec2{0, 0}
						}
					} else if startLeadCar && c.Velocity.Len() < c.TargetVelocity.Len() {
						c.Acceleration = mgl32.Vec2{3, 0}
					} else {
						c.Acceleration = mgl32.Vec2{0, 0}
					}
				}

				c.Velocity = c.Velocity.Add(c.Acceleration.Mul(float32(dt.Seconds())))
				c.Position = c.Position.Add(c.Velocity.Mul(float32(dt.Seconds())))

				current = next

				i++
			}

		}
	}()

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render
		gl.UseProgram(program)
		gl.BindVertexArray(vao)

		current := cars.Front()
		for current != nil {
			c := current.Value.(*car.Car)

			colorArray = [4]float32{c.Color.R, c.Color.G, c.Color.B, c.Color.A}
			gl.Uniform4fv(colorUniformLoc, 1, &colorArray[0])

			transform = mgl32.Mat3FromCols(
				mgl32.Vec3{scale * c.Length, 0.00, scale * (c.Position.X() - cameraPosition.X())},
				mgl32.Vec3{0.00, scale * c.Length * 0.667, scale * (c.Position.Y() - cameraPosition.Y())},
				mgl32.Vec3{0.00, 0.00, 1.0})
			gl.UniformMatrix3fv(transformUniformLoc, 1, false, &transform[0])

			gl.DrawArrays(gl.TRIANGLES, 0, 6)

			current = current.Next()
		}

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

var stopLeadCar = false
var startLeadCar = false
var scale float32 = 1 / 400.0
var cameraPosition = mgl32.Vec2{0, 0}

const minScale float32 = 1 / 3000.0

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyQ && action == glfw.Press {
		logrus.Info("Exiting...")
		w.SetShouldClose(true)
	}
	if key == glfw.KeyS && action == glfw.Press {
		logrus.Info("Stopping lead car")
		stopLeadCar = true
		startLeadCar = false
	}
	if key == glfw.KeyG && action == glfw.Press {
		logrus.Info("Starting lead car")
		startLeadCar = true
		stopLeadCar = false
	}
	if key == glfw.KeyRight && (action == glfw.Press || action == glfw.Repeat) {
		logrus.Info(cameraPosition)
		cameraPosition = cameraPosition.Add(mgl32.Vec2{10.0, 0})
	}
	if key == glfw.KeyLeft && (action == glfw.Press || action == glfw.Repeat) {
		logrus.Info(cameraPosition)
		cameraPosition = cameraPosition.Add(mgl32.Vec2{-10.0, 0})
	}
}

func scrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	var scrollSensitivity float32 = 1.0 / 20.0
	scale += scrollSensitivity * float32(yoff) * scale
	if scale < minScale {
		scale = minScale
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
