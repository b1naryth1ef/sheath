package ecs

import (
	"reflect"
	"unsafe"
)

const (
	blockSize = 256
)

// componentBlock represents a block of components with tracking for filled slots
type componentBlock struct {
	data   unsafe.Pointer
	filled [blockSize]bool
}

// gcBlock represents a block of GC-scannable components
type gcBlock struct {
	data   [blockSize]any
	filled [blockSize]bool
}

// ComponentStorage manages blocks of memory for a single component type
type ComponentStorage struct {
	blocks []componentBlock
	stride uintptr
	typ    reflect.Type

	containsPointers bool
	gcBlocks         []gcBlock

	// freeSlots stores the indices of deleted components that can be reused
	freeSlots []int
	// nextIndex is the next available index for new components when freeSlots is empty
	nextIndex int
}

// typeContainsPointers recursively checks if a type contains pointers, maps, slices, or channels
func typeContainsPointers(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice:
		return true
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if typeContainsPointers(t.Field(i).Type) {
				return true
			}
		}
	case reflect.Array:
		return typeContainsPointers(t.Elem())
	}
	return false
}

// NewComponentStorage creates a new component storage for the given type
func NewComponentStorage(t reflect.Type, initialCapacity int) *ComponentStorage {
	stride := t.Size()

	containsPointers := typeContainsPointers(t)

	if containsPointers {
		numBlocks := (initialCapacity + blockSize - 1) / blockSize
		if numBlocks == 0 {
			numBlocks = 1
		}

		gcBlocks := make([]gcBlock, numBlocks)

		return &ComponentStorage{
			typ:              t,
			stride:           stride,
			containsPointers: true,
			gcBlocks:         gcBlocks,
			nextIndex:        0,
		}
	}

	numBlocks := (initialCapacity + blockSize - 1) / blockSize
	if numBlocks == 0 {
		numBlocks = 1
	}

	blocks := make([]componentBlock, numBlocks)
	blockBytes := stride * blockSize
	for i := range blocks {
		// Allocate raw memory for the block
		blocks[i].data = unsafe.Pointer(&make([]byte, blockBytes)[0])
	}

	return &ComponentStorage{
		blocks:           blocks,
		stride:           stride,
		typ:              t,
		containsPointers: false,
		nextIndex:        0,
	}
}

// Append adds a component to storage and returns its index
// It reuses slots from deleted components or appends to the end.
func (cs *ComponentStorage) Append(item any) int {
	if cs.containsPointers {
		return cs.appendGC(item)
	}
	return cs.appendBlock(item)
}

func (cs *ComponentStorage) appendGC(item any) int {
	// Try to reuse a slot from the free list
	if len(cs.freeSlots) > 0 {
		index := cs.freeSlots[len(cs.freeSlots)-1]
		cs.freeSlots = cs.freeSlots[:len(cs.freeSlots)-1]

		blockIdx := index / blockSize
		slotIdx := index % blockSize

		block := &cs.gcBlocks[blockIdx]
		block.data[slotIdx] = item
		block.filled[slotIdx] = true
		return index
	}

	// No free slots, append to the end
	index := cs.nextIndex
	cs.nextIndex++

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	// Allocate new block if necessary
	if blockIdx >= len(cs.gcBlocks) {
		var newBlock gcBlock
		cs.gcBlocks = append(cs.gcBlocks, newBlock)
	}

	block := &cs.gcBlocks[blockIdx]
	block.data[slotIdx] = item
	block.filled[slotIdx] = true
	return index
}

