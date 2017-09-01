package kvbos

import (
	"encoding/binary"
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

//type Key struct {
//	valuePointer uint64
//	valueSize	 uint32
//	keySize 	 uint32
//	key		     []byte
//}

// Try to read it again
//
const KeyHeaderSize = 8 + 4 + 4

type KeyHeader []byte

func newKeyHeader(b []byte, size uint32) KeyHeader { return KeyHeader(b[:size]) }

func (k KeyHeader) ValuePointer() uint64 { return uint64(binary.LittleEndian.Uint64(k[0:])) }

func (k KeyHeader) ValueSize() uint32 { return uint32(binary.LittleEndian.Uint32(k[8:])) }

func (k KeyHeader) KeySize() uint32 { return uint32(binary.LittleEndian.Uint32(k[12:])) }

func (k KeyHeader) SetValuePointer(vp uint64) { binary.LittleEndian.PutUint64(k[0:], vp) }

func (k KeyHeader) SetValueSize(vs uint32) { binary.LittleEndian.PutUint32(k[8:], vs) }

func (k KeyHeader) SetKeySize(ks uint32) { binary.LittleEndian.PutUint32(k[12:], ks) }
