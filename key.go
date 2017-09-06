package kvbos

import (
	"bytes"
	"encoding/binary"
	_ "fmt"
)

// TODO: Add size of block
//

// TODO: Add version number??

// TODO: Consider writing keys in increasing order (allow for true append only write log if we leave out the sorted list)

//type KeyBlock struct {
//  Entries        uint64 /* number of keys & pointers in this block */
//	SortedPointers []uint64
//
//	Keys           []Key /* stored from the top down */
//}

//type Key(Header) struct {
//	key		     []byte //
//	valuePointer uint64
//	valueSize	 uint32
//	keySize 	 uint16
//	crc16 	     uint16
//}

// TODO: Try to read it again when mismatch detected while reading merged blocks
//
const KeyHeaderSize = 8 + 4 + 4

type KeyHeader []byte

func newKeyHeader(b []byte, size uint32) KeyHeader { return KeyHeader(b[:size]) }

func (kh KeyHeader) ValuePointer() uint64 { return uint64(binary.LittleEndian.Uint64(kh[0:])) }

func (kh KeyHeader) ValueSize() uint32 { return uint32(binary.LittleEndian.Uint32(kh[8:])) }

func (kh KeyHeader) KeySize() uint16 { return uint16(binary.LittleEndian.Uint32(kh[12:])) }

func (kh KeyHeader) Crc16() uint16 { return uint16(binary.LittleEndian.Uint32(kh[14:])) }

func (kh KeyHeader) KeyAlignedSize() uint64 {
	return uint64((kh.KeySize() + KeyAlign - 1) & ^uint16(KeyAlign-1))
}

func (kh KeyHeader) SetValuePointer(vp uint64) { binary.LittleEndian.PutUint64(kh[0:], vp) }

func (kh KeyHeader) SetValueSize(size uint32) { binary.LittleEndian.PutUint32(kh[8:], size) }

func (kh KeyHeader) SetKeySize(size uint16) { binary.LittleEndian.PutUint16(kh[12:], size) }

func (kh KeyHeader) SetCrc16(crc uint16) { binary.LittleEndian.PutUint16(kh[14:], crc) }

//type KeyBlockHeader struct {
//	entries uint64
//  sortedKeyPointers []uint64
//}

type KeyBlockHeader []byte

const KeyBlockFixedHeaderSize = 8

func newKeyBlockHeader(b []byte) KeyBlockHeader { return KeyBlockHeader(b) }

func (kbh KeyBlockHeader) Get(key []byte, kvb *KVBos) ([]byte, bool) {

	entries := kbh.Entries()
	// TODO: Do binary search
	for e := uint32(0); e < entries; e++ {
		pItem := binary.LittleEndian.Uint64(kbh[KeyBlockFixedHeaderSize+e*8:])
		cmp := kvb.CompareKey(pItem, key)
		if cmp == 0 { // found
			pItemHdr := newKeyHeader(kvb.KeyBlocks[kvb.getKeyBlockIndex(pItem)][pItem&KeyBlockMask:], KeyHeaderSize)

			v := make([]byte, pItemHdr.ValueSize())
			vp := pItemHdr.ValuePointer()
			copy(v, kvb.ValueBlocks[vp>>ValueBlockShift][vp&ValueBlockMask:(vp&ValueBlockMask)+uint64(pItemHdr.ValueSize())])
			return v, true
		}
	}

	return []byte{}, false
}

func (kbh KeyBlockHeader) Entries() uint32 { return uint32(binary.LittleEndian.Uint32(kbh[0:])) }

func (kbh KeyBlockHeader) AddSortedPointer(keyPointer uint64, kvb *KVBos) {

	// TODO: Make sure sequential adds are optimized (allow fast bulk load of keys in sequential order)

	entries := kbh.Entries()

	startIndex := uint32(0)
	endIndex := entries

	// (Binary) search for position to add new pointer
	for startIndex < endIndex {
		cmp := kvb.Compare(binary.LittleEndian.Uint64(kbh[KeyBlockFixedHeaderSize+((startIndex+endIndex)>>1)*8:]), keyPointer)
		if cmp == 0 { // same item
			panic("Same item")
			return
		}
		center := (endIndex - startIndex) >> 1
		if center == 0 {
			break
		}
		if cmp == -1 {
			startIndex += center
		} else {
			endIndex -= center
		}
	}

	if startIndex+1 < endIndex {
		panic("move memory up")
	}

	// insert item into position
	binary.LittleEndian.PutUint64(kbh[KeyBlockFixedHeaderSize+endIndex*8:], keyPointer)

	binary.LittleEndian.PutUint32(kbh[0:], entries+1)
}

func (kbh KeyBlockHeader) Verify() {
}

func (kbh KeyBlockHeader) Dump() {
}

func (kvb *KVBos) Compare(a, b uint64) int {
	akh := newKeyHeader(kvb.KeyBlocks[kvb.getKeyBlockIndex(a)][a&KeyBlockMask:], KeyHeaderSize)
	bkh := newKeyHeader(kvb.KeyBlocks[kvb.getKeyBlockIndex(b)][b&KeyBlockMask:], KeyHeaderSize)

	pKeyDataA := (a & KeyBlockMask) - akh.KeyAlignedSize()
	pKeyDataB := (b & KeyBlockMask) - bkh.KeyAlignedSize()

	return bytes.Compare(kvb.KeyBlocks[kvb.getKeyBlockIndex(a)][pKeyDataA:pKeyDataA+uint64(akh.KeySize())], kvb.KeyBlocks[kvb.getKeyBlockIndex(b)][pKeyDataB:pKeyDataB+uint64(bkh.KeySize())])
}

func (kvb *KVBos) CompareKey(a uint64, key []byte) int {
	akh := newKeyHeader(kvb.KeyBlocks[kvb.getKeyBlockIndex(a)][a&KeyBlockMask:], KeyHeaderSize)
	pKeyDataA := (a & KeyBlockMask) - akh.KeyAlignedSize()

	return bytes.Compare(kvb.KeyBlocks[kvb.getKeyBlockIndex(a)][pKeyDataA:pKeyDataA+uint64(akh.KeySize())], key)
}

//func (kbh KeyBlockHeader) ScanFromBack(key []byte) []byte {
//
//	keyPointer := uint64(KeyBlockSize - 1)
//
//	for i := 0; i < 3; i++ {
//		pKeyHdr := keyPointer - (KeyHeaderSize - 1)
//		kh := newKeyHeader(KeyBlocks[0][pKeyHdr:], KeyHeaderSize)
//
//		pKeyData := pKeyHdr - kh.KeyAlignedSize()
//		if bytes.Compare( /*k*/ KeyBlocks[0][pKeyData:pKeyData+uint64(kh.KeySize())], key) == 0 {
//			fmt.Println("kh-e-y   f-o-u-n-d =", string(KeyBlocks[0][pKeyData:pKeyData+uint64(kh.KeySize())]))
//
//			v := make([]byte, kh.ValueSize())
//			vp := kh.ValuePointer()
//			copy(v, ValueBlocks[vp>>ValueBlockShift][vp&ValueBlockMask:(vp&ValueBlockMask)+uint64(kh.ValueSize())])
//			return v
//		}
//
//		keyPointer -= KeyHeaderSize + uint64(kh.KeyAlignedSize())
//	}
//
//	return []byte{}
//}
