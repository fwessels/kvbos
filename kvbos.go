package kvbos

import (
	"fmt"
	"bytes"
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
	fmt.Println(kbh.Entries())
	kbh.AddSortedPointer(KeyPointer + 1 + keyAlignedSize)
}

func (kvb *KVBos) Get(key []byte) []byte {

	keyPointer := uint64(0x7f)

	for i := 0; i < 3; i++ {
		pKeyHdr := keyPointer-(KeyHeaderSize-1)
		kh := newKeyHeader(KeyBlocks[0][pKeyHdr:], KeyHeaderSize)

		pKeyData := pKeyHdr-kh.KeyAlignedSize()
		if bytes.Compare(/*k*/KeyBlocks[0][pKeyData:pKeyData+uint64(kh.KeySize())], key) == 0 {
			fmt.Println("kh-e-y   f-o-u-n-d =", string(KeyBlocks[0][pKeyData:pKeyData+uint64(kh.KeySize())]))

			v := make([]byte, kh.ValueSize())
			copy(v, ValueBlocks[0][kh.ValuePointer():kh.ValuePointer()+uint64(kh.ValueSize())])
			return v
		}

		keyPointer -= KeyHeaderSize + uint64(kh.KeyAlignedSize())
	}

	return []byte{}
}
