package ecs

import "reflect"

// View represents a query for entities with a specific combination of components
// The type T should be a struct with embedded pointer fields for each component type
type View[T any] struct {
	types []reflect.Type
}

// NewView creates a new view for the given struct type
// The struct T should have embedded fields that are pointers to component types
// Example: struct { *Position; *Velocity }
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
