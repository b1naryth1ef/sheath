package ecs_test

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
)

// Additional component types for benchmarking
type V1 struct{ A uint8 }
type V2 struct{ B uint16 }
type V3 struct{ C uint32 }
type V4 struct{ D uint64 }

// BenchmarkSpawnEntity tests entity spawning performance
func BenchmarkSpawnEntity(b *testing.B) {
	storage := ecs.NewStorage()

	b.ResetTimer()
	for range b.N {
		storage.Spawn(&Position{X: 1.0, Y: 1.0})
	}
}

// BenchmarkSpawnEntityMultiComponent tests spawning with multiple components
func BenchmarkSpawnEntityMultiComponent(b *testing.B) {
	storage := ecs.NewStorage()

	b.ResetTimer()
	for range b.N {
		storage.Spawn(
			&Position{X: 1.0, Y: 1.0},
			&Velocity{DX: 0.5, DY: 0.5},
			&Health{Current: 100, Max: 100},
			&Name{Value: "Entity"},
		)
	}
}

// BenchmarkSpawn100k spawns 100k entities
func BenchmarkSpawn100k(b *testing.B) {
	for range b.N {
		storage := ecs.NewStorage()
		for range 100000 {
			storage.Spawn(&Position{X: 1.0, Y: 1.0})
		}
	}
}

// BenchmarkSpawn100kMultiComponent spawns 100k entities with multiple components
func BenchmarkSpawn100kMultiComponent(b *testing.B) {
	for range b.N {
		storage := ecs.NewStorage()
		for range 100000 {
			storage.Spawn(
				&Position{X: 1.0, Y: 1.0},
				&Velocity{DX: 0.5, DY: 0.5},
				&Health{Current: 100, Max: 100},
			)
		}
	}
}

// BenchmarkGetComponent tests component retrieval performance
func BenchmarkGetComponent(b *testing.B) {
	storage := ecs.NewStorage()

	// Pre-populate with entities
	ids := make([]ecs.EntityId, 100000)
	for i := range 100000 {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)}, &Velocity{DX: 0.1, DY: 0.1})
	}

	posType := reflect.TypeOf(Position{})

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(100000)
		comp := storage.GetComponent(ids[idx], posType)
		if comp == nil {
			b.Fatal("component is nil")
		}
	}
}

// BenchmarkGetComponent100 gets 100 components in a batch
func BenchmarkGetComponent100(b *testing.B) {
	storage := ecs.NewStorage()

	// Pre-populate with entities
	ids := make([]ecs.EntityId, 100000)
	for i := range 100000 {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)}, &Velocity{DX: 0.1, DY: 0.1})
	}

	posType := reflect.TypeOf(Position{})

	b.ResetTimer()
	for range b.N {
		for range 100 {
			idx := rand.IntN(100000)
			comp := storage.GetComponent(ids[idx], posType)
			if comp == nil {
				b.Fatal("component is nil")
			}
		}
	}
}

// BenchmarkDeleteEntity tests entity deletion performance
func BenchmarkDeleteEntity(b *testing.B) {
	storage := ecs.NewStorage()

	// Pre-populate and track IDs for deletion
	ids := make([]ecs.EntityId, 100000)
	for i := range 100000 {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)})
	}

	b.ResetTimer()
	for i := range b.N {
		if i >= len(ids) {
			break
		}
		storage.Delete(ids[i])
	}
}

// BenchmarkMixedOperations simulates realistic mixed workload (reads, spawns, deletes)
func BenchmarkMixedOperations(b *testing.B) {
	storage := ecs.NewStorage()

	// Pre-populate with some entities
	ids := make([]ecs.EntityId, 0, 10000)
	for range 1000 {
		id := storage.Spawn(
			&Position{X: rand.Float32(), Y: rand.Float32()},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
		)
		ids = append(ids, id)
	}

	posType := reflect.TypeOf(Position{})

	b.ResetTimer()
	for range b.N {
		// 70% reads, 20% spawns, 10% deletes
		op := rand.IntN(100)

		switch {
		case op < 70: // Read
			if len(ids) > 0 {
				idx := rand.IntN(len(ids))
				storage.GetComponent(ids[idx], posType)
			}

		case op < 90: // Spawn
			id := storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
			ids = append(ids, id)

		default: // Delete
			if len(ids) > 0 {
				idx := rand.IntN(len(ids))
				storage.Delete(ids[idx])
				ids = append(ids[:idx], ids[idx+1:]...)
			}
		}
	}
}

