package ecs

import (
	"reflect"
	"unsafe"
)

// iface represents the internal memory layout of an interface{}.
type iface struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

const (
	blockSize = 64
)

// componentBlock represents a block of components with a bitmap for tracking filled slots
type componentBlock struct {
	data   []byte
	filled uint64
}

// ComponentStorage manages blocks of memory for a single component type
type ComponentStorage struct {
	blocks []componentBlock
	stride uintptr
	typ    reflect.Type
}

// NewComponentStorage creates a new component storage for the given type
func NewComponentStorage(t reflect.Type, initialCapacity int) *ComponentStorage {
	stride := t.Size()

	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Map || t.Kind() == reflect.Chan || t.Kind() == reflect.Func {
		panic("ComponentStorage only supports value types (structs/primitives), not pointers/maps/channels/funcs")
	}

	numBlocks := (initialCapacity + blockSize - 1) / blockSize
	if numBlocks == 0 {
		numBlocks = 1
	}

	blocks := make([]componentBlock, numBlocks)
	for i := range blocks {
		blocks[i].data = make([]byte, stride*blockSize)
		blocks[i].filled = 0
	}

	return &ComponentStorage{
		blocks: blocks,
		stride: stride,
		typ:    t,
	}
}

// Append adds a component to storage and returns its index
// Finds the first empty slot or allocates a new one
func (cs *ComponentStorage) Append(item any) int {
	for blockIdx := range cs.blocks {
		block := &cs.blocks[blockIdx]

		if block.filled != ^uint64(0) {
			for slotIdx := range blockSize {
				mask := uint64(1) << slotIdx
				if block.filled&mask == 0 {
					block.filled |= mask
					globalIndex := blockIdx*blockSize + slotIdx
					cs.writeComponent(blockIdx, slotIdx, item)
					return globalIndex
				}
			}
		}
	}

	// No empty slots found, need to allocate a new block
	newBlock := componentBlock{
		data:   make([]byte, cs.stride*blockSize),
		filled: 1,
	}
	cs.blocks = append(cs.blocks, newBlock)
	blockIdx := len(cs.blocks) - 1
	cs.writeComponent(blockIdx, 0, item)
	return blockIdx * blockSize
}

// writeComponent writes a component to a specific block and slot
func (cs *ComponentStorage) writeComponent(blockIdx, slotIdx int, item any) {
	ifacePtr := (*iface)(unsafe.Pointer(&item))
	offset := uintptr(slotIdx) * cs.stride
	targetPtr := unsafe.Pointer(&cs.blocks[blockIdx].data[offset])
	srcBytes := unsafe.Slice((*byte)(ifacePtr.data), cs.stride)
	dstBytes := unsafe.Slice((*byte)(targetPtr), cs.stride)
	copy(dstBytes, srcBytes)
}

// Get returns a pointer to the component at the given index
func (cs *ComponentStorage) Get(index int) any {
	if index < 0 {
		return nil
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.blocks) {
		return nil
	}

	block := &cs.blocks[blockIdx]
	mask := uint64(1) << slotIdx

	if block.filled&mask == 0 {
		return nil
	}

	offset := uintptr(slotIdx) * cs.stride
	elementPtr := unsafe.Pointer(&block.data[offset])
	return reflect.NewAt(cs.typ, elementPtr).Interface()
}

// Delete marks a component slot as empty
func (cs *ComponentStorage) Delete(index int) {
	if index < 0 {
		return
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.blocks) {
		return
	}

	block := &cs.blocks[blockIdx]
	mask := uint64(1) << slotIdx

	block.filled &^= mask
}

// Has checks if a component exists at the given index
func (cs *ComponentStorage) Has(index int) bool {
	if index < 0 {
		return false
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.blocks) {
		return false
	}

	block := &cs.blocks[blockIdx]
	mask := uint64(1) << slotIdx

	return block.filled&mask != 0
}

// Compact reorganizes component storage to remove empty slots and reduce fragmentation
// Returns a map of old index -> new index for updating entity references
func (cs *ComponentStorage) Compact() map[int]int {
	indexMap := make(map[int]int)
	writePos := 0

	// Temporary buffer for the new compacted data
	tempBlocks := make([]componentBlock, 0, len(cs.blocks))
	currentBlock := componentBlock{
		data:   make([]byte, cs.stride*blockSize),
		filled: 0,
	}

	// Iterate through all blocks and slots, copying filled slots to the beginning
	for blockIdx := range cs.blocks {
		block := &cs.blocks[blockIdx]

		for slotIdx := range blockSize {
			mask := uint64(1) << slotIdx
			if block.filled&mask != 0 {
				oldIndex := blockIdx*blockSize + slotIdx

				// Calculate new position
				newSlotIdx := writePos % blockSize

				indexMap[oldIndex] = writePos

				// Allocate new block if needed
				if newSlotIdx == 0 && writePos > 0 {
					tempBlocks = append(tempBlocks, currentBlock)
					currentBlock = componentBlock{
						data:   make([]byte, cs.stride*blockSize),
						filled: 0,
					}
				}

				// Copy component data
				srcOffset := uintptr(slotIdx) * cs.stride
				dstOffset := uintptr(newSlotIdx) * cs.stride
				srcBytes := block.data[srcOffset : srcOffset+cs.stride]
				dstBytes := currentBlock.data[dstOffset : dstOffset+cs.stride]
				copy(dstBytes, srcBytes)

				// Mark slot as filled
				currentBlock.filled |= uint64(1) << newSlotIdx

				writePos++
			}
		}
	}

	// Append the last block if it has any data
	if currentBlock.filled != 0 {
		tempBlocks = append(tempBlocks, currentBlock)
	}

	// Replace old blocks with compacted blocks
	// Keep at least one block even if empty
	if len(tempBlocks) == 0 {
		tempBlocks = []componentBlock{{
			data:   make([]byte, cs.stride*blockSize),
			filled: 0,
		}}
	}

	cs.blocks = tempBlocks

	return indexMap
}
