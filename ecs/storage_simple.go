package ecs

import (
	"iter"
	"log"
	"reflect"
)

// SimpleEntityStorage implements EntityStorage with an in-memory map of structs
// containing pointers to components. This is designed to be very simple and very
// inefficient.
type SimpleEntityStorage struct {
	idInc    EntityId
	entities map[EntityId]*SimpleEntityData
}

// NewSimpleEntityStorage creates a new instance of SimpleEntityStorage
func NewSimpleEntityStorage() *SimpleEntityStorage {
	return &SimpleEntityStorage{entities: make(map[EntityId]*SimpleEntityData)}
}

// Filter implements EntityStorage.Filter
func (s *SimpleEntityStorage) Filter(filter EntityFilter) iter.Seq[EntityData] {
	return func(yield func(EntityData) bool) {
		for _, v := range s.entities {
			if !filter.Exec(v) {
				continue
			}

			if !yield(v) {
				break
			}
		}
	}
}

// Delete implements EntityStorage.Delete
func (s *SimpleEntityStorage) Delete(id EntityId) {
	delete(s.entities, id)
}

// Get implements EntityStorage.Get
func (s *SimpleEntityStorage) Get(id EntityId) EntityData {
	// NB: this check is required because Go's type system is silly and a nil
	// value of a pointer type is not equal to the nil value of an interface type.
	value, ok := s.entities[id]
	if !ok {
		return nil
	}
	return value
}

// Create implements EntityStorage.Create
func (s *SimpleEntityStorage) Create(components ...any) EntityId {
	s.idInc += 1
	id := s.idInc
	types := make([]reflect.Type, 0, len(components))
	for _, comp := range components {
		types = append(types, reflect.TypeOf(comp))
	}
	s.entities[id] = &SimpleEntityData{
		id:    id,
		types: types,
		data:  components,
	}
	return id
}

// SimpleEntityData implements EntityData for SimpleEntityStorage. All component
// data is stored within an array of pointers inside this struct.
type SimpleEntityData struct {
	id    EntityId
	types []reflect.Type
	data  []any
}

func (s *SimpleEntityData) index(typ reflect.Type) int {
	for idx, compType := range s.types {
		if compType == typ {
			return idx
		}
	}
	return -1
}

// Id implements EntityData.Id
func (s *SimpleEntityData) Id() EntityId {
	return s.id
}

// Read implements EntityData.Read
func (s *SimpleEntityData) Read(component any) bool {
	targetValue := reflect.ValueOf(component).Elem()
	targetType := targetValue.Type()

	index := s.index(targetType)
	if index == -1 {
		return false
	}
	targetValue.Set(reflect.ValueOf(s.data[index]))
	return true
}

// HasComponents implements EntityData.HasComponent
func (s *SimpleEntityData) HasComponent(types ...reflect.Type) bool {
	for _, compType := range types {
		index := s.index(compType)
		if index == -1 {
			return false
		}
	}
	return true
}

func (s *SimpleEntityData) GetComponent(componentType reflect.Type) any {
	index := s.index(componentType)
	if index == -1 {
		return nil
	}

	return s.data[index]
}

func (s *SimpleEntityData) AddComponent(component any) {
	s.types = append(s.types, reflect.TypeOf(component))
	s.data = append(s.data, component)
}

func (s *SimpleEntityData) RemoveComponent(componentType reflect.Type) bool {
	index := s.index(componentType)
	if index == -1 {
		return false
	}

	s.types[index] = s.types[len(s.types)-1]
	s.types = s.types[:len(s.types)-1]
	s.data[index] = s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	return true
}

// Fill implements EntityData.Fill
func (s *SimpleEntityData) Fill(target any) bool {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Pointer {
		log.Panicf("EntityData.Fill() must take a pointer to a struct")
	}

	targetValue := reflect.ValueOf(target).Elem()
	targetType = targetType.Elem()

	complete := true
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)

		index := s.index(field.Type)
		if index == -1 {
			complete = false
		} else {
			targetValue.Field(i).Set(reflect.ValueOf(s.data[index]))
		}
	}
	return complete
}