func (cs *ComponentStorage) appendBlock(item any) int {
	// Try to reuse a slot from the free list
	if len(cs.freeSlots) > 0 {
		index := cs.freeSlots[len(cs.freeSlots)-1]
		cs.freeSlots = cs.freeSlots[:len(cs.freeSlots)-1]

		blockIdx := index / blockSize
		slotIdx := index % blockSize

		block := &cs.blocks[blockIdx]
		block.filled[slotIdx] = true
		cs.writeComponent(blockIdx, slotIdx, item)
		return index
	}

	// No free slots, append to the end
	index := cs.nextIndex
	cs.nextIndex++

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	// Allocate new block if necessary
	if blockIdx >= len(cs.blocks) {
		var newBlock componentBlock
		blockBytes := cs.stride * blockSize
		newBlock.data = unsafe.Pointer(&make([]byte, blockBytes)[0])

		cs.blocks = append(cs.blocks, newBlock)
	}

	block := &cs.blocks[blockIdx]
	block.filled[slotIdx] = true
	cs.writeComponent(blockIdx, slotIdx, item)
	return index
}

// writeComponent writes a component to a specific block and slot
func (cs *ComponentStorage) writeComponent(blockIdx, slotIdx int, item any) {
	ifacePtr := (*iface)(unsafe.Pointer(&item))
	offset := uintptr(slotIdx) * cs.stride
	targetPtr := unsafe.Pointer(uintptr(cs.blocks[blockIdx].data) + offset)

	// Use efficient bulk copy via unsafe slices
	srcBytes := unsafe.Slice((*byte)(ifacePtr.data), cs.stride)
	dstBytes := unsafe.Slice((*byte)(targetPtr), cs.stride)
	copy(dstBytes, srcBytes)
}

// Get returns a pointer to the component at the given index
func (cs *ComponentStorage) Get(index int) any {
	if cs.containsPointers {
		return cs.getGC(index)
	}
	return cs.getBlock(index)
}

func (cs *ComponentStorage) getGC(index int) any {
	if index < 0 {
		return nil
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.gcBlocks) {
		return nil
	}

	block := &cs.gcBlocks[blockIdx]

	if !block.filled[slotIdx] {
		return nil
	}

	return block.data[slotIdx]
}

func (cs *ComponentStorage) getBlock(index int) any {
	if index < 0 {
		return nil
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.blocks) {
		return nil
	}

	block := &cs.blocks[blockIdx]

	if !block.filled[slotIdx] {
		return nil
	}

	offset := uintptr(slotIdx) * cs.stride
	elementPtr := unsafe.Pointer(uintptr(block.data) + offset)
	return reflect.NewAt(cs.typ, elementPtr).Interface()
}

