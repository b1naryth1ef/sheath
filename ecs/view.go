package ecs

import (
	"iter"
	"reflect"
)

// View represents a query for entities with a specific combination of components
// The type T should be a struct with embedded pointer fields for each component type
type View[T any] struct {
	types []reflect.Type
}

// NewView creates a new view for the given struct type
// The struct T should have embedded fields that are pointers to component types
func NewView[T any]() *View[T] {
	var zero T
	structType := reflect.TypeOf(zero)

	if structType.Kind() != reflect.Struct {
		panic("View type parameter must be a struct")
	}

	types := make([]reflect.Type, 0, structType.NumField())

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldType := field.Type

		if fieldType.Kind() != reflect.Ptr {
			panic("View struct fields must be pointer types")
		}

		componentType := fieldType.Elem()
		types = append(types, componentType)
	}

	return &View[T]{types: types}
}

// Fill populates the provided struct pointer with component data for the given entity
func (v *View[T]) Fill(storage *Storage, id EntityId, ptr *T) bool {
	archetypeId := id.ArchetypeId()
	archetype, ok := storage.archetypes[archetypeId]
	if !ok {
		return false
	}

	structValue := reflect.ValueOf(ptr).Elem()
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structValue.Field(i)
		componentType := v.types[i]

		component := archetype.GetComponent(id.Index(), componentType)
		if component == nil {
			return false
		}

		field.Set(reflect.ValueOf(component))
	}

	return true
}

// Get returns a populated view struct for the given entity, or nil if the entity
// doesn't have all the required components
func (v *View[T]) Get(storage *Storage, id EntityId) *T {
	var result T
	if !v.Fill(storage, id, &result) {
		return nil
	}
	return &result
}

// matchesArchetype checks if an archetype contains all the component types required by this view
func (v *View[T]) matchesArchetype(archetype *Archetype) bool {
	for _, requiredType := range v.types {
		if !archetype.HasComponent(requiredType) {
			return false
		}
	}
	return true
}

// Iter returns an iterator over all entities that have all the components required by this view
// The iterator yields (EntityId, T) pairs where T is the populated view struct
func (v *View[T]) Iter(storage *Storage) iter.Seq2[EntityId, T] {
	return func(yield func(EntityId, T) bool) {
		for archetypeId, archetype := range storage.archetypes {
			if !v.matchesArchetype(archetype) {
				continue
			}

			// Pre-compute the mapping from view component types to archetype storage indices
			storageIndices := make([]int, len(v.types))
			for i, requiredType := range v.types {
				storageIndices[i] = -1
				for idx, archetypeType := range archetype.types {
					if archetypeType == requiredType {
						storageIndices[i] = idx
						break
					}
				}
			}

			// Get the first component storage to determine entity count
			// All storages in an archetype have the same capacity and indices
			firstStorage := archetype.storages[0]

			// Pre-allocate result struct value for reflection operations
			var result T
			resultValue := reflect.ValueOf(&result).Elem()

			for blockIdx, block := range firstStorage.blocks {
				for slotIdx := range blockSize {
					mask := uint64(1) << slotIdx
					if block.filled&mask == 0 {
						continue
					}

					entityIndex := uint32(blockIdx*blockSize + slotIdx)
					entityId := NewEntityId(archetypeId, entityIndex)

					allComponentsFound := true
					for i, storageIdx := range storageIndices {
						if storageIdx == -1 {
							allComponentsFound = false
							break
						}

						component := archetype.storages[storageIdx].Get(int(entityIndex))
						if component == nil {
							allComponentsFound = false
							break
						}

						// Set the field directly
						resultValue.Field(i).Set(reflect.ValueOf(component))
					}

					if !allComponentsFound {
						continue
					}

					// Yield the entity ID and view struct
					if !yield(entityId, result) {
						return
					}
				}
			}
		}
	}
}

// IterValues returns an iterator over just the view structs (without entity IDs)
// This is useful when you only care about the component data, not which entity it belongs to
func (v *View[T]) IterValues(storage *Storage) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, value := range v.Iter(storage) {
			if !yield(value) {
				return
			}
		}
	}
}
