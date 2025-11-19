package ecs

import (
	"reflect"
	"sort"
	"unsafe"
)

// Storage is the main ECS storage interface
type Storage struct {
	archetypes map[uint32]*Archetype
}

// NewStorage creates a new ECS storage system
func NewStorage() *Storage {
	return &Storage{
		archetypes: make(map[uint32]*Archetype),
	}
}

// Spawn creates a new entity with the provided components
func (s *Storage) Spawn(components ...any) EntityId {
	if len(components) == 0 {
		panic("cannot spawn entity without components")
	}

	types := extractComponentTypes(components)
	archetypeId := hashTypesToUint32(types)

	archetype, exists := s.archetypes[archetypeId]
	if !exists {
		archetype = NewArchetype(archetypeId, types)
		s.archetypes[archetypeId] = archetype
	}

	entityIndex := archetype.Spawn(components)
	return NewEntityId(archetypeId, entityIndex)
}

// Delete removes all data related to the entity ID
func (s *Storage) Delete(id EntityId) {
	archetypeId := id.ArchetypeId()
	entityIndex := id.Index()

	archetype, ok := s.archetypes[archetypeId]
	if !ok {
		return
	}

	archetype.Delete(entityIndex)
}

// GetComponent returns the component for the given entity ID and component type
func (s *Storage) GetComponent(id EntityId, compType reflect.Type) any {
	archetypeId := id.ArchetypeId()
	entityIndex := id.Index()

	archetype, ok := s.archetypes[archetypeId]
	if !ok {
		return nil
	}

	return archetype.GetComponent(entityIndex, compType)
}

// HasComponent checks if an entity has a specific component type
func (s *Storage) HasComponent(id EntityId, compType reflect.Type) bool {
	archetypeId := id.ArchetypeId()
	archetype, ok := s.archetypes[archetypeId]
	if !ok {
		return false
	}
	return archetype.HasComponent(compType)
}

// extractComponentTypes extracts and sorts component types from a slice of components
func extractComponentTypes(components []any) []reflect.Type {
	types := make([]reflect.Type, 0, len(components))
	for _, comp := range components {
		compType := reflect.TypeOf(comp)

		// If it's a pointer, get the underlying type
		if compType.Kind() == reflect.Ptr {
			compType = compType.Elem()
		}

		// Components can be structs or primitives (int, string, etc.)
		// But not pointers, maps, channels, or functions (those aren't value types)
		if compType.Kind() == reflect.Ptr || compType.Kind() == reflect.Map ||
			compType.Kind() == reflect.Chan || compType.Kind() == reflect.Func {
			panic("components cannot be pointers, maps, channels, or functions")
		}

		types = append(types, compType)
	}
	sort.Sort(byTypeName(types))
	return types
}

// hashTypesToUint32 generates a uint32 hash for a sorted slice of types
func hashTypesToUint32(types []reflect.Type) uint32 {
	var h uint32 = 2166136261     // FNV-1a 32-bit offset basis
	const prime uint32 = 16777619 // FNV-1a 32-bit prime

	for _, t := range types {
		// Use the type's pointer as a unique identifier
		ptr := (*iface)(unsafe.Pointer(&t)).data
		val := uint32(uintptr(ptr))

		// Mix in all 4 bytes if on 64-bit system
		if unsafe.Sizeof(uintptr(0)) == 8 {
			val ^= uint32(uintptr(ptr) >> 32)
		}

		h ^= val
		h *= prime
	}

	return h
}
