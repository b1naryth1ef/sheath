package ecs_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
	"github.com/stretchr/testify/assert"
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

func TestRemoveComponent(t *testing.T) {
	universe := createTestUniverse(1000)
	id := universe.Spawn(&Position{X: 1, Y: 1}, &PlayerController{})
	ref := universe.Get(id)

	playerControllerType := reflect.TypeOf(&PlayerController{})
	assert.True(t, ref.HasComponent(playerControllerType))
	assert.True(t, ref.RemoveComponent(playerControllerType))
	assert.False(t, ref.HasComponent(playerControllerType))
}

func TestAddComponent(t *testing.T) {
	universe := createTestUniverse(1000)
	id := universe.Spawn(&Position{X: 1, Y: 1})
	ref := universe.Get(id)

	v1Type := reflect.TypeOf(&V1{})
	assert.False(t, ref.HasComponent(v1Type))
	ref.AddComponent(&V1{A: uint8(42)})
	assert.True(t, ref.HasComponent(v1Type))

	var v1 *V1
	assert.True(t, ref.Read(&v1))
	assert.Equal(t, uint8(42), v1.A)
}

func TestGetComponent(t *testing.T) {
	universe := createTestUniverse(1000)
	id := universe.Spawn(&Position{X: 3, Y: 4})
	ref := universe.Get(id)

	pos := ref.GetComponent(reflect.TypeOf(&Position{})).(*Position)
	assert.Equal(t, float32(3), pos.X)
	assert.Equal(t, float32(4), pos.Y)
}

func TestHasComponent(t *testing.T) {
	universe := createTestUniverse(1000)
	id := universe.Spawn(&Position{X: 3, Y: 4})
	ref := universe.Get(id)

	assert.True(t, ref.HasComponent(reflect.TypeOf(&Position{})))
	assert.False(t, ref.HasComponent(reflect.TypeOf(&PlayerController{})))
}
