package kvbos

import (
	"encoding/binary"
	_ "bytes"
)

// TODO: Add size of block
//

// TODO: Add version number??

//type KeyBlock struct {
//	BlockSize      uint64
//  Entries        uint32 /* number of keys & pointers in this block */
//  prevBlock      uint64
//  nextBlock      uint64
//	SortedPointers []uint64
//
//	Keys           []Key /* stored from the top down */
//}

//type Key(Header) struct {
//	valuePointer uint64
//	valueSize	 uint32
//	keySize 	 uint32
//	key		     []byte
//}

// TODO: Try to read it again when mismatch detected while reading merged blocks
//
const KeyHeaderSize = 8 + 4 + 4

type KeyHeader []byte

func newKeyHeader(b []byte, size uint32) KeyHeader { return KeyHeader(b[:size]) }

func (kh KeyHeader) ValuePointer() uint64 { return uint64(binary.LittleEndian.Uint64(kh[0:])) }

func (kh KeyHeader) ValueSize() uint32 { return uint32(binary.LittleEndian.Uint32(kh[8:])) }

func (kh KeyHeader) KeySize() uint32 { return uint32(binary.LittleEndian.Uint32(kh[12:])) }

func (kh KeyHeader) KeyAlignedSize() uint64 { return uint64((kh.KeySize() + KeyAlign-1) & ^uint32(KeyAlign-1)) }

func (kh KeyHeader) SetValuePointer(vp uint64) { binary.LittleEndian.PutUint64(kh[0:], vp) }

func (kh KeyHeader) SetValueSize(vs uint32) { binary.LittleEndian.PutUint32(kh[8:], vs) }

func (kh KeyHeader) SetKeySize(ks uint32) { binary.LittleEndian.PutUint32(kh[12:], ks) }

//type KeyBlockHeader struct {
//	entries uint32
//	unused uint32
//  sortedKeyPointers []uint64
//}

type KeyBlockHeader []byte

func newKeyBlockHeader(b []byte) KeyBlockHeader { return KeyBlockHeader(b) }

func (kbh KeyBlockHeader) Entries() uint32 { return uint32(binary.LittleEndian.Uint32(kbh[0:])) }

func (kbh KeyBlockHeader) AddSortedPointer(keyPointer uint64) {

	entries := kbh.Entries()
	if entries == 0 {
		binary.LittleEndian.PutUint64(kbh[8:], keyPointer)
	} else {
		//cmp = Compare(binary.LittleEndian.Uint64(kbh[8:], keyPointer)
	}

	binary.LittleEndian.PutUint32(kbh[0:], entries + 1)
}

//func Compare(a, b uint64) int {
//	akh := newKeyHeader(KeyBlocks[0][a:], KeyHeaderSize)
//	bkh := newKeyHeader(KeyBlocks[0][b:], KeyHeaderSize)
//	return bytes.Compare()
//}