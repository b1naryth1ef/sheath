package ecs

import (
	"log"
	"reflect"
	"strings"
)

type EntityId = uint64

var entityIdType = reflect.TypeOf(EntityId(0))

type EntityData struct {
	Id         EntityId
	Types      []reflect.Type
	Components []any
}

func (e *EntityData) Has(components ...any) bool {
	for _, comp := range components {
		compType := reflect.TypeOf(comp)
		matched := false
		for _, ourCompType := range e.Types {
			if compType == ourCompType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (e *EntityData) Exec(target any) bool {
	// typeOf(target) == T*
	targetType := reflect.TypeOf(target).Elem()
	targetValue := reflect.ValueOf(target).Elem()

	for i := 0; i < targetType.NumField(); i++ {
		targetField := targetValue.Field(i)
		targetTypeField := targetType.Field(i)

		if targetTypeField.Type == entityIdType {
			targetField.SetUint(e.Id)
			continue
		}

		matched := false

		tag := targetTypeField.Tag.Get("ecs")
		if tag != "" {
			tagParts := strings.Split(targetTypeField.Tag.Get("ecs"), ",")
			for _, part := range tagParts {
				if part == "optional" {
					matched = true
				} else {
					log.Panicf("invalid ecs struct tag: %v", part)
				}
			}
		}

		for fieldIdx, fieldType := range e.Types {
			if targetTypeField.Type == fieldType {
				targetField.Set(reflect.ValueOf(e.Components[fieldIdx]))
				matched = true
				break
			}
		}

		if !matched {
			return false
		}
	}

	return true
}
