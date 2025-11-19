package ecs

import (
	"iter"
	"reflect"
)

// View represents a query for entities with a specific combination of components
// The type T should be a struct with embedded pointer fields for each component type
// Named fields can be marked as optional using the `ecs:"optional"` struct tag
type View[T any] struct {
	types    []reflect.Type
	optional []bool // Tracks which components are optional
}

// NewView creates a new view for the given struct type
// The struct T should have embedded or named fields that are pointers to component types
// Embedded fields are always required
// Named fields can be marked as optional using the `ecs:"optional"` struct tag
func NewView[T any]() *View[T] {
	var zero T
	structType := reflect.TypeOf(zero)

	if structType.Kind() != reflect.Struct {
		panic("View type parameter must be a struct")
	}

	types := make([]reflect.Type, 0, structType.NumField())
	optional := make([]bool, 0, structType.NumField())

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldType := field.Type

		if fieldType.Kind() != reflect.Ptr {
			panic("View struct fields must be pointer types")
		}

		componentType := fieldType.Elem()
		types = append(types, componentType)

		// Parse struct tag to check if component is optional
		// Embedded fields (field.Anonymous) are always required
		isOptional := false
		if !field.Anonymous {
			tag := field.Tag.Get("ecs")
			if tag != "" {
				if tag == "optional" {
					isOptional = true
				} else {
					panic("invalid ecs tag value: \"" + tag + "\" (only \"optional\" is supported)")
				}
			}
		}
		optional = append(optional, isOptional)
	}

	return &View[T]{
		types:    types,
		optional: optional,
	}
}

// Fill populates the provided struct pointer with component data for the given entity
// Returns false if the entity is missing any required components
// Optional components are set to nil if not present
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
			// If this is a required component, fail
			if !v.optional[i] {
				return false
			}
			// Optional component is missing, set field to nil
			field.Set(reflect.Zero(field.Type()))
		} else {
			// Component found, set the field
			field.Set(reflect.ValueOf(component))
		}
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

// matchesArchetype checks if an archetype contains all the required component types for this view
// Optional components are not checked - they may or may not be present
func (v *View[T]) matchesArchetype(archetype *Archetype) bool {
	for i, requiredType := range v.types {
		// Skip optional components
		if v.optional[i] {
			continue
		}
		// Required component must be present
		if !archetype.HasComponent(requiredType) {
			return false
		}
	}
	return true
}

// Iter returns an iterator over all entities that have all the required components for this view
// The iterator yields (EntityId, T) pairs where T is the populated view struct
// Optional components are set to nil if not present
func (v *View[T]) Iter(storage *Storage) iter.Seq2[EntityId, T] {
	return func(yield func(EntityId, T) bool) {
		for archetypeId, archetype := range storage.archetypes {
			if !v.matchesArchetype(archetype) {
				continue
			}

			// Pre-compute the mapping from view component types to archetype storage indices
			storageIndices := make([]int, len(v.types))
			for i, componentType := range v.types {
				storageIndices[i] = -1
				for idx, archetypeType := range archetype.types {
					if archetypeType == componentType {
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

					// Populate all components
					allRequiredComponentsFound := true
					for i, storageIdx := range storageIndices {
						if storageIdx == -1 {
							if v.optional[i] {
								resultValue.Field(i).Set(reflect.Zero(resultValue.Field(i).Type()))
								continue
							} else {
								allRequiredComponentsFound = false
								break
							}
						}

						component := archetype.storages[storageIdx].Get(int(entityIndex))
						if component == nil {
							if v.optional[i] {
								resultValue.Field(i).Set(reflect.Zero(resultValue.Field(i).Type()))
								continue
							} else {
								allRequiredComponentsFound = false
								break
							}
						}

						// Set the field directly
						resultValue.Field(i).Set(reflect.ValueOf(component))
					}

					if !allRequiredComponentsFound {
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
