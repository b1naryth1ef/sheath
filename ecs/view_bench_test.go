package ecs_test

import (
	"math/rand/v2"
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
)

// BenchmarkViewCreation tests the cost of creating a view
func BenchmarkViewCreation(b *testing.B) {
	b.Run("TwoComponents", func(b *testing.B) {
		for range b.N {
			_ = ecs.NewView[struct {
				*Position
				*Velocity
			}]()
		}
	})

	b.Run("FourComponents", func(b *testing.B) {
		for range b.N {
			_ = ecs.NewView[struct {
				*Position
				*Velocity
				*Health
				*Name
			}]()
		}
	})
}

// BenchmarkViewGet tests single entity retrieval performance
func BenchmarkViewGet(b *testing.B) {
	storage := ecs.NewStorage()
	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Pre-populate with entities
	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(
			&Position{X: float32(i), Y: float32(i)},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
		)
	}

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(ids))
		item := view.Get(storage, ids[idx])
		if item == nil {
			b.Fatal("item is nil")
		}
	}
}

// BenchmarkViewFill tests the Fill method performance
func BenchmarkViewFill(b *testing.B) {
	storage := ecs.NewStorage()
	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	// Pre-populate with entities
	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(
			&Position{X: float32(i), Y: float32(i)},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
		)
	}

	var result struct {
		*Position
		*Velocity
	}

	b.ResetTimer()
	for range b.N {
		idx := rand.IntN(len(ids))
		if !view.Fill(storage, ids[idx], &result) {
			b.Fatal("fill failed")
		}
	}
}

// BenchmarkViewIter tests iteration over entities
func BenchmarkViewIter(b *testing.B) {
	b.Run("100Entities", func(b *testing.B) {
		storage := ecs.NewStorage()
		for range 100 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			count := 0
			for range view.Iter(storage) {
				count++
			}
			if count != 100 {
				b.Fatalf("expected 100, got %d", count)
			}
		}
	})

	b.Run("1000Entities", func(b *testing.B) {
		storage := ecs.NewStorage()
		for range 1000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			count := 0
			for range view.Iter(storage) {
				count++
			}
			if count != 1000 {
				b.Fatalf("expected 1000, got %d", count)
			}
		}
	})

	b.Run("10000Entities", func(b *testing.B) {
		storage := ecs.NewStorage()
		for range 10000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			count := 0
			for range view.Iter(storage) {
				count++
			}
			if count != 10000 {
				b.Fatalf("expected 10000, got %d", count)
			}
		}
	})

	b.Run("100000Entities", func(b *testing.B) {
		storage := ecs.NewStorage()
		for range 100000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			count := 0
			for range view.Iter(storage) {
				count++
			}
			if count != 100000 {
				b.Fatalf("expected 100000, got %d", count)
			}
		}
	})
}

// BenchmarkViewIterValues tests values-only iteration
func BenchmarkViewIterValues(b *testing.B) {
	storage := ecs.NewStorage()
	for range 10000 {
		storage.Spawn(
			&Position{X: rand.Float32(), Y: rand.Float32()},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
		)
	}

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	b.ResetTimer()
	for range b.N {
		count := 0
		for range view.IterValues(storage) {
			count++
		}
		if count != 10000 {
			b.Fatalf("expected 10000, got %d", count)
		}
	}
}

// BenchmarkViewIterWithMutation tests iteration with component updates
func BenchmarkViewIterWithMutation(b *testing.B) {
	storage := ecs.NewStorage()
	for range 10000 {
		storage.Spawn(
			&Position{X: rand.Float32(), Y: rand.Float32()},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
		)
	}

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	b.ResetTimer()
	for range b.N {
		for _, item := range view.Iter(storage) {
			item.Position.X += item.Velocity.DX
			item.Position.Y += item.Velocity.DY
		}
	}
}

