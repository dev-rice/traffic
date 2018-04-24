package car

import (
	"time"

	"math/rand"

	"github.com/donutmonger/traffic/color"
	"github.com/donutmonger/traffic/vector"
)

const humanReactionTime = 250 * time.Millisecond

type Car struct {
	Position       vector.Vector2
	Velocity       vector.Vector2
	TargetVelocity vector.Vector2
	Acceleration   vector.Vector2

	Color        color.Color
	ReactionTime time.Duration

	TimeSinceAction time.Duration
}

func New(position vector.Vector2, targetVelocity vector.Vector2) *Car {
	return &Car{
		Position:       position,
		Velocity:       vector.Vector2{X: 0, Y: 0},
		TargetVelocity: targetVelocity,
		Acceleration:   vector.Vector2{X: 0, Y: 0},

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
