package ecs_test

import (
	"fmt"
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

func TestViewIter(t *testing.T) {
	storage := ecs.NewStorage()

	// Spawn entities with Position and Velocity
	id1 := storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1})
	id2 := storage.Spawn(&Position{X: 2, Y: 2}, &Velocity{DX: 0.2, DY: 0.2})
	id3 := storage.Spawn(&Position{X: 3, Y: 3}, &Velocity{DX: 0.3, DY: 0.3})

	// Spawn an entity with only Position (should not be included)
	storage.Spawn(&Position{X: 99, Y: 99})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Collect all entities from the iterator
	entities := make(map[ecs.EntityId]struct {
		*Position
		*Velocity
	})

	for id, item := range view.Iter(storage) {
		entities[id] = item
	}

	// Should have exactly 3 entities
	assert.Equal(t, 3, len(entities))

	// Verify each entity is present with correct data
	assert.Contains(t, entities, id1)
	assert.Equal(t, float32(1), entities[id1].Position.X)
	assert.Equal(t, float32(0.1), entities[id1].Velocity.DX)

	assert.Contains(t, entities, id2)
	assert.Equal(t, float32(2), entities[id2].Position.X)
	assert.Equal(t, float32(0.2), entities[id2].Velocity.DX)

	assert.Contains(t, entities, id3)
	assert.Equal(t, float32(3), entities[id3].Position.X)
	assert.Equal(t, float32(0.3), entities[id3].Velocity.DX)
}

func TestViewIterEmpty(t *testing.T) {
	storage := ecs.NewStorage()

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	count := 0
	for range view.Iter(storage) {
		count++
	}

	assert.Equal(t, 0, count)
}

func TestViewIterMultipleArchetypes(t *testing.T) {
	storage := ecs.NewStorage()

	// Create entities with different archetype combinations
	// Archetype 1: Position + Velocity
	id1 := storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1})
	id2 := storage.Spawn(&Position{X: 2, Y: 2}, &Velocity{DX: 0.2, DY: 0.2})

	// Archetype 2: Position + Velocity + Name
	id3 := storage.Spawn(&Position{X: 3, Y: 3}, &Velocity{DX: 0.3, DY: 0.3}, &Name{Value: "Entity3"})
	id4 := storage.Spawn(&Position{X: 4, Y: 4}, &Velocity{DX: 0.4, DY: 0.4}, &Name{Value: "Entity4"})

	// Archetype 3: Position only (should not match)
	storage.Spawn(&Position{X: 99, Y: 99})

	// Archetype 4: Velocity only (should not match)
	storage.Spawn(&Velocity{DX: 99, DY: 99})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Collect all entities
	entities := make(map[ecs.EntityId]bool)
	for id := range view.Iter(storage) {
		entities[id] = true
	}

	// Should match entities from both archetypes 1 and 2
	assert.Equal(t, 4, len(entities))
	assert.True(t, entities[id1])
	assert.True(t, entities[id2])
	assert.True(t, entities[id3])
	assert.True(t, entities[id4])
}

func TestViewIterValues(t *testing.T) {
	storage := ecs.NewStorage()

	storage.Spawn(&Position{X: 1, Y: 10}, &Velocity{DX: 0.1, DY: 1.0})
	storage.Spawn(&Position{X: 2, Y: 20}, &Velocity{DX: 0.2, DY: 2.0})
	storage.Spawn(&Position{X: 3, Y: 30}, &Velocity{DX: 0.3, DY: 3.0})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Collect X values
	xValues := make([]float32, 0)
	for item := range view.IterValues(storage) {
		xValues = append(xValues, item.Position.X)
	}

	assert.Equal(t, 3, len(xValues))
	assert.Contains(t, xValues, float32(1))
	assert.Contains(t, xValues, float32(2))
	assert.Contains(t, xValues, float32(3))
}

func TestViewIterMutation(t *testing.T) {
	storage := ecs.NewStorage()

	id1 := storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0, DY: 0})
	id2 := storage.Spawn(&Position{X: 2, Y: 2}, &Velocity{DX: 0, DY: 0})
	id3 := storage.Spawn(&Position{X: 3, Y: 3}, &Velocity{DX: 0, DY: 0})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Mutate all entities through the iterator
	for _, item := range view.Iter(storage) {
		item.Velocity.DX = item.Position.X * 10
		item.Velocity.DY = item.Position.Y * 10
	}

	// Verify mutations persisted
	vel1 := storage.GetComponent(id1, reflect.TypeOf(Velocity{})).(*Velocity)
	assert.Equal(t, float32(10), vel1.DX)
	assert.Equal(t, float32(10), vel1.DY)

	vel2 := storage.GetComponent(id2, reflect.TypeOf(Velocity{})).(*Velocity)
	assert.Equal(t, float32(20), vel2.DX)
	assert.Equal(t, float32(20), vel2.DY)

	vel3 := storage.GetComponent(id3, reflect.TypeOf(Velocity{})).(*Velocity)
	assert.Equal(t, float32(30), vel3.DX)
	assert.Equal(t, float32(30), vel3.DY)
}

