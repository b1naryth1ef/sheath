package ecs

import (
	"reflect"
	"slices"

	"github.com/kamstrup/intmap"
)

type byTypeName []reflect.Type

func (a byTypeName) Len() int           { return len(a) }
func (a byTypeName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTypeName) Less(i, j int) bool { return a[i].String() < a[j].String() }

// Archetype represents a unique combination of component types
type Archetype struct {
	id       uint32
	types    []reflect.Type
	storages []*ComponentStorage
	refs     *intmap.Map[EntityId, *reference]
}

// NewArchetype creates a new archetype with the given ID and sorted component types
func NewArchetype(id uint32, types []reflect.Type) *Archetype {
	a := &Archetype{
		id:       id,
		types:    types,
		storages: make([]*ComponentStorage, len(types)),
		refs:     intmap.New[EntityId, *reference](256),
	}

	// Initialize storage for each component type
	for idx, typ := range types {
		a.storages[idx] = NewComponentStorage(typ, 256)
	}

	return a
}

// Spawn creates a new entity in this archetype with the given components
// Returns the storage position as the entity index
func (a *Archetype) Spawn(components []any) uint32 {
	var storagePos int
	for _, comp := range components {
		compType := reflect.TypeOf(comp)
		if compType.Kind() == reflect.Ptr {
			compType = compType.Elem()
		}

		for idx, typ := range a.types {
			if typ == compType {
				storagePos = a.storages[idx].Append(comp)
			}
		}
	}

	return uint32(storagePos)
}

// GetComponent returns the component of the given type for the entity at entityIndex
// The entityIndex is the storage position directly
func (a *Archetype) GetComponent(entityIndex uint32, compType reflect.Type) any {
	var idx int = -1
	for i, typ := range a.types {
		if typ == compType {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}

	return a.storages[idx].Get(int(entityIndex))
}

// Delete marks an entity's components as deleted
// Indices remain stable - the slot is simply marked as empty
func (a *Archetype) Delete(entityIndex uint32) {
	entityId := NewEntityId(a.id, entityIndex)

	ref, ok := a.refs.Get(entityId)
	if ok {
		ref.deleted = true
		a.refs.Del(entityId)
	}

	for _, storage := range a.storages {
		storage.Delete(int(entityIndex))
	}
}

// HasComponent checks if this archetype has the given component type
func (a *Archetype) HasComponent(compType reflect.Type) bool {
	return slices.Contains(a.types, compType)
}

// ID returns the archetype's unique identifier
func (a *Archetype) ID() uint32 {
	return a.id
}

// Types returns the sorted component types for this archetype
func (a *Archetype) Types() []reflect.Type {
	return a.types
}

// Compact reorganizes all component storage to eliminate empty slots and reduce fragmentation
// EntityRefs remain valid and are automatically updated to point to the new indices
func (a *Archetype) Compact() {
	if len(a.storages) == 0 {
		return
	}

	// Compact the first storage and use it as the canonical index mapping
	indexMap := a.storages[0].Compact()
	for i := 1; i < len(a.storages); i++ {
		a.storages[i].Compact()
	}

	for oldId, newId := range indexMap {
		ref, ok := a.refs.Get(NewEntityId(a.id, uint32(oldId)))
		if ok {
			ref.current = NewEntityId(a.id, uint32(newId))
		}
	}
}
