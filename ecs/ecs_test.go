package ecs_test

import (
	"fmt"

	"github.com/b1naryth1ef/sheath/ecs"
)

type Position struct {
	X float32
	Y float32
}

type Name struct {
	Value string
}

type PlayerController struct{}

type EntityView struct {
	*Position
	*Name
}

type PlayerView struct {
	*Position
	*Name
	*PlayerController
}

func ExampleUniverse() {
	universe := ecs.NewUniverse()

	// Spawn some entities into the universe
	for i := range 32 {
		universe.Spawn(&Position{float32(5 * i), float32(10 * i)}, &Name{Value: fmt.Sprintf("Joe #%d", i)})
	}

	// Create a view for all our entities
	view := ecs.View[EntityView](universe)

	// Iterate over entities in our view
	var x, y float32
	for entity := range view.Iter() {
		x += entity.X
		y += entity.Y
	}

	fmt.Printf("%0.0f, %0.0f\n", x, y)
	// Output: 2480, 4960
}
