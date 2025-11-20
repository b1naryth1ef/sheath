package ecs_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
	"github.com/stretchr/testify/assert"
)

// Test component types
type Position struct {
	X, Y float32
}

type Velocity struct {
	DX, DY float32
}

type Name struct {
	Value string
}

type Health struct {
	Current int
	Max     int
}

type PlayerController struct{}

type AI struct {
	State int
}

// Custom primitive types for testing non-pointer components
type Score int32
type Tag string
type Temperature float64

// Test EntityId encoding/decoding
func TestEntityIdEncoding(t *testing.T) {
	archetypeId := uint32(12345)
	index := uint32(67890)

	entityId := ecs.NewEntityId(archetypeId, index)

	assert.Equal(t, archetypeId, entityId.ArchetypeId())
	assert.Equal(t, index, entityId.Index())
}

func TestEntityIdEdgeCases(t *testing.T) {
	tests := []struct {
		archetypeId uint32
		index       uint32
	}{
		{0, 0},
		{0xFFFFFFFF, 0xFFFFFFFF},
		{1, 0},
		{0, 1},
		{0x12345678, 0x9ABCDEF0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("archetype=%d,index=%d", tt.archetypeId, tt.index), func(t *testing.T) {
			entityId := ecs.NewEntityId(tt.archetypeId, tt.index)
			assert.Equal(t, tt.archetypeId, entityId.ArchetypeId())
			assert.Equal(t, tt.index, entityId.Index())
		})
	}
}

// Test basic storage operations
func TestSpawnEntity(t *testing.T) {
	storage := ecs.NewStorage()

	id := storage.Spawn(&Position{X: 1.0, Y: 2.0}, &Velocity{DX: 0.5, DY: 0.5}, Score(32))
	assert.NotEqual(t, ecs.EntityId(0), id)

	// Verify archetype ID is encoded
	assert.Greater(t, id.ArchetypeId(), uint32(0))
}

func TestGetComponent(t *testing.T) {
	storage := ecs.NewStorage()

	id := storage.Spawn(&Position{X: 3.0, Y: 4.0}, &Name{Value: "Test Entity"})

	// Get Position component
	posComp := storage.GetComponent(id, reflect.TypeOf(Position{}))
	assert.NotNil(t, posComp)
	pos := posComp.(*Position)
	assert.Equal(t, float32(3.0), pos.X)
	assert.Equal(t, float32(4.0), pos.Y)

	// Get Name component
	nameComp := storage.GetComponent(id, reflect.TypeOf(Name{}))
	assert.NotNil(t, nameComp)
	name := nameComp.(*Name)
	assert.Equal(t, "Test Entity", name.Value)

	// Try to get non-existent component
	velocityComp := storage.GetComponent(id, reflect.TypeOf(Velocity{}))
	assert.Nil(t, velocityComp)
}

func TestDeleteEntity(t *testing.T) {
	storage := ecs.NewStorage()

	id := storage.Spawn(&Position{X: 1.0, Y: 1.0}, &Health{Current: 100, Max: 100})

	// Verify entity exists
	comp := storage.GetComponent(id, reflect.TypeOf(Position{}))
	assert.NotNil(t, comp)

	// Delete entity
	storage.Delete(id)

	// Verify entity is gone
	comp = storage.GetComponent(id, reflect.TypeOf(Position{}))
	assert.Nil(t, comp)
}

