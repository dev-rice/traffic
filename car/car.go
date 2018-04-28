package car

import (
	"time"

	"math/rand"

	"github.com/donutmonger/traffic/color"
	"github.com/go-gl/mathgl/mgl32"
)

const humanReactionTime = 250 * time.Millisecond

type Car struct {
	Position       mgl32.Vec2
	Velocity       mgl32.Vec2
	TargetVelocity mgl32.Vec2
	Acceleration   mgl32.Vec2
	Length         float32

	Color        color.Color
	ReactionTime time.Duration

	TimeSinceAction time.Duration
}

func New(position mgl32.Vec2, targetVelocity mgl32.Vec2) *Car {
	return &Car{
		Position:       position,
		Velocity:       mgl32.Vec2{0, 0},
		TargetVelocity: targetVelocity,
		Acceleration:   mgl32.Vec2{0, 0},
		Length:         4.8,

		Color:           getRandomColor(),
		ReactionTime:    humanReactionTime,
		TimeSinceAction: 0 * time.Second,
	}
}

func (c *Car) AddTimeWaited(t time.Duration) {
	c.TimeSinceAction += t
}

func (c Car) HasReacted() bool {
	return c.TimeSinceAction >= c.ReactionTime
}

func getRandomColor() color.Color {
	rand.Seed(int64(time.Now().Nanosecond()))
	return color.Color{R: getRandomColorValue(), G: getRandomColorValue(), B: getRandomColorValue(), A: 1.0}
}

func getRandomColorValue() float32 {
	r := rand.Intn(2)
	if r == 0 {
		return 0.541
	} else {
		return 0.803
	}
}
