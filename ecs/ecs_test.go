package ecs_test

import (
	"log"
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
	"github.com/stretchr/testify/assert"
)

type A struct{ A int }
type B struct{ B string }
type C struct{ C bool }
type D struct {
	X int
	Y int
}

func TestBlah(t *testing.T) {
	universe := ecs.NewUniverse()
	entityId := universe.Spawn(&A{A: 1}, &B{B: "1"}, &C{C: true})
	log.Printf("Spawned %v", entityId)

	entity := universe.Get(entityId)

	assert.True(t, entity.Has(&A{}))
	assert.False(t, entity.Has(&D{}))

	results := ecs.Exec[struct {
		A *A
		B *B
		C *C
		D *D `ecs:"optional"`
	}](universe)
	assert.Equal(t, 1, results.Len())

	for result := range results.Iter() {
		log.Printf("[%v] %v / %v / %v", entityId, result.A, result.B, result.C)
	}
}

func BenchmarkUniverse(b *testing.B) {
	universe := ecs.NewUniverse()

	for range 10000 {
		universe.Spawn(&A{A: 1}, &B{B: "1"}, &C{C: true})
	}

	for range b.N {
		results := ecs.Exec[struct {
			Id ecs.EntityId
			A  *A
			B  *B
			C  *C
			D  *D `ecs:"optional"`
		}](universe)

		count := 0
		for range results.Iter() {
			count += 1
		}
	}
}