func TestViewIterEarlyBreak(t *testing.T) {
	storage := ecs.NewStorage()

	storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1})
	storage.Spawn(&Position{X: 2, Y: 2}, &Velocity{DX: 0.2, DY: 0.2})
	storage.Spawn(&Position{X: 3, Y: 3}, &Velocity{DX: 0.3, DY: 0.3})
	storage.Spawn(&Position{X: 4, Y: 4}, &Velocity{DX: 0.4, DY: 0.4})
	storage.Spawn(&Position{X: 5, Y: 5}, &Velocity{DX: 0.5, DY: 0.5})

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Break after processing 2 entities
	count := 0
	for range view.Iter(storage) {
		count++
		if count == 2 {
			break
		}
	}

	assert.Equal(t, 2, count)
}

func TestViewIterWithDeletedEntities(t *testing.T) {
	storage := ecs.NewStorage()

	id1 := storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1})
	id2 := storage.Spawn(&Position{X: 2, Y: 2}, &Velocity{DX: 0.2, DY: 0.2})
	id3 := storage.Spawn(&Position{X: 3, Y: 3}, &Velocity{DX: 0.3, DY: 0.3})
	id4 := storage.Spawn(&Position{X: 4, Y: 4}, &Velocity{DX: 0.4, DY: 0.4})

	// Delete the middle entity
	storage.Delete(id2)

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Collect all entities
	entities := make(map[ecs.EntityId]bool)
	for id := range view.Iter(storage) {
		entities[id] = true
	}

	// Should have 3 entities (id2 deleted)
	assert.Equal(t, 3, len(entities))
	assert.True(t, entities[id1])
	assert.False(t, entities[id2]) // Deleted, should not be present
	assert.True(t, entities[id3])
	assert.True(t, entities[id4])
}

func TestViewIterLargeDataset(t *testing.T) {
	storage := ecs.NewStorage()

	const numEntities = 1000

	// Spawn many entities
	for i := 0; i < numEntities; i++ {
		storage.Spawn(
			&Position{X: float32(i), Y: float32(i * 2)},
			&Velocity{DX: float32(i) * 0.1, DY: float32(i) * 0.2},
		)
	}

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Count and verify
	count := 0
	sum := float32(0)
	for _, item := range view.Iter(storage) {
		count++
		sum += item.Position.X
	}

	assert.Equal(t, numEntities, count)
	// Sum of 0 to 999 is 499500
	assert.Equal(t, float32(499500), sum)
}

func TestViewIterWithPrimitives(t *testing.T) {
	storage := ecs.NewStorage()

	storage.Spawn(&Position{X: 1, Y: 1}, Score(100))
	storage.Spawn(&Position{X: 2, Y: 2}, Score(200))
	storage.Spawn(&Position{X: 3, Y: 3}, Score(300))

	view := ecs.NewView[struct {
		*Position
		*Score
	}]()

	totalScore := Score(0)
	for _, item := range view.Iter(storage) {
		totalScore += *item.Score
	}

	assert.Equal(t, Score(600), totalScore)
}

// ExampleView demonstrates basic view usage for querying single entities
func ExampleView() {
	storage := ecs.NewStorage()

	// Spawn entities with different component combinations
	player := storage.Spawn(
		&Position{X: 10, Y: 20},
		&Velocity{DX: 1.5, DY: 0.5},
		&Health{Current: 100, Max: 100},
	)

	enemy := storage.Spawn(
		&Position{X: 50, Y: 30},
		&Velocity{DX: -0.5, DY: 0},
	)

	// Create a view for entities with Position and Velocity
	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Query a specific entity
	if item := view.Get(storage, player); item != nil {
		fmt.Printf("Player at (%.0f, %.0f) moving at (%.1f, %.1f)\n",
			item.Position.X, item.Position.Y,
			item.Velocity.DX, item.Velocity.DY)
	}

	// Query another entity
	if item := view.Get(storage, enemy); item != nil {
		fmt.Printf("Enemy at (%.0f, %.0f) moving at (%.1f, %.1f)\n",
			item.Position.X, item.Position.Y,
			item.Velocity.DX, item.Velocity.DY)
	}

	// Output:
	// Player at (10, 20) moving at (1.5, 0.5)
	// Enemy at (50, 30) moving at (-0.5, 0.0)
}

