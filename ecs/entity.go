package ecs

import (
	"reflect"
)

// EntityId wraps a uint64 to be used as the id for entities
type EntityId = uint64

var entityIdType = reflect.TypeOf(EntityId(0))
