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

	valuePointer := ValuePointer
	valueSize := uint32(len(value))
	keySize := uint32(len(key))
	copy(ValueBlocks[0][valuePointer:], value[:])
	//fmt.Println(ValueBlocks[0])

	ValuePointer += uint64(valueSize)

	kh := newKeyHeader(make([]byte, KeyHeaderSize), KeyHeaderSize) //Key{valuePointer: valuePointer, valueSize: valueSize, keySize: keySize}
	kh.SetValuePointer(valuePointer)
	kh.SetValueSize(valueSize)
	kh.SetKeySize(keySize)

	fmt.Println(kh)
	copy(KeyBlocks[0][KeyPointer-uint64(KeyHeaderSize-1):], kh[:])

	keyAlignedSize := (keySize + KeyAlign-1) & ^uint32(KeyAlign-1)

	copy(KeyBlocks[0][KeyPointer-uint64(KeyHeaderSize+keyAlignedSize-1):], key[:])

	KeyPointer -= KeyHeaderSize + uint64(keyAlignedSize)
}

func (kvb *KVBos) Get(key []byte) []byte {

	keyPointer := uint64(0x7f)

	for i := 0; i < 3; i++ {
		kh := newKeyHeader(KeyBlocks[0][keyPointer-uint64(KeyHeaderSize-1):], KeyHeaderSize)
		fmt.Println("kh", kh, kh.ValuePointer(), kh.KeySize())

		keyAlignedSize := (kh.KeySize() + KeyAlign-1) & ^uint32(KeyAlign-1)

		k := make([]byte, kh.KeySize())
		copy(k, KeyBlocks[0][keyPointer-uint64(KeyHeaderSize+keyAlignedSize-1):keyPointer-uint64(KeyHeaderSize+keyAlignedSize-1-kh.KeySize())])

		if bytes.Compare(k, key) == 0 {
			fmt.Println("k-e-y   f-o-u-n-d =", string(k))

			v := make([]byte, kh.ValueSize())
			copy(v, ValueBlocks[0][kh.ValuePointer():kh.ValuePointer()+uint64(kh.ValueSize())])
			return v
		}

		keyPointer -= KeyHeaderSize + uint64(keyAlignedSize)
	}

	return []byte{}
}
