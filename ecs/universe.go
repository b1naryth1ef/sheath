package ecs

import "iter"

// Universe manages everything required for an ECS-based simulation.
type Universe struct {
	storage EntityStorage
}

// NewUniverse creates a new universe with the basic built-in entity storage
func NewUniverse() *Universe {
	return &Universe{
		storage: NewSimpleEntityStorage(),
	}
}

// Spawn creates a new entity with the provided components
func (u *Universe) Spawn(components ...any) EntityId {
	return u.storage.Create(components...)
}

// Get returns `EntityData` for the given entity id, or nil if non-existant
func (u *Universe) Get(id EntityId) EntityData {
	return u.storage.Get(id)
}

// Delete removes a entity by id
func (u *Universe) Delete(id EntityId) {
	u.storage.Delete(id)
}

// Filter returns a filtered iterator over all entity data in the storage
func (u *Universe) Filter(filter EntityFilter) iter.Seq[EntityData] {
	return u.storage.Filter(filter)
}