func TestMultipleEntitiesSameArchetype(t *testing.T) {
	storage := ecs.NewStorage()

	// Spawn multiple entities with same component types
	id1 := storage.Spawn(&Position{X: 1.0, Y: 1.0}, &Velocity{DX: 0.1, DY: 0.1})
	id2 := storage.Spawn(&Position{X: 2.0, Y: 2.0}, &Velocity{DX: 0.2, DY: 0.2})
	id3 := storage.Spawn(&Position{X: 3.0, Y: 3.0}, &Velocity{DX: 0.3, DY: 0.3})

	// They should all have the same archetype ID
	assert.Equal(t, id1.ArchetypeId(), id2.ArchetypeId())
	assert.Equal(t, id1.ArchetypeId(), id3.ArchetypeId())

	// But different entity indices
	assert.NotEqual(t, id1.Index(), id2.Index())
	assert.NotEqual(t, id1.Index(), id3.Index())
	assert.NotEqual(t, id2.Index(), id3.Index())

	// Verify components are correct
	pos1 := storage.GetComponent(id1, reflect.TypeOf(Position{})).(*Position)
	pos2 := storage.GetComponent(id2, reflect.TypeOf(Position{})).(*Position)
	pos3 := storage.GetComponent(id3, reflect.TypeOf(Position{})).(*Position)

	assert.Equal(t, float32(1.0), pos1.X)
	assert.Equal(t, float32(2.0), pos2.X)
	assert.Equal(t, float32(3.0), pos3.X)
}

func TestMultipleDifferentArchetypes(t *testing.T) {
	storage := ecs.NewStorage()

	id1 := storage.Spawn(&Position{X: 1.0, Y: 1.0})
	id2 := storage.Spawn(&Position{X: 2.0, Y: 2.0}, &Velocity{DX: 0.1, DY: 0.1})
	id3 := storage.Spawn(&Position{X: 3.0, Y: 3.0}, &Name{Value: "Entity 3"})
	id4 := storage.Spawn(&Health{Current: 50, Max: 100})

	// All should have different archetype IDs
	assert.NotEqual(t, id1.ArchetypeId(), id2.ArchetypeId())
	assert.NotEqual(t, id1.ArchetypeId(), id3.ArchetypeId())
	assert.NotEqual(t, id1.ArchetypeId(), id4.ArchetypeId())
	assert.NotEqual(t, id2.ArchetypeId(), id3.ArchetypeId())
	assert.NotEqual(t, id2.ArchetypeId(), id4.ArchetypeId())
	assert.NotEqual(t, id3.ArchetypeId(), id4.ArchetypeId())

	// Verify components
	assert.NotNil(t, storage.GetComponent(id1, reflect.TypeOf(Position{})))
	assert.Nil(t, storage.GetComponent(id1, reflect.TypeOf(Velocity{})))

	assert.NotNil(t, storage.GetComponent(id2, reflect.TypeOf(Position{})))
	assert.NotNil(t, storage.GetComponent(id2, reflect.TypeOf(Velocity{})))
	assert.Nil(t, storage.GetComponent(id2, reflect.TypeOf(Name{})))

	assert.NotNil(t, storage.GetComponent(id4, reflect.TypeOf(Health{})))
	assert.Nil(t, storage.GetComponent(id4, reflect.TypeOf(Position{})))
}

func TestHasComponent(t *testing.T) {
	storage := ecs.NewStorage()

	id := storage.Spawn(&Position{X: 1.0, Y: 1.0}, &Velocity{DX: 0.5, DY: 0.5})

	assert.True(t, storage.HasComponent(id, reflect.TypeOf(Position{})))
	assert.True(t, storage.HasComponent(id, reflect.TypeOf(Velocity{})))
	assert.False(t, storage.HasComponent(id, reflect.TypeOf(Name{})))
	assert.False(t, storage.HasComponent(id, reflect.TypeOf(Health{})))
}

func TestComponentMutation(t *testing.T) {
	storage := ecs.NewStorage()

	id := storage.Spawn(&Position{X: 1.0, Y: 1.0})

	// Get and mutate component
	pos := storage.GetComponent(id, reflect.TypeOf(Position{})).(*Position)
	pos.X = 10.0
	pos.Y = 20.0

	// Verify mutation persisted
	pos2 := storage.GetComponent(id, reflect.TypeOf(Position{})).(*Position)
	assert.Equal(t, float32(10.0), pos2.X)
	assert.Equal(t, float32(20.0), pos2.Y)
}

