package ecs_test

import (
	"testing"

	"github.com/b1naryth1ef/sheath/ecs"
	"github.com/stretchr/testify/assert"
)

type V1 struct{ A uint8 }
type V2 struct{ B uint16 }
type V3 struct{ C uint32 }
type V4 struct{ D uint64 }

func TestSimpleEntityStorage(t *testing.T) {
	storage := ecs.NewSimpleEntityStorage()

	data := storage.Get(1)
	assert.Nil(t, data)

	id := storage.Create(&V1{1}, &V2{2}, &V3{3})
	assert.True(t, id > 0)

	data = storage.Get(id)
	assert.NotNil(t, data)

	var v1 *V1
	assert.True(t, data.Read(&v1))
	assert.Equal(t, uint8(1), v1.A)

	var v4 *V4
	assert.False(t, data.Read(&v4))
	assert.Nil(t, v4)

	var fullData struct {
		*V1
		*V2
		*V3
	}
	assert.True(t, data.Fill(&fullData))
	assert.Equal(t, uint8(1), fullData.V1.A)
	assert.Equal(t, uint16(2), fullData.V2.B)
	assert.Equal(t, uint32(3), fullData.V3.C)

	count := 0
	for range storage.Filter(ecs.EntityFilter{}.WithComponents(&V1{})) {
		count += 1
	}
	assert.Equal(t, 1, count)

}

func TestSimpleEntityStorageIteration(t *testing.T) {
	storage := createSimpleEntityStorage(100_000)
	count := 0
	for range storage.Filter(ecs.EntityFilter{}.WithComponents(&V1{})) {
		count += 1
	}
	assert.Equal(t, 100_000, count)

	count = 0
	for range storage.Filter(ecs.EntityFilter{}.WithComponents(&V2{})) {
		count += 1
	}
	assert.Equal(t, 50_001, count)

	count = 0
	for range storage.Filter(ecs.EntityFilter{}.WithComponents(&V4{})) {
		count += 1
	}
	assert.Equal(t, 1, count)
}

func createSimpleEntityStorage(size int) *ecs.SimpleEntityStorage {
	storage := ecs.NewSimpleEntityStorage()

	for i := range size - 1 {
		if i%2 == 0 {
			storage.Create(&V1{1}, &V2{2}, &V3{uint32(i)})
		} else {
			storage.Create(&V1{1}, &V3{uint32(i)})
		}
	}

	storage.Create(&V1{1}, &V2{2}, &V3{3}, &V4{4})

	return storage
}
