package ecs

import (
	"reflect"
)

type Universe struct {
	idInc    EntityId
	entities map[EntityId]*EntityData
}

func NewUniverse() *Universe {
	return &Universe{
		entities: make(map[EntityId]*EntityData),
	}
}

func (u *Universe) Get(id EntityId) *EntityData {
	return u.entities[id]
}

func (u *Universe) Spawn(components ...any) EntityId {
	types := make([]reflect.Type, 0, len(components))
	for _, comp := range components {
		types = append(types, reflect.TypeOf(comp))
	}
	data := &EntityData{Types: types, Components: components}
	u.idInc += 1
	u.entities[u.idInc] = data
	return u.idInc
}