// ExampleView_Iter demonstrates iterating over all entities matching a view
func ExampleView_Iter() {
	storage := ecs.NewStorage()

	// Spawn multiple entities
	storage.Spawn(&Position{X: 0, Y: 0}, &Velocity{DX: 1, DY: 0})
	storage.Spawn(&Position{X: 10, Y: 5}, &Velocity{DX: 0, DY: 1})
	storage.Spawn(&Position{X: 20, Y: 10}, &Velocity{DX: -1, DY: 0})

	// Entity without velocity - won't be included
	storage.Spawn(&Position{X: 100, Y: 100})

	// Create a view
	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	fmt.Println("Moving entities:")
	// Iterate over all entities with Position and Velocity
	count := 0
	for _, item := range view.Iter(storage) {
		count++
		fmt.Printf("Entity %d: position (%.0f, %.0f), velocity (%.0f, %.0f)\n",
			count, item.Position.X, item.Position.Y,
			item.Velocity.DX, item.Velocity.DY)
	}

	// Output:
	// Moving entities:
	// Entity 1: position (0, 0), velocity (1, 0)
	// Entity 2: position (10, 5), velocity (0, 1)
	// Entity 3: position (20, 10), velocity (-1, 0)
}

// ExampleView_Iter_update demonstrates updating components during iteration
func ExampleView_Iter_update() {
	storage := ecs.NewStorage()

	// Spawn entities
	storage.Spawn(&Position{X: 0, Y: 0}, &Velocity{DX: 1, DY: 2})
	storage.Spawn(&Position{X: 10, Y: 10}, &Velocity{DX: 3, DY: 4})
	storage.Spawn(&Position{X: 20, Y: 20}, &Velocity{DX: 5, DY: 6})

	// Create a view
	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Update all entities: apply velocity to position
	fmt.Println("Applying velocity to position:")
	for item := range view.IterValues(storage) {
		oldX, oldY := item.Position.X, item.Position.Y
		item.Position.X += item.Velocity.DX
		item.Position.Y += item.Velocity.DY
		fmt.Printf("Position moved from (%.0f, %.0f) to (%.0f, %.0f)\n",
			oldX, oldY, item.Position.X, item.Position.Y)
	}

	// Output:
	// Applying velocity to position:
	// Position moved from (0, 0) to (1, 2)
	// Position moved from (10, 10) to (13, 14)
	// Position moved from (20, 20) to (25, 26)
}

// ExampleView_multipleArchetypes demonstrates views matching across different archetypes
func ExampleView_multipleArchetypes() {
	storage := ecs.NewStorage()

	// Archetype 1: Position + Velocity
	storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 1, DY: 0})

	// Archetype 2: Position + Velocity + Health
	storage.Spawn(&Position{X: 2, Y: 2}, &Velocity{DX: 0, DY: 1}, &Health{Current: 100, Max: 100})

	// Archetype 3: Position + Velocity + Name
	storage.Spawn(&Position{X: 3, Y: 3}, &Velocity{DX: 1, DY: 1}, &Name{Value: "Player"})

	// Archetype 4: Position only (won't match)
	storage.Spawn(&Position{X: 99, Y: 99})

	// Create a view that matches Position + Velocity
	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// This will match entities from archetypes 1, 2, and 3 (all have Position + Velocity)
	count := 0
	for item := range view.IterValues(storage) {
		count++
		fmt.Printf("Entity %d: position (%.0f, %.0f)\n", count, item.Position.X, item.Position.Y)
	}
	fmt.Printf("Total entities with Position and Velocity: %d\n", count)

	// Output:
	// Entity 1: position (1, 1)
	// Entity 2: position (2, 2)
	// Entity 3: position (3, 3)
	// Total entities with Position and Velocity: 3
}

// ExampleView_filtering demonstrates filtering entities during iteration
func ExampleView_filtering() {
	storage := ecs.NewStorage()

	// Spawn entities with different health values
	storage.Spawn(&Position{X: 1, Y: 1}, &Health{Current: 25, Max: 100})
	storage.Spawn(&Position{X: 2, Y: 2}, &Health{Current: 75, Max: 100})
	storage.Spawn(&Position{X: 3, Y: 3}, &Health{Current: 30, Max: 100})
	storage.Spawn(&Position{X: 4, Y: 4}, &Health{Current: 90, Max: 100})

	// Create a view
	view := ecs.NewView[struct {
		*Position
		*Health
	}]()

	// Find entities with low health (< 50)
	fmt.Println("Entities with low health:")
	for item := range view.IterValues(storage) {
		if item.Health.Current < 50 {
			fmt.Printf("Position (%.0f, %.0f): health %d/%d\n",
				item.Position.X, item.Position.Y,
				item.Health.Current, item.Health.Max)
		}
	}

	// Output:
	// Entities with low health:
	// Position (1, 1): health 25/100
	// Position (3, 3): health 30/100
}