// BenchmarkViewIterMultipleArchetypes tests iteration across different archetypes
func BenchmarkViewIterMultipleArchetypes(b *testing.B) {
	b.Run("2Archetypes", func(b *testing.B) {
		storage := ecs.NewStorage()

		// Archetype 1: Position + Velocity
		for range 5000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
		}

		// Archetype 2: Position + Velocity + Health
		for range 5000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
				&Health{Current: 100, Max: 100},
			)
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			count := 0
			for range view.Iter(storage) {
				count++
			}
			if count != 10000 {
				b.Fatalf("expected 10000, got %d", count)
			}
		}
	})

	b.Run("5Archetypes", func(b *testing.B) {
		storage := ecs.NewStorage()

		// Create entities across 5 different archetypes, all containing Position + Velocity
		for range 2000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
		}
		for range 2000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
				&Health{Current: 100, Max: 100},
			)
		}
		for range 2000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
				&Name{Value: "Entity"},
			)
		}
		for range 2000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
				Score(100),
			)
		}
		for range 2000 {
			storage.Spawn(
				&Position{X: rand.Float32(), Y: rand.Float32()},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
				&Health{Current: 100, Max: 100},
				&Name{Value: "Entity"},
			)
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			count := 0
			for range view.Iter(storage) {
				count++
			}
			if count != 10000 {
				b.Fatalf("expected 10000, got %d", count)
			}
		}
	})

	b.Run("10Archetypes", func(b *testing.B) {
		storage := ecs.NewStorage()

		// Create entities across 10 different archetypes
		archetypeFactories := []func(){
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()})
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, &Health{Current: 100, Max: 100})
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, &Name{Value: "Entity"})
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, Score(100))
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, Temperature(98.6))
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, &Health{Current: 100, Max: 100}, &Name{Value: "Entity"})
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, &Health{Current: 100, Max: 100}, Score(100))
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, &Name{Value: "Entity"}, Score(100))
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, Temperature(98.6), Score(100))
			},
			func() {
				storage.Spawn(&Position{X: rand.Float32(), Y: rand.Float32()}, &Velocity{DX: rand.Float32(), DY: rand.Float32()}, &Health{Current: 100, Max: 100}, &Name{Value: "Entity"}, Score(100))
			},
		}

		for _, factory := range archetypeFactories {
			for range 1000 {
				factory()
			}
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			count := 0
			for range view.Iter(storage) {
				count++
			}
			if count != 10000 {
				b.Fatalf("expected 10000, got %d", count)
			}
		}
	})
}

// BenchmarkViewIterWithFilter simulates a common pattern of filtering during iteration
func BenchmarkViewIterWithFilter(b *testing.B) {
	storage := ecs.NewStorage()
	for i := range 10000 {
		storage.Spawn(
			&Position{X: float32(i), Y: float32(i)},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			&Health{Current: rand.IntN(100), Max: 100},
		)
	}

	view := ecs.NewView[struct {
		*Position
		*Velocity
		*Health
	}]()

	b.ResetTimer()
	for range b.N {
		count := 0
		for _, item := range view.Iter(storage) {
			// Filter: only process entities with low health
			if item.Health.Current < 50 {
				item.Position.X += item.Velocity.DX * 2 // Move faster when injured
				item.Position.Y += item.Velocity.DY * 2
				count++
			}
		}
	}
}

// BenchmarkViewIterSparseEntities tests iteration with many deleted entities
func BenchmarkViewIterSparseEntities(b *testing.B) {
	storage := ecs.NewStorage()

	// Create entities
	ids := make([]ecs.EntityId, 10000)
	for i := range ids {
		ids[i] = storage.Spawn(
			&Position{X: float32(i), Y: float32(i)},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
		)
	}

	// Delete 50% of entities (create sparsity)
	for i := 0; i < len(ids); i += 2 {
		storage.Delete(ids[i])
	}

	view := ecs.NewView[struct {
		*Position
		*Velocity
	}]()

	b.ResetTimer()
	for range b.N {
		count := 0
		for range view.Iter(storage) {
			count++
		}
		if count != 5000 {
			b.Fatalf("expected 5000, got %d", count)
		}
	}
}

// BenchmarkViewIterComplex tests iteration with many component types
func BenchmarkViewIterComplex(b *testing.B) {
	storage := ecs.NewStorage()

	for range 5000 {
		storage.Spawn(
			&Position{X: rand.Float32(), Y: rand.Float32()},
			&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			&Health{Current: rand.IntN(100), Max: 100},
			&Name{Value: "Entity"},
		)
	}

	view := ecs.NewView[struct {
		*Position
		*Velocity
		*Health
		*Name
	}]()

	b.ResetTimer()
	for range b.N {
		for _, item := range view.Iter(storage) {
			item.Position.X += item.Velocity.DX
			item.Position.Y += item.Velocity.DY
			if item.Position.X > 100 {
				item.Health.Current--
			}
		}
	}
}

// BenchmarkViewVsDirectAccess compares view iteration to direct component access
func BenchmarkViewVsDirectAccess(b *testing.B) {
	b.Run("ViewIteration", func(b *testing.B) {
		storage := ecs.NewStorage()
		ids := make([]ecs.EntityId, 1000)
		for i := range ids {
			ids[i] = storage.Spawn(
				&Position{X: float32(i), Y: float32(i)},
				&Velocity{DX: rand.Float32(), DY: rand.Float32()},
			)
		}

		view := ecs.NewView[struct {
			*Position
			*Velocity
		}]()

		b.ResetTimer()
		for range b.N {
			for _, item := range view.Iter(storage) {
				item.Position.X += item.Velocity.DX
				item.Position.Y += item.Velocity.DY
			}
		}
	})
}