// BenchmarkMultipleArchetypes tests performance with many different archetypes
func BenchmarkMultipleArchetypes(b *testing.B) {
	storage := ecs.NewStorage()

	componentSets := [][]any{
		{&Position{X: 1, Y: 1}},
		{&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1}},
		{&Position{X: 1, Y: 1}, &Health{Current: 100, Max: 100}},
		{&Position{X: 1, Y: 1}, &Name{Value: "Test"}},
		{&Velocity{DX: 0.1, DY: 0.1}, &Health{Current: 100, Max: 100}},
		{&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1}, &Health{Current: 100, Max: 100}},
		{&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1}, &Name{Value: "Test"}},
		{&V1{A: 1}, &V2{B: 2}, &V3{C: 3}, &V4{D: 4}},
	}

	b.ResetTimer()
	for range b.N {
		for range 100 {
			compSet := componentSets[rand.IntN(len(componentSets))]
			storage.Spawn(compSet...)
		}
	}
}

// BenchmarkRandomAccess tests random access patterns
func BenchmarkRandomAccess(b *testing.B) {
	storage := ecs.NewStorage()

	// Create entities with random component combinations
	ids := make([]ecs.EntityId, 10000)
	factories := []func() []any{
		func() []any {
			return []any{&Position{X: rand.Float32(), Y: rand.Float32()}}
		},
		func() []any {
			return []any{
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			}
		},
		func() []any {
			return []any{
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Health{Current: rand.IntN(100), Max: 100},
			}
		},
		func() []any {
			return []any{
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
				&Health{Current: rand.IntN(100), Max: 100},
			}
		},
	}

	for i := range ids {
		factory := factories[rand.IntN(len(factories))]
		ids[i] = storage.Spawn(factory()...)
	}

	posType := reflect.TypeOf(Position{})
	velType := reflect.TypeOf(Velocity{})

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(ids))
		storage.GetComponent(ids[idx], posType)
		storage.GetComponent(ids[idx], velType)
	}
}

// BenchmarkEntityIdOperations tests EntityId encoding/decoding performance
func BenchmarkEntityIdOperations(b *testing.B) {
	b.Run("Encode", func(b *testing.B) {
		for range b.N {
			ecs.NewEntityId(12345, 67890)
		}
	})

	b.Run("DecodeArchetype", func(b *testing.B) {
		id := ecs.NewEntityId(12345, 67890)
		b.ResetTimer()
		for range b.N {
			_ = id.ArchetypeId()
		}
	})

	b.Run("DecodeIndex", func(b *testing.B) {
		id := ecs.NewEntityId(12345, 67890)
		b.ResetTimer()
		for range b.N {
			_ = id.Index()
		}
	})
}

// BenchmarkArchetypeLookup tests archetype lookup performance
func BenchmarkArchetypeLookup(b *testing.B) {
	storage := ecs.NewStorage()

	// Create entities in multiple archetypes
	for range 1000 {
		storage.Spawn(&Position{X: 1, Y: 1})
		storage.Spawn(&Position{X: 1, Y: 1}, &Velocity{DX: 0.1, DY: 0.1})
		storage.Spawn(&Position{X: 1, Y: 1}, &Health{Current: 100, Max: 100})
	}

	posType := reflect.TypeOf(Position{})

	b.ResetTimer()
	for range b.N {
		// Spawn with existing archetype (should reuse)
		id := storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()})
		storage.GetComponent(id, posType)
	}
}

func BenchmarkEntityRefCreate(b *testing.B) {
	storage := ecs.NewStorage()

	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)})
	}

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(ids))
		storage.CreateEntityRef(ids[idx])
	}
}