func TestDeleteWithStableIndices(t *testing.T) {
	storage := ecs.NewStorage()

	// Spawn several entities with same archetype
	id1 := storage.Spawn(&Position{X: 1.0, Y: 1.0}, &Velocity{DX: 0.1, DY: 0.1})
	id2 := storage.Spawn(&Position{X: 2.0, Y: 2.0}, &Velocity{DX: 0.2, DY: 0.2})
	id3 := storage.Spawn(&Position{X: 3.0, Y: 3.0}, &Velocity{DX: 0.3, DY: 0.3})
	id4 := storage.Spawn(&Position{X: 4.0, Y: 4.0}, &Velocity{DX: 0.4, DY: 0.4})

	// Delete middle entity
	storage.Delete(id2)

	// Verify id2 is gone
	assert.Nil(t, storage.GetComponent(id2, reflect.TypeOf(Position{})))

	// Verify others still exist with correct data (indices remain stable)
	pos1 := storage.GetComponent(id1, reflect.TypeOf(Position{})).(*Position)
	assert.Equal(t, float32(1.0), pos1.X)

	pos3 := storage.GetComponent(id3, reflect.TypeOf(Position{})).(*Position)
	assert.Equal(t, float32(3.0), pos3.X)

	pos4 := storage.GetComponent(id4, reflect.TypeOf(Position{})).(*Position)
	assert.Equal(t, float32(4.0), pos4.X)

	// Spawn a new entity - it should reuse the deleted slot
	id5 := storage.Spawn(&Position{X: 5.0, Y: 5.0}, &Velocity{DX: 0.5, DY: 0.5})

	// Verify new entity uses same archetype
	assert.Equal(t, id1.ArchetypeId(), id5.ArchetypeId())

	// Verify new entity data
	pos5 := storage.GetComponent(id5, reflect.TypeOf(Position{})).(*Position)
	assert.Equal(t, float32(5.0), pos5.X)
}

func TestLargeNumberOfEntities(t *testing.T) {
	storage := ecs.NewStorage()

	const numEntities = 10000

	ids := make([]ecs.EntityId, numEntities)
	for i := range numEntities {
		ids[i] = storage.Spawn(
			&Position{X: float32(i), Y: float32(i * 2)},
			&Health{Current: i, Max: i * 10},
		)
	}

	// Verify all entities
	for i, id := range ids {
		pos := storage.GetComponent(id, reflect.TypeOf(Position{})).(*Position)
		assert.Equal(t, float32(i), pos.X)
		assert.Equal(t, float32(i*2), pos.Y)

		health := storage.GetComponent(id, reflect.TypeOf(Health{})).(*Health)
		assert.Equal(t, i, health.Current)
		assert.Equal(t, i*10, health.Max)
	}
}

func TestComponentTypeOrderIndependence(t *testing.T) {
	storage := ecs.NewStorage()

	// Spawn entities with same components but in different order
	id1 := storage.Spawn(&Position{X: 1.0, Y: 1.0}, &Velocity{DX: 0.1, DY: 0.1}, &Name{Value: "A"})
	id2 := storage.Spawn(&Velocity{DX: 0.2, DY: 0.2}, &Name{Value: "B"}, &Position{X: 2.0, Y: 2.0})
	id3 := storage.Spawn(&Name{Value: "C"}, &Position{X: 3.0, Y: 3.0}, &Velocity{DX: 0.3, DY: 0.3})

	// All should have the same archetype ID (components are sorted internally)
	assert.Equal(t, id1.ArchetypeId(), id2.ArchetypeId())
	assert.Equal(t, id1.ArchetypeId(), id3.ArchetypeId())

	// Verify components are stored correctly
	pos1 := storage.GetComponent(id1, reflect.TypeOf(Position{})).(*Position)
	pos2 := storage.GetComponent(id2, reflect.TypeOf(Position{})).(*Position)
	pos3 := storage.GetComponent(id3, reflect.TypeOf(Position{})).(*Position)

	assert.Equal(t, float32(1.0), pos1.X)
	assert.Equal(t, float32(2.0), pos2.X)
	assert.Equal(t, float32(3.0), pos3.X)
}

