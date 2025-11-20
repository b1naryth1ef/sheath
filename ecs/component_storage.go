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
	blockSize = 256
)

// componentBlock represents a block of components with tracking for filled slots
type componentBlock struct {
	data   []byte
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
		}
	}

	numBlocks := (initialCapacity + blockSize - 1) / blockSize
	if numBlocks == 0 {
		numBlocks = 1
	}

	blocks := make([]componentBlock, numBlocks)
	for i := range blocks {
		blocks[i].data = make([]byte, stride*blockSize)
	}

	return &ComponentStorage{
		blocks:           blocks,
		stride:           stride,
		typ:              t,
		containsPointers: false,
	}
}

// Append adds a component to storage and returns its index
// Finds the first empty slot or allocates a new one
func (cs *ComponentStorage) Append(item any) int {
	if cs.containsPointers {
		return cs.appendGC(item)
	}
	return cs.appendBlock(item)
}

func (cs *ComponentStorage) appendGC(item any) int {
	for blockIdx := range cs.gcBlocks {
		block := &cs.gcBlocks[blockIdx]

		for slotIdx := range blockSize {
			if !block.filled[slotIdx] {
				block.data[slotIdx] = item
				block.filled[slotIdx] = true
				return blockIdx*blockSize + slotIdx
			}
		}
	}

	// No empty slots found, allocate a new block
	var newBlock gcBlock
	newBlock.data[0] = item
	newBlock.filled[0] = true
	cs.gcBlocks = append(cs.gcBlocks, newBlock)
	return (len(cs.gcBlocks) - 1) * blockSize
}

func (cs *ComponentStorage) appendBlock(item any) int {
	for blockIdx := range cs.blocks {
		block := &cs.blocks[blockIdx]

		for slotIdx := range blockSize {
			if !block.filled[slotIdx] {
				block.filled[slotIdx] = true
				globalIndex := blockIdx*blockSize + slotIdx
				cs.writeComponent(blockIdx, slotIdx, item)
				return globalIndex
			}
		}
	}

	// No empty slots found, need to allocate a new block
	var newBlock componentBlock
	newBlock.data = make([]byte, cs.stride*blockSize)
	newBlock.filled[0] = true
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
	elementPtr := unsafe.Pointer(&block.data[offset])
	return reflect.NewAt(cs.typ, elementPtr).Interface()
}

// Delete marks a component slot as empty
func (cs *ComponentStorage) Delete(index int) {
	if cs.containsPointers {
		cs.deleteGC(index)
		return
	}
	cs.deleteBlock(index)
}

func (cs *ComponentStorage) deleteGC(index int) {
	if index < 0 {
		return
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.gcBlocks) {
		return
	}

	block := &cs.gcBlocks[blockIdx]
	block.filled[slotIdx] = false
	block.data[slotIdx] = nil
}

func (cs *ComponentStorage) deleteBlock(index int) {
	if index < 0 {
		return
	}

	blockIdx := index / blockSize
	slotIdx := index % blockSize

	if blockIdx >= len(cs.blocks) {
		return
	}

	block := &cs.blocks[blockIdx]
	block.filled[slotIdx] = false
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

	return indexMap
}

func (cs *ComponentStorage) compactBlock() map[int]int {
	indexMap := make(map[int]int)
	writePos := 0

	// Temporary buffer for the new compacted data
	tempBlocks := make([]componentBlock, 0, len(cs.blocks))
	var currentBlock componentBlock
	currentBlock.data = make([]byte, cs.stride*blockSize)

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
					currentBlock.data = make([]byte, cs.stride*blockSize)
				}

				// Copy component data
				srcOffset := uintptr(slotIdx) * cs.stride
				dstOffset := uintptr(newSlotIdx) * cs.stride
				srcBytes := block.data[srcOffset : srcOffset+cs.stride]
				dstBytes := currentBlock.data[dstOffset : dstOffset+cs.stride]
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
		emptyBlock.data = make([]byte, cs.stride*blockSize)
		tempBlocks = []componentBlock{emptyBlock}
	}

	cs.blocks = tempBlocks

	return indexMap
}