func BenchmarkEntityRefResolve(b *testing.B) {
	storage := ecs.NewStorage()

	ids := make([]ecs.EntityId, 10000)
	refs := make([]ecs.EntityRef, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)})
		refs[i] = storage.CreateEntityRef(ids[i])
	}

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(refs))
		storage.ResolveEntityRef(refs[idx])
	}
}

func BenchmarkAddComponent(b *testing.B) {
	storage := ecs.NewStorage()

	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)})
		storage.CreateEntityRef(ids[i])
	}

	b.ResetTimer()
	for i := range b.N {
		if i >= len(ids) {
			break
		}
		storage.AddComponent(ids[i], &Velocity{DX: 1.0, DY: 1.0})
	}
}

func BenchmarkRemoveComponent(b *testing.B) {
	storage := ecs.NewStorage()

	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)}, &Velocity{DX: 1.0, DY: 1.0})
		storage.CreateEntityRef(ids[i])
	}

	b.ResetTimer()
	for i := range b.N {
		if i >= len(ids) {
			break
		}
		storage.RemoveComponent(ids[i], reflect.TypeOf(Velocity{}))
	}
}

func BenchmarkAddRemoveComponent(b *testing.B) {
	storage := ecs.NewStorage()

	ids := make([]ecs.EntityId, 1000)
	refs := make([]ecs.EntityRef, 1000)
	for i := range ids {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)})
		refs[i] = storage.CreateEntityRef(ids[i])
	}

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(refs))
		id, ok := storage.ResolveEntityRef(refs[idx])
		if !ok {
			continue
		}

		if storage.HasComponent(id, reflect.TypeOf(Velocity{})) {
			storage.RemoveComponent(id, reflect.TypeOf(Velocity{}))
		} else {
			storage.AddComponent(id, &Velocity{DX: 1.0, DY: 1.0})
		}
	}
}

func BenchmarkValueComponents(b *testing.B) {
	storage := ecs.NewStorage()

	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)}, &Velocity{DX: 1.0, DY: 1.0})
	}

	posType := reflect.TypeOf(Position{})

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(ids))
		comp := storage.GetComponent(ids[idx], posType)
		if comp == nil {
			b.Fatal("component is nil")
		}
	}
}

func BenchmarkPointerComponents(b *testing.B) {
	storage := ecs.NewStorage()

	type AI struct {
		Target *Position
	}

	target := &Position{X: 100.0, Y: 200.0}
	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(&AI{Target: target})
	}

	aiType := reflect.TypeOf(AI{})

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(ids))
		comp := storage.GetComponent(ids[idx], aiType)
		if comp == nil {
			b.Fatal("component is nil")
		}
	}
}

func BenchmarkMixedComponents(b *testing.B) {
	storage := ecs.NewStorage()

	type Link struct {
		Next *Name
	}

	next := &Name{Value: "target"}
	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(&Position{X: float32(i), Y: float32(i)}, &Link{Next: next})
	}

	posType := reflect.TypeOf(Position{})
	linkType := reflect.TypeOf(Link{})

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(ids))
		storage.GetComponent(ids[idx], posType)
		storage.GetComponent(ids[idx], linkType)
	}
}

func BenchmarkStorageBySize(b *testing.B) {
	type Tiny struct {
		A uint8
	}

	type Small struct {
		A, B uint32
	}

	type Medium struct {
		A, B, C, D float32
	}

	type Large struct {
		A, B, C, D, E, F, G, H float64
	}

	type Huge struct {
		Data [128]byte
	}

	runBenchmark := func(b *testing.B, component any, name string) {
		b.Run(name, func(b *testing.B) {
			storage := ecs.NewStorage()
			ids := make([]ecs.EntityId, 10000)
			for i := range ids {
				ids[i] = storage.Spawn(component)
			}

			compType := reflect.TypeOf(component).Elem()

			b.ResetTimer()
			for range b.N {
				idx := rand.IntN(len(ids))
				comp := storage.GetComponent(ids[idx], compType)
				if comp == nil {
					b.Fatal("component is nil")
				}
			}
		})
	}

	runBenchmark(b, &Tiny{A: 1}, "1byte")
	runBenchmark(b, &Small{A: 1, B: 2}, "8bytes")
	runBenchmark(b, &Medium{A: 1, B: 2, C: 3, D: 4}, "16bytes")
	runBenchmark(b, &Large{}, "64bytes")
	runBenchmark(b, &Huge{}, "128bytes")
}