func TestInvalidEntityId(t *testing.T) {
	storage := ecs.NewStorage()

	// Try to get component for non-existent entity
	fakeId := ecs.NewEntityId(9999, 9999)
	comp := storage.GetComponent(fakeId, reflect.TypeOf(Position{}))
	assert.Nil(t, comp)

	// Delete non-existent entity (should not panic)
	storage.Delete(fakeId)
}

func TestPrimitiveComponents(t *testing.T) {
	storage := ecs.NewStorage()

	// Test with custom primitive types (non-pointer)
	id := storage.Spawn(Score(1337), Tag("player"), Temperature(98.6))

	// Verify we can get the components back
	scoreComp := storage.GetComponent(id, reflect.TypeOf(Score(0)))
	assert.NotNil(t, scoreComp)
	score := scoreComp.(*Score)
	assert.Equal(t, Score(1337), *score)

	tagComp := storage.GetComponent(id, reflect.TypeOf(Tag("")))
	assert.NotNil(t, tagComp)
	tag := tagComp.(*Tag)
	assert.Equal(t, Tag("player"), *tag)

	tempComp := storage.GetComponent(id, reflect.TypeOf(Temperature(0)))
	assert.NotNil(t, tempComp)
	temp := tempComp.(*Temperature)
	assert.Equal(t, Temperature(98.6), *temp)
}

func TestMixedStructAndPrimitiveComponents(t *testing.T) {
	storage := ecs.NewStorage()

	// Mix struct pointers and primitive values
	id := storage.Spawn(&Position{X: 10, Y: 20}, Score(100), &Name{Value: "test"})

	// Verify all components
	pos := storage.GetComponent(id, reflect.TypeOf(Position{})).(*Position)
	assert.Equal(t, float32(10), pos.X)

	score := storage.GetComponent(id, reflect.TypeOf(Score(0))).(*Score)
	assert.Equal(t, Score(100), *score)

	name := storage.GetComponent(id, reflect.TypeOf(Name{})).(*Name)
	assert.Equal(t, "test", name.Value)
}

func TestPrimitiveMutation(t *testing.T) {
	storage := ecs.NewStorage()

	id := storage.Spawn(Score(100))

	// Get and mutate the component
	score := storage.GetComponent(id, reflect.TypeOf(Score(0))).(*Score)
	*score = 500

	// Verify mutation persisted
	score2 := storage.GetComponent(id, reflect.TypeOf(Score(0))).(*Score)
	assert.Equal(t, Score(500), *score2)
}

func TestBuiltinPrimitives(t *testing.T) {
	storage := ecs.NewStorage()

	// Test with built-in types (not custom types)
	id := storage.Spawn(int32(42), float64(3.14), string("hello"))

	// Verify we can get them back
	intComp := storage.GetComponent(id, reflect.TypeOf(int32(0))).(*int32)
	assert.Equal(t, int32(42), *intComp)

	floatComp := storage.GetComponent(id, reflect.TypeOf(float64(0))).(*float64)
	assert.Equal(t, 3.14, *floatComp)

	strComp := storage.GetComponent(id, reflect.TypeOf(string(""))).(*string)
	assert.Equal(t, "hello", *strComp)
}

type TestA string
type TestB string

func TestComponentReader(t *testing.T) {
	storage := ecs.NewStorage()
	id := storage.Spawn(TestA("A"), TestB("B"))

	testA := ecs.ReadComponent[TestA](storage, id)
	assert.Equal(t, *testA, TestA("A"))

	testB := ecs.ReadComponent[TestB](storage, id)
	assert.Equal(t, *testB, TestB("B"))
}

func TestGetArchetype(t *testing.T) {
	storage := ecs.NewStorage()

	id := storage.Spawn(TestA("A"), TestB("B"))

	arch1 := storage.GetArchetype(TestA("A"), TestB("B"))
	arch2 := storage.GetArchetypeByTypes([]reflect.Type{reflect.TypeFor[TestA](), reflect.TypeFor[TestB]()})
	assert.Equal(t, arch1, arch2)

	assert.Equal(t, *arch1.GetComponent(id.Index(), reflect.TypeFor[TestA]()).(*TestA), TestA("A"))
}
