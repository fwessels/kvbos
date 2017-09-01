package kvbos

import (
	"fmt"
)

type Value struct {
	value []byte
}

type KVBos struct {
}

const (
	ValueBlockSize = 128
	KeyBlockSize = 128
	KeyAlign = 8
)

type ValueBlock [ValueBlockSize]byte
type KeyBlock [KeyBlockSize]byte

var ValueBlocks [10]ValueBlock
var ValuePointer = uint64(0x0000000000000000)

var KeyBlocks [10]KeyBlock
var KeyPointer = uint64(0x7f)

func (kvb *KVBos) Put(key []byte, value []byte) {

	// Copy value into block
	valuePointer := ValuePointer
	valueSize := uint32(len(value))
	keySize := uint32(len(key))
	copy(ValueBlocks[0][valuePointer:], value[:])

	ValuePointer += uint64(valueSize)

	kh := newKeyHeader(make([]byte, KeyHeaderSize), KeyHeaderSize)
	kh.SetValuePointer(valuePointer)
	kh.SetValueSize(valueSize)
	kh.SetKeySize(keySize)

	// Copy key header first (has a deterministic size,
	// so we can iterate manually 'downwards' in memory if need be)
	copy(KeyBlocks[0][KeyPointer-uint64(KeyHeaderSize-1):], kh[:])

	// Copy key itself
	keyAlignedSize := kh.KeyAlignedSize()
	copy(KeyBlocks[0][KeyPointer-uint64(KeyHeaderSize+keyAlignedSize-1):], key[:])

	KeyPointer -= KeyHeaderSize + uint64(keyAlignedSize)

	kbh := newKeyBlockHeader(KeyBlocks[0][:])
	kbh.AddSortedPointer(KeyPointer + 1 + keyAlignedSize)
	fmt.Println(KeyBlocks[0])
}

func (kvb *KVBos) Get(key []byte) []byte {

	kbh := newKeyBlockHeader(KeyBlocks[0][:])
	return kbh.Get(key)
}