func BenchmarkStorageBySizeWithPointers(b *testing.B) {
	type TinyPtr struct {
		A *uint8
	}

	type SmallPtr struct {
		A, B *uint32
	}

	type MediumPtr struct {
		A, B, C, D *float32
	}

	type LargePtr struct {
		A, B, C, D, E, F, G, H *float64
	}

	type HugePtr struct {
		Data []*byte
	}

	val := uint8(1)
	runBenchmark := func(b *testing.B, component any, name string) {
		b.Run(name+"_ptr", func(b *testing.B) {
			storage := ecs.NewStorage()
			ids := make([]ecs.EntityId, 10000)
			for i := range ids {
				ids[i] = storage.Spawn(component)
			}

			compType := reflect.TypeOf(component).Elem()

			b.ResetTimer()
			for range b.N {
				idx := rand.IntN(len(ids))
				comp := storage.GetComponent(ids[idx], compType)
				if comp == nil {
					b.Fatal("component is nil")
				}
			}
		})
	}

	runBenchmark(b, &TinyPtr{A: &val}, "1byte")
	runBenchmark(b, &SmallPtr{}, "8bytes")
	runBenchmark(b, &MediumPtr{}, "16bytes")
	runBenchmark(b, &LargePtr{}, "64bytes")
	runBenchmark(b, &HugePtr{Data: make([]*byte, 16)}, "128bytes")
}

func BenchmarkCacheEfficiency(b *testing.B) {
	type Component8 struct {
		A, B uint32
	}

	type Component8Ptr struct {
		A *uint32
		B *uint32
	}

	b.Run("Sequential_Value", func(b *testing.B) {
		storage := ecs.NewStorage()
		ids := make([]ecs.EntityId, 10000)
		for i := range ids {
			ids[i] = storage.Spawn(&Component8{A: uint32(i), B: uint32(i * 2)})
		}

		compType := reflect.TypeOf(Component8{})

		b.ResetTimer()
		for range b.N {
			sum := uint32(0)
			for _, id := range ids {
				comp := storage.GetComponent(id, compType).(*Component8)
				sum += comp.A + comp.B
			}
			if sum == 0 {
				b.Fatal("sum is zero")
			}
		}
	})

	b.Run("Sequential_Pointer", func(b *testing.B) {
		storage := ecs.NewStorage()
		ids := make([]ecs.EntityId, 10000)
		val := uint32(42)
		for i := range ids {
			ids[i] = storage.Spawn(&Component8Ptr{A: &val, B: &val})
		}

		compType := reflect.TypeOf(Component8Ptr{})

		b.ResetTimer()
		for range b.N {
			sum := uint32(0)
			for _, id := range ids {
				comp := storage.GetComponent(id, compType).(*Component8Ptr)
				if comp.A != nil {
					sum += *comp.A
				}
				if comp.B != nil {
					sum += *comp.B
				}
			}
			if sum == 0 {
				b.Fatal("sum is zero")
			}
		}
	})

	b.Run("Random_Value", func(b *testing.B) {
		storage := ecs.NewStorage()
		ids := make([]ecs.EntityId, 10000)
		for i := range ids {
			ids[i] = storage.Spawn(&Component8{A: uint32(i), B: uint32(i * 2)})
		}

		compType := reflect.TypeOf(Component8{})

		b.ResetTimer()
		for range b.N {
			sum := uint32(0)
			for range 1000 {
				idx := rand.IntN(len(ids))
				comp := storage.GetComponent(ids[idx], compType).(*Component8)
				sum += comp.A + comp.B
			}
			if sum == 0 {
				b.Fatal("sum is zero")
			}
		}
	})

	b.Run("Random_Pointer", func(b *testing.B) {
		storage := ecs.NewStorage()
		ids := make([]ecs.EntityId, 10000)
		val := uint32(42)
		for i := range ids {
			ids[i] = storage.Spawn(&Component8Ptr{A: &val, B: &val})
		}

		compType := reflect.TypeOf(Component8Ptr{})

		b.ResetTimer()
		for range b.N {
			sum := uint32(0)
			for range 1000 {
				idx := rand.IntN(len(ids))
				comp := storage.GetComponent(ids[idx], compType).(*Component8Ptr)
				if comp.A != nil {
					sum += *comp.A
				}
				if comp.B != nil {
					sum += *comp.B
				}
			}
			if sum == 0 {
				b.Fatal("sum is zero")
			}
		}
	})
}