// Delete marks a component slot as empty
func (cs *ComponentStorage) Delete(index int) {
	if index < 0 {
		return
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if cs.containsPointers {
		if blockIdx >= len(cs.gcBlocks) {
			return
		}
		block := &cs.gcBlocks[blockIdx]
		if block.filled[slotIdx] {
			block.filled[slotIdx] = false
			block.data[slotIdx] = nil
			cs.freeSlots = append(cs.freeSlots, index)
		}
	} else {
		if blockIdx >= len(cs.blocks) {
			return
		}
		block := &cs.blocks[blockIdx]
		if block.filled[slotIdx] {
			block.filled[slotIdx] = false
			cs.freeSlots = append(cs.freeSlots, index)
		}
	}
}

// Has checks if a component exists at the given index
func (cs *ComponentStorage) Has(index int) bool {
	if cs.containsPointers {
		return cs.hasGC(index)
	}
	return cs.hasBlock(index)
}

func (cs *ComponentStorage) hasGC(index int) bool {
	if index < 0 {
		return false
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.gcBlocks) {
		return false
	}

	block := &cs.gcBlocks[blockIdx]
	return block.filled[slotIdx]
}

func (cs *ComponentStorage) hasBlock(index int) bool {
	if index < 0 {
		return false
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.blocks) {
		return false
	}

	block := &cs.blocks[blockIdx]
	return block.filled[slotIdx]
}

// Compact reorganizes component storage to remove empty slots and reduce fragmentation
// Returns a map of old index -> new index for updating entity references
func (cs *ComponentStorage) Compact() map[int]int {
	if cs.containsPointers {
		return cs.compactGC()
	}
	return cs.compactBlock()
}

func (cs *ComponentStorage) compactGC() map[int]int {
	indexMap := make(map[int]int)
	writePos := 0

	tempBlocks := make([]gcBlock, 0, len(cs.gcBlocks))
	var currentBlock gcBlock

	for blockIdx := range cs.gcBlocks {
		block := &cs.gcBlocks[blockIdx]

		for slotIdx := range blockSize {
			if block.filled[slotIdx] {
				oldIndex := blockIdx*blockSize + slotIdx

				newSlotIdx := writePos % blockSize

				indexMap[oldIndex] = writePos

				if newSlotIdx == 0 && writePos > 0 {
					tempBlocks = append(tempBlocks, currentBlock)
					currentBlock = gcBlock{}
				}

				currentBlock.data[newSlotIdx] = block.data[slotIdx]
				currentBlock.filled[newSlotIdx] = true

				writePos++
			}
		}
	}

	hasData := false
	for _, filled := range currentBlock.filled {
		if filled {
			hasData = true
			break
		}
	}
	if hasData {
		tempBlocks = append(tempBlocks, currentBlock)
	}

	if len(tempBlocks) == 0 {
		tempBlocks = []gcBlock{{}}
	}

	cs.gcBlocks = tempBlocks
	cs.freeSlots = nil
	cs.nextIndex = writePos

	return indexMap
}

func (cs *ComponentStorage) compactBlock() map[int]int {
	indexMap := make(map[int]int)
	writePos := 0

	// Temporary buffer for the new compacted data
	tempBlocks := make([]componentBlock, 0, len(cs.blocks))
	var currentBlock componentBlock
	blockBytes := cs.stride * blockSize
	currentBlock.data = unsafe.Pointer(&make([]byte, blockBytes)[0])

	// Iterate through all blocks and slots, copying filled slots to the beginning
	for blockIdx := range cs.blocks {
		block := &cs.blocks[blockIdx]

		for slotIdx := range blockSize {
			if block.filled[slotIdx] {
				oldIndex := blockIdx*blockSize + slotIdx

				// Calculate new position
				newSlotIdx := writePos % blockSize

				indexMap[oldIndex] = writePos

				// Allocate new block if needed
				if newSlotIdx == 0 && writePos > 0 {
					tempBlocks = append(tempBlocks, currentBlock)
					currentBlock = componentBlock{}
					currentBlock.data = unsafe.Pointer(&make([]byte, blockBytes)[0])
				}

				// Copy component data using raw pointers
				srcOffset := uintptr(slotIdx) * cs.stride
				dstOffset := uintptr(newSlotIdx) * cs.stride
				srcPtr := unsafe.Pointer(uintptr(block.data) + srcOffset)
				dstPtr := unsafe.Pointer(uintptr(currentBlock.data) + dstOffset)

				srcBytes := unsafe.Slice((*byte)(srcPtr), cs.stride)
				dstBytes := unsafe.Slice((*byte)(dstPtr), cs.stride)
				copy(dstBytes, srcBytes)

				// Mark slot as filled
				currentBlock.filled[newSlotIdx] = true

				writePos++
			}
		}
	}

	// Append the last block if it has any data
	hasData := false
	for _, filled := range currentBlock.filled {
		if filled {
			hasData = true
			break
		}
	}
	if hasData {
		tempBlocks = append(tempBlocks, currentBlock)
	}

	// Replace old blocks with compacted blocks
	// Keep at least one block even if empty
	if len(tempBlocks) == 0 {
		var emptyBlock componentBlock
		emptyBlock.data = unsafe.Pointer(&make([]byte, blockBytes)[0])
		tempBlocks = []componentBlock{emptyBlock}
	}

	cs.blocks = tempBlocks
	cs.freeSlots = nil
	cs.nextIndex = writePos

	return indexMap
}
