package ecs_test

import (
	"reflect"
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
	"github.com/stretchr/testify/assert"
)

func TestView(t *testing.T) {
	storage := ecs.NewStorage()
	entityId := storage.Spawn(&Position{
		X: 1,
		Y: 2,
	}, Temperature(32))

	view := ecs.NewView[struct {
		*Position
		*Temperature
	}]()

	item := view.Get(storage, entityId)
	assert.NotNil(t, item)
	assert.Equal(t, Temperature(32), *item.Temperature)
	assert.Equal(t, float32(1), item.Position.X)
	assert.Equal(t, float32(2), item.Position.Y)
}

func TestViewMultipleComponents(t *testing.T) {
	storage := ecs.NewStorage()
	entityId := storage.Spawn(
		&Position{X: 10, Y: 20},
		&Velocity{DX: 1.5, DY: 2.5},
		&Name{Value: "Test Entity"},
	)

	view := ecs.NewView[struct {
		*Position
		*Velocity
		*Name
	}]()

	item := view.Get(storage, entityId)
	assert.NotNil(t, item)
	assert.Equal(t, float32(10), item.Position.X)
	assert.Equal(t, float32(20), item.Position.Y)
	assert.Equal(t, float32(1.5), item.Velocity.DX)
	assert.Equal(t, float32(2.5), item.Velocity.DY)
	assert.Equal(t, "Test Entity", item.Name.Value)
}

func TestViewMissingComponent(t *testing.T) {
	storage := ecs.NewStorage()
	// Entity only has Position, not Velocity
	entityId := storage.Spawn(&Position{X: 5, Y: 10})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Should return nil because entity is missing Velocity
	item := view.Get(storage, entityId)
	assert.Nil(t, item)
}

func TestViewFill(t *testing.T) {
	storage := ecs.NewStorage()
	entityId := storage.Spawn(&Position{X: 3, Y: 4}, &Health{Current: 50, Max: 100})

	view := ecs.NewView[struct {
		*Position
		*Health
	}]()

	var result struct {
		*Position
		*Health
	}

	// Fill should return true and populate the struct
	ok := view.Fill(storage, entityId, &result)
	assert.True(t, ok)
	assert.NotNil(t, result.Position)
	assert.NotNil(t, result.Health)
	assert.Equal(t, float32(3), result.Position.X)
	assert.Equal(t, 50, result.Health.Current)
}

func TestViewFillMissingComponent(t *testing.T) {
	storage := ecs.NewStorage()
	entityId := storage.Spawn(&Position{X: 1, Y: 2})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	var result struct {
		*Position
		*Velocity
	}

	// Fill should return false because Velocity is missing
	ok := view.Fill(storage, entityId, &result)
	assert.False(t, ok)
}

func TestViewComponentMutation(t *testing.T) {
	storage := ecs.NewStorage()
	entityId := storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0, DY: 0})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	item := view.Get(storage, entityId)
	assert.NotNil(t, item)

	// Mutate the components through the view
	item.Position.X = 100
	item.Position.Y = 200
	item.Velocity.DX = 5
	item.Velocity.DY = 10

	// Verify mutations are persisted in storage
	pos := storage.GetComponent(entityId, reflect.TypeOf(Position{})).(*Position)
	assert.Equal(t, float32(100), pos.X)
	assert.Equal(t, float32(200), pos.Y)

	vel := storage.GetComponent(entityId, reflect.TypeOf(Velocity{})).(*Velocity)
	assert.Equal(t, float32(5), vel.DX)
	assert.Equal(t, float32(10), vel.DY)
}

func TestViewWithPrimitiveComponents(t *testing.T) {
	storage := ecs.NewStorage()
	entityId := storage.Spawn(&Position{X: 7, Y: 8}, Score(1000))

	view := ecs.NewView[struct {
		*Position
		*Score
	}]()

	item := view.Get(storage, entityId)
	assert.NotNil(t, item)
	assert.Equal(t, float32(7), item.Position.X)
	assert.Equal(t, Score(1000), *item.Score)

	// Mutate the primitive
	*item.Score = 2000

	// Verify mutation persisted
	score := storage.GetComponent(entityId, reflect.TypeOf(Score(0))).(*Score)
	assert.Equal(t, Score(2000), *score)
}

func TestViewInvalidEntityId(t *testing.T) {
	storage := ecs.NewStorage()
	fakeId := ecs.NewEntityId(9999, 9999)

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	item := view.Get(storage, fakeId)
	assert.Nil(t, item)
}

func TestViewMultipleEntities(t *testing.T) {
	storage := ecs.NewStorage()

	// Create multiple entities with same components
	id1 := storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1})
	id2 := storage.Spawn(&Position{X: 2, Y: 2}, &Velocity{DX: 0.2, DY: 0.2})
	id3 := storage.Spawn(&Position{X: 3, Y: 3}, &Velocity{DX: 0.3, DY: 0.3})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Verify each entity can be queried correctly
	item1 := view.Get(storage, id1)
	assert.NotNil(t, item1)
	assert.Equal(t, float32(1), item1.Position.X)
	assert.Equal(t, float32(0.1), item1.Velocity.DX)

	item2 := view.Get(storage, id2)
	assert.NotNil(t, item2)
	assert.Equal(t, float32(2), item2.Position.X)
	assert.Equal(t, float32(0.2), item2.Velocity.DX)

	item3 := view.Get(storage, id3)
	assert.NotNil(t, item3)
	assert.Equal(t, float32(3), item3.Position.X)
	assert.Equal(t, float32(0.3), item3.Velocity.DX)
}

func TestViewSubset(t *testing.T) {
	storage := ecs.NewStorage()

	// Entity has more components than the view requires
	entityId := storage.Spawn(
		&Position{X: 5, Y: 5},
		&Velocity{DX: 1, DY: 1},
		&Name{Value: "Extra"},
		&Health{Current: 100, Max: 100},
	)

	// View only asks for a subset
	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	item := view.Get(storage, entityId)
	assert.NotNil(t, item)
	assert.Equal(t, float32(5), item.Position.X)
	assert.Equal(t, float32(1), item.Velocity.DX)
}
