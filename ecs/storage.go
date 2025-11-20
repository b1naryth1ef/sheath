package ecs

import (
	"reflect"
	"sort"
	"unsafe"

	"github.com/kamstrup/intmap"
)

type reference struct {
	ref     EntityRef
	current EntityId
	deleted bool
}

// Storage is the main ECS storage interface
type Storage struct {
	archetypes map[uint32]*Archetype
	refs       *intmap.Map[EntityRef, *reference]

	refId EntityRef
}

// NewStorage creates a new ECS storage system
func NewStorage() *Storage {
	return &Storage{
		archetypes: make(map[uint32]*Archetype),
		refs:       intmap.New[EntityRef, *reference](256),
	}
}

func (s *Storage) CreateEntityRef(id EntityId) EntityRef {
	existing, ok := s.archetypes[id.ArchetypeId()].refs.Get(id)
	if ok {
		return existing.ref
	}

	refId := s.refId
	s.refId++

	ref := &reference{
		ref:     refId,
		current: id,
		deleted: false,
	}

	s.refs.Put(refId, ref)
	s.archetypes[id.ArchetypeId()].refs.Put(id, ref)

	return refId
}

func (s *Storage) ResolveEntityRef(refId EntityRef) (EntityId, bool) {
	ref, ok := s.refs.Get(refId)
	if !ok {
		return 0, false
	}
	if ref.deleted {
		return 0, false
	}
	return ref.current, true
}

func (s *Storage) InvalidateEntityRef(refId EntityRef) bool {
	ref, ok := s.refs.Get(refId)
	if !ok {
		return false
	}
	s.archetypes[ref.current.ArchetypeId()].refs.Del(ref.current)
	s.refs.Del(refId)
	return true
}

// GetArchetype returns an archetype storage (if one exists)
func (s *Storage) GetArchetype(components ...any) *Archetype {
	types := extractComponentTypes(components)
	archetypeId := hashTypesToUint32(types)
	return s.archetypes[archetypeId]
}

// GetArchetypeByTypes returns an archetype storage (if one exists) based on reflect.Type
func (s *Storage) GetArchetypeByTypes(types []reflect.Type) *Archetype {
	sort.Sort(byTypeName(types))
	archetypeId := hashTypesToUint32(types)
	return s.archetypes[archetypeId]
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

func (s *Storage) AddComponent(id EntityId, component any) EntityId {
	oldArchetype := s.archetypes[id.ArchetypeId()]

	compType := reflect.TypeOf(component)
	if compType.Kind() == reflect.Ptr {
		compType = compType.Elem()
	}

	newTypes := make([]reflect.Type, 0, len(oldArchetype.types)+1)
	newTypes = append(newTypes, oldArchetype.types...)
	newTypes = append(newTypes, compType)
	sort.Sort(byTypeName(newTypes))

	newArchetypeId := hashTypesToUint32(newTypes)
	newArchetype, exists := s.archetypes[newArchetypeId]
	if !exists {
		newArchetype = NewArchetype(newArchetypeId, newTypes)
		s.archetypes[newArchetypeId] = newArchetype
	}

	ref, hasRef := oldArchetype.refs.Get(id)

	components := make([]any, 0, len(newTypes))
	for _, typ := range newTypes {
		if typ == compType {
			components = append(components, component)
		} else {
			comp := oldArchetype.GetComponent(id.Index(), typ)
			components = append(components, comp)
		}
	}

	newIndex := newArchetype.Spawn(components)
	newId := NewEntityId(newArchetypeId, newIndex)

	if hasRef {
		oldArchetype.refs.Del(id)
		ref.current = newId
		newArchetype.refs.Put(newId, ref)
	}

	oldArchetype.Delete(id.Index())
	return newId
}

func (s *Storage) RemoveComponent(id EntityId, compType reflect.Type) EntityId {
	oldArchetype := s.archetypes[id.ArchetypeId()]

	newTypes := make([]reflect.Type, 0, len(oldArchetype.types)-1)
	for _, typ := range oldArchetype.types {
		if typ != compType {
			newTypes = append(newTypes, typ)
		}
	}

	ref, hasRef := oldArchetype.refs.Get(id)

	if len(newTypes) == 0 {
		if hasRef {
			oldArchetype.refs.Del(id)
			ref.deleted = true
			s.refs.Del(ref.ref)
		}
		oldArchetype.Delete(id.Index())
		return 0
	}

	newArchetypeId := hashTypesToUint32(newTypes)
	newArchetype, exists := s.archetypes[newArchetypeId]
	if !exists {
		newArchetype = NewArchetype(newArchetypeId, newTypes)
		s.archetypes[newArchetypeId] = newArchetype
	}

	components := make([]any, 0, len(newTypes))
	for _, typ := range newTypes {
		comp := oldArchetype.GetComponent(id.Index(), typ)
		components = append(components, comp)
	}

	newIndex := newArchetype.Spawn(components)
	newId := NewEntityId(newArchetypeId, newIndex)

	if hasRef {
		oldArchetype.refs.Del(id)
		ref.current = newId
		newArchetype.refs.Put(newId, ref)
	}

	oldArchetype.Delete(id.Index())
	return newId
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

type ComponentReader interface {
	GetComponent(EntityId, reflect.Type) any
}

func ReadComponent[T any](reader ComponentReader, entityId EntityId) *T {
	return reader.GetComponent(entityId, reflect.TypeFor[T]()).(*T)
}