func BenchmarkViewIterCacheEfficiency(b *testing.B) {
	type Component16 struct {
		A, B, C, D float32
	}

	type Component16Ptr struct {
		A *float32
		B *float32
		C *float32
		D *float32
	}

	b.Run("Value_10k", func(b *testing.B) {
		storage := ecs.NewStorage()
		for range 10000 {
			storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Component16{})
		}

		view := ecs.NewView[struct {
			*Position
			*Component16
		}](storage)

		b.ResetTimer()
		for range b.N {
			for _, item := range view.Iter() {
				item.Position.X += 1.0
				item.Component16.A += 1.0
			}
		}
	})

	b.Run("Pointer_10k", func(b *testing.B) {
		storage := ecs.NewStorage()
		val := float32(1.0)
		for range 10000 {
			storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Component16Ptr{A: &val, B: &val, C: &val, D: &val})
		}

		view := ecs.NewView[struct {
			*Position
			*Component16Ptr
		}](storage)

		b.ResetTimer()
		for range b.N {
			for _, item := range view.Iter() {
				item.Position.X += 1.0
				if item.Component16Ptr.A != nil {
					*item.Component16Ptr.A += 1.0
				}
			}
		}
	})
}

func BenchmarkSpawn100kComparison(b *testing.B) {
	b.Run("ValueComponents", func(b *testing.B) {
		for range b.N {
			storage := ecs.NewStorage()
			for range 100000 {
				storage.Spawn(&Position{X: 1.0, Y: 2.0}, &Velocity{DX: 0.5, DY: 0.5})
			}
		}
	})

	b.Run("PointerComponents", func(b *testing.B) {
		type Link struct {
			Next *Position
		}
		next := &Position{X: 1.0, Y: 2.0}
		for range b.N {
			storage := ecs.NewStorage()
			for range 100000 {
				storage.Spawn(&Link{Next: next})
			}
		}
	})

	b.Run("MixedComponents", func(b *testing.B) {
		type AI struct {
			Target *Position
		}
		target := &Position{X: 1.0, Y: 2.0}
		for range b.N {
			storage := ecs.NewStorage()
			for range 100000 {
				storage.Spawn(&Position{X: 1.0, Y: 2.0}, &AI{Target: target})
			}
		}
	})
}

// Example showing basic usage
func ExampleStorage() {
	storage := ecs.NewStorage()

	// Spawn an entity with components
	playerId := storage.Spawn(
		&Position{X: 10.0, Y: 20.0},
		&Velocity{DX: 1.0, DY: 0.5},
		&Health{Current: 100, Max: 100},
	)

	// Get a component
	pos := storage.GetComponent(playerId, reflect.TypeOf(Position{})).(*Position)
	fmt.Printf("Player position: (%.1f, %.1f)\n", pos.X, pos.Y)

	// Modify the component
	pos.X += 5.0

	// Get updated component
	pos = storage.GetComponent(playerId, reflect.TypeOf(Position{})).(*Position)
	fmt.Printf("Updated position: (%.1f, %.1f)\n", pos.X, pos.Y)

	// Output: Player position: (10.0, 20.0)
	// Updated position: (15.0, 20.0)
}
