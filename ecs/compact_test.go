package ecs

import (
	"reflect"
	"testing"
)

type Position struct {
	X, Y float64
}

type Velocity struct {
	DX, DY float64
}

func TestComponentStorageCompact_NoGaps(t *testing.T) {
	// Test compacting storage with no gaps
	cs := NewComponentStorage(reflect.TypeOf(Position{}), 64)

	// Add 3 components with no deletions
	idx0 := cs.Append(Position{X: 1.0, Y: 2.0})
	idx1 := cs.Append(Position{X: 3.0, Y: 4.0})
	idx2 := cs.Append(Position{X: 5.0, Y: 6.0})

	// Compact should not change anything
	indexMap := cs.Compact()

	// Verify mapping
	if len(indexMap) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(indexMap))
	}
	if indexMap[idx0] != idx0 || indexMap[idx1] != idx1 || indexMap[idx2] != idx2 {
		t.Errorf("Expected identity mapping, got %v", indexMap)
	}

	// Verify data is still accessible at same indices
	pos0 := cs.Get(idx0).(*Position)
	if pos0.X != 1.0 || pos0.Y != 2.0 {
		t.Errorf("Position 0 data corrupted: %+v", pos0)
	}

	pos1 := cs.Get(idx1).(*Position)
	if pos1.X != 3.0 || pos1.Y != 4.0 {
		t.Errorf("Position 1 data corrupted: %+v", pos1)
	}

	pos2 := cs.Get(idx2).(*Position)
	if pos2.X != 5.0 || pos2.Y != 6.0 {
		t.Errorf("Position 2 data corrupted: %+v", pos2)
	}
}

func TestComponentStorageCompact_WithGaps(t *testing.T) {
	cs := NewComponentStorage(reflect.TypeOf(Position{}), 64)

	// Add 5 components
	idx0 := cs.Append(Position{X: 1.0, Y: 2.0})
	idx1 := cs.Append(Position{X: 3.0, Y: 4.0})
	idx2 := cs.Append(Position{X: 5.0, Y: 6.0})
	idx3 := cs.Append(Position{X: 7.0, Y: 8.0})
	idx4 := cs.Append(Position{X: 9.0, Y: 10.0})

	// Delete indices 1 and 3 to create gaps
	cs.Delete(idx1)
	cs.Delete(idx3)

	// Now we should have: [0: valid, 1: empty, 2: valid, 3: empty, 4: valid]
	// After compact: [0, 1, 2] with data from old [0, 2, 4]

	indexMap := cs.Compact()

	// Verify mapping size
	if len(indexMap) != 3 {
		t.Errorf("Expected 3 mappings (3 valid components), got %d", len(indexMap))
	}

	// Verify mappings
	if indexMap[idx0] != 0 {
		t.Errorf("Expected idx0 (%d) -> 0, got %d", idx0, indexMap[idx0])
	}
	if indexMap[idx2] != 1 {
		t.Errorf("Expected idx2 (%d) -> 1, got %d", idx2, indexMap[idx2])
	}
	if indexMap[idx4] != 2 {
		t.Errorf("Expected idx4 (%d) -> 2, got %d", idx4, indexMap[idx4])
	}

	// Verify data at new indices
	pos0 := cs.Get(indexMap[idx0]).(*Position)
	if pos0.X != 1.0 || pos0.Y != 2.0 {
		t.Errorf("Position at new index 0 incorrect: %+v", pos0)
	}

	pos2 := cs.Get(indexMap[idx2]).(*Position)
	if pos2.X != 5.0 || pos2.Y != 6.0 {
		t.Errorf("Position at new index 1 incorrect: %+v", pos2)
	}

	pos4 := cs.Get(indexMap[idx4]).(*Position)
	if pos4.X != 9.0 || pos4.Y != 10.0 {
		t.Errorf("Position at new index 2 incorrect: %+v", pos4)
	}

	// Verify we only have 3 filled slots after compaction (at indices 0, 1, 2)
	filledCount := 0
	for i := range 64 {
		if cs.Has(i) {
			filledCount++
		}
	}
	if filledCount != 3 {
		t.Errorf("Expected 3 filled slots after compact, got %d", filledCount)
	}

	// Verify the deleted indices are not in the mapping
	// (old indices idx1 and idx3 should not appear as keys in indexMap)
	if _, exists := indexMap[idx1]; exists {
		t.Errorf("Deleted index %d should not be in index map", idx1)
	}
	if _, exists := indexMap[idx3]; exists {
		t.Errorf("Deleted index %d should not be in index map", idx3)
	}
}

func TestComponentStorageCompact_MultipleBlocks(t *testing.T) {
	cs := NewComponentStorage(reflect.TypeOf(Position{}), 64)

	// Add 130 components (spanning 3 blocks: 64 + 64 + 2)
	indices := make([]int, 130)
	for i := 0; i < 130; i++ {
		indices[i] = cs.Append(Position{X: float64(i), Y: float64(i * 2)})
	}

	// Delete every other component
	for i := 1; i < 130; i += 2 {
		cs.Delete(indices[i])
	}

	// Should have 65 components remaining
	indexMap := cs.Compact()

	if len(indexMap) != 65 {
		t.Errorf("Expected 65 mappings, got %d", len(indexMap))
	}

	// Verify all remaining components have correct data
	expectedIdx := 0
	for i := 0; i < 130; i += 2 {
		newIdx := indexMap[indices[i]]
		if newIdx != expectedIdx {
			t.Errorf("Expected component at old index %d to map to %d, got %d", indices[i], expectedIdx, newIdx)
		}

		pos := cs.Get(newIdx).(*Position)
		if pos.X != float64(i) || pos.Y != float64(i*2) {
			t.Errorf("Component data corrupted at new index %d: expected {%f, %f}, got %+v", newIdx, float64(i), float64(i*2), pos)
		}

		expectedIdx++
	}
}

func TestComponentStorageCompact_Empty(t *testing.T) {
	cs := NewComponentStorage(reflect.TypeOf(Position{}), 64)

	// Compact empty storage
	indexMap := cs.Compact()

	if len(indexMap) != 0 {
		t.Errorf("Expected empty mapping, got %d entries", len(indexMap))
	}

	// Should still have at least one block
	if len(cs.blocks) == 0 {
		t.Error("Expected at least one block after compacting empty storage")
	}
}
