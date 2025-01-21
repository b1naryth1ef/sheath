package ecs_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
	"github.com/stretchr/testify/assert"
)

func TestViewSpawn(t *testing.T) {
	universe := ecs.NewUniverse()
	view := ecs.View[EntityView](universe)
	id := view.Spawn(EntityView{
		Position: &Position{1, 1},
		Name:     &Name{Value: "Test!"},
	})
	obj := view.Get(id)
	assert.Equal(t, obj.Name.Value, "Test!")
	assert.Equal(t, obj.Position.X, float32(1))
	assert.Equal(t, obj.Position.Y, float32(1))
}

func TestViewGet(t *testing.T) {
	universe := ecs.NewUniverse()
	id := universe.Spawn(&Position{1, 1}, &Name{Value: "Test!"})
	id2 := universe.Spawn(&Position{1, 1})
	id3 := universe.Spawn(&Name{Value: "Test!"})
	id4 := universe.Spawn(&PlayerController{})
	id5 := ecs.EntityId(10000)

	view := ecs.View[EntityView](universe)

	fullObj := view.Get(id)
	assert.Equal(t, fullObj.Name.Value, "Test!")
	assert.Equal(t, fullObj.Position.X, float32(1))
	assert.Equal(t, fullObj.Position.Y, float32(1))

	obj, ok := view.MaybeGet(id)
	assert.True(t, ok)
	assert.Equal(t, obj.Name.Value, "Test!")
	assert.Equal(t, obj.Position.X, float32(1))
	assert.Equal(t, obj.Position.Y, float32(1))

	obj, ok = view.MaybeGet(id2)
	assert.False(t, ok)
	assert.Nil(t, obj.Name)
	assert.Equal(t, obj.Position.X, float32(1))
	assert.Equal(t, obj.Position.Y, float32(1))

	obj, ok = view.MaybeGet(id3)
	assert.False(t, ok)
	assert.Nil(t, obj.Position)
	assert.Equal(t, obj.Name.Value, "Test!")

	obj, ok = view.MaybeGet(id4)
	assert.False(t, ok)
	assert.Nil(t, obj.Position)
	assert.Nil(t, obj.Name)

	obj, ok = view.MaybeGet(id5)
	assert.False(t, ok)
	assert.Nil(t, obj.Position)
	assert.Nil(t, obj.Name)
}

func TestViewIter(t *testing.T) {
	universe := ecs.NewUniverse()
	for i := range 10000 {
		universe.Spawn(&Position{float32(i), float32(i)}, &Name{Value: fmt.Sprintf("Entity %d", i)})
	}
	universe.Spawn(&PlayerController{})
	universe.Spawn(&PlayerController{}, &Name{Value: "YOLO"})

	view := ecs.View[EntityView](universe)
	count := 0
	for range view.Iter() {
		count += 1
	}
	assert.Equal(t, 10000, count)

	otherView := ecs.View[struct {
		*PlayerController
		*Position `ecs:"optional"`
		*Name     `ecs:"excluded"`
	}](universe)

	count = 0
	for it := range otherView.Iter() {
		assert.Nil(t, it.Position)
		count += 1
	}
	assert.Equal(t, 1, count)
}

func ExampleView() {
	universe := ecs.NewUniverse()

	// Create a new player view for our universe
	playerView := ecs.View[PlayerView](universe)

	// Spawn via the view, equivalent to universe.Spawn(&Position{1.34, 7.82}, &Name{Value: "Player"}, &PlayerController{})
	playerId := playerView.Spawn(PlayerView{
		Position: &Position{
			1.34,
			7.82,
		},
		Name: &Name{
			Value: "Player",
		},
		PlayerController: &PlayerController{},
	})

	// Fetch a PlayerView{} for the given player id
	player := playerView.Get(playerId)

	// If we're not sure a player exists (or has all the correct components), we
	// can just use MaybeGet(...)
	otherPlayer, ok := playerView.MaybeGet(1337)

	// In this case since our entity id is junk, we will get all nil components
	// and a non-true ok boolean
	if otherPlayer.Position != nil || otherPlayer.Name != nil || ok {
		panic("Uh oh...")
	}

	fmt.Printf("[%s] %0.2f, %0.2f\n", player.Name.Value, player.X, player.Y)
	// Output: [Player] 1.34, 7.82
}

func createTestUniverse(size int) *ecs.Universe {
	universe := ecs.NewUniverse()

	factories := []func() any{
		func() any {
			return &Position{
				rand.Float32(),
				rand.Float32(),
			}
		},
		func() any {
			return &Name{
				Value: fmt.Sprintf("Entity %d", rand.Int32()),
			}
		},
		func() any {
			return &V1{A: uint8(rand.Int32())}
		},
		func() any {
			return &V2{B: uint16(rand.Int32())}
		},
		func() any {
			return &V3{C: uint32(rand.Int32())}
		},
		func() any {
			return &V4{D: uint64(rand.Int32())}
		},
	}

	for range size {
		numComponents := rand.IntN(len(factories)-2) + 1

		components := []any{}
		for _, value := range rand.Perm(numComponents) {
			components = append(components, factories[value]())
		}

		universe.Spawn(components...)
	}

	return universe
}

func BenchmarkViewIter(b *testing.B) {
	universe := createTestUniverse(10000)

	view := ecs.View[EntityView](universe)

	for range b.N {
		count := 0
		for range view.Iter() {
			count += 1
		}
	}
}

func BenchmarkViewGet100(b *testing.B) {
	universe := createTestUniverse(100000)

	view := ecs.View[EntityView](universe)

	for range b.N {
		for range 100 {
			id := ecs.EntityId(rand.IntN(99999) + 1)
			obj := view.Get(id)
			if obj.Position == nil {
				panic("position is nil?")
			}
		}
	}
}

func BenchmarkCreateUniverse(b *testing.B) {
	for range b.N {
		createTestUniverse(100000)
	}
}
