package ecs

import (
	"iter"
	"reflect"
	"strings"
)

// UniverseView wraps a universe in a struct containing a collection of
// components. This wrapper provides a variety of utility functions for accessing
// and iterating over entities that match the given set of components.
type UniverseView[T any] struct {
	u        *Universe
	types    []reflect.Type
	optional []reflect.Type
	excluded []reflect.Type
}

// View creates a new UniverseView for the given Universe. The provided type
// must be a struct type or this function will panic. It's recommened to re-use
// the result of this function when possible to improve performance.
func View[T any](u *Universe) *UniverseView[T] {
	var t T
	typeT := reflect.TypeOf(t)
	if typeT.Kind() != reflect.Struct {
		panic("ecs.View() requires a struct-type")
	}

	types := []reflect.Type{}
	optional := []reflect.Type{}
	excluded := []reflect.Type{}
	for i := 0; i < typeT.NumField(); i++ {
		tag := typeT.Field(i).Tag.Get("ecs")

		isOptional := false
		isExcluded := false
		if tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if part == "optional" {
					isOptional = true
				} else if part == "excluded" {
					isExcluded = true
				} else {
					panic("unsupported ecs tag: " + part)
				}
			}
		}
		if isExcluded {
			excluded = append(excluded, typeT.Field(i).Type)
		} else if isOptional {
			optional = append(optional, typeT.Field(i).Type)
		} else {
			types = append(types, typeT.Field(i).Type)
		}
	}

	return &UniverseView[T]{u: u, types: types, optional: optional, excluded: excluded}
}

// Get returns data for the given entity. If the entity doesn't exist the
// component pointers will be nil. If the entity doesn't have a given component
// it will be nil. For a safer version of this function see `MaybeGet`.
func (v *UniverseView[T]) Get(id EntityId) T {
	var t T
	data := v.u.storage.Get(id)
	data.Fill(&t)
	return t
}

// MaybeGet returns data for the given entity, checking both if the entity exists
// and has all of the required component types. If the entity does not exist all
// component pointer will be nil and the second return value will be false. If
// some components are missing from the target entity, only those pointers will
// be nil however the second return value will still be false.
func (v *UniverseView[T]) MaybeGet(id EntityId) (T, bool) {
	var t T
	data := v.u.storage.Get(id)
	if data == nil {
		return t, false
	}
	if !data.Fill(&t) {
		return t, false
	}
	return t, true
}

// Iter returns an iterator over all entities that match this view.
func (v *UniverseView[T]) Iter() iter.Seq[T] {
	filter := EntityFilter{
		ComponentTypes:         v.types,
		ExcludedComponentTypes: v.excluded,
	}
	return func(yield func(T) bool) {
		var t T
		for data := range v.u.storage.Filter(filter) {
			data.Fill(&t)
			if !yield(t) {
				break
			}
		}
	}
}

// Spawn creates a new entity based on the components in the view.
func (v *UniverseView[T]) Spawn(value T) EntityId {
	valueValue := reflect.ValueOf(value)
	values := make([]any, 0, valueValue.NumField())

	for i := 0; i < valueValue.NumField(); i++ {
		values = append(values, valueValue.Field(i).Interface())
	}

	return v.u.Spawn(values...)
}
