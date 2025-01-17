package ecs

import (
	"iter"
	"reflect"
)

// EntityStorage is the interface which describes the underlying storage of
// entities and their associated components. You could implement this with a
// custom storage implementation based on your needs.
type EntityStorage interface {
	Get(EntityId) EntityData
	Create(...any) EntityId
	Delete(EntityId)
	Filter(EntityFilter) iter.Seq[EntityData]
}

// EntityData is designed to be a reference based wrapper around storage-data for
// an entity. In the simple case this could just be a struct which stores the
// required data in memory, however in more complex implementations one could
// copy component data from a variety of memory locations.
type EntityData interface {
	Id() EntityId
	GetComponents(...any) bool
	HasComponents(...reflect.Type) bool
	Fill(any) bool
}

// EntityFilter describes a filter that can be applied when iterating over entities
type EntityFilter struct {
	ComponentTypes []reflect.Type
}

// WithComponents creates a copy of this filter with the given components included
// as required.
func (e EntityFilter) WithComponents(components ...any) EntityFilter {
	additionalTypes := make([]reflect.Type, 0, len(components))
	for _, comp := range components {
		additionalTypes = append(additionalTypes, reflect.TypeOf(comp))
	}
	return e.WithComponentTypes(additionalTypes...)
}

// WithComponentTypes creates a copy of this filter with the given component types
// included as required.
func (e EntityFilter) WithComponentTypes(componentTypes ...reflect.Type) EntityFilter {
	return EntityFilter{
		ComponentTypes: append(e.ComponentTypes, componentTypes...),
	}
}

// WithView creates a copy of this filter with the given view structs component
// types included as required.
func (e EntityFilter) WithView(view any) EntityFilter {
	types := []reflect.Type{}
	viewType := reflect.TypeOf(view).Elem()
	for i := 0; i < viewType.NumField(); i++ {
		types = append(types, viewType.Field(i).Type)
	}
	return e.WithComponentTypes(types...)
}

// Exec applies this filter to the given entity data returning true if it matches.
func (e EntityFilter) Exec(target EntityData) bool {
	if e.ComponentTypes != nil && !target.HasComponents(e.ComponentTypes...) {
		return false
	}

	return true
}
