package kvbos

import (
	"encoding/binary"
	"bytes"
	"fmt"
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

func (kbh KeyBlockHeader) Get(key []byte) []byte {

	entries := kbh.Entries()
	// TODO: Do binary search
	for e := uint32(0); e < entries; e++ {
		pItem := binary.LittleEndian.Uint64(kbh[(e+1)*8:])
		cmp := CompareKey(pItem, key)
		if cmp == 0 { // found
			pItemHdr := newKeyHeader(KeyBlocks[0][pItem:], KeyHeaderSize)

			v := make([]byte, pItemHdr.ValueSize())
			copy(v, ValueBlocks[0][pItemHdr.ValuePointer():pItemHdr.ValuePointer()+uint64(pItemHdr.ValueSize())])
			return v
		}
	}

	return []byte{}
}

func (kbh KeyBlockHeader) Entries() uint32 { return uint32(binary.LittleEndian.Uint32(kbh[0:])) }

func (kbh KeyBlockHeader) AddSortedPointer(keyPointer uint64) {

	entries := kbh.Entries()
	// TODO: Do binary search
	for e := uint32(0); e < entries; e++ {
		cmp := Compare(binary.LittleEndian.Uint64(kbh[(e+1)*8:]), keyPointer)
		if cmp == 1 { // insert item before
			panic("Insert item before")
		} else if cmp == 0 { // same item
			panic("Same item")
			return
		}
		fmt.Println(cmp)
	}
	// insert item at the end
	binary.LittleEndian.PutUint64(kbh[(entries+1)*8:], keyPointer)

	binary.LittleEndian.PutUint32(kbh[0:], entries + 1)
}

func Compare(a, b uint64) int {
	akh := newKeyHeader(KeyBlocks[0][a&KeyBlockMask:], KeyHeaderSize)
	bkh := newKeyHeader(KeyBlocks[0][b&KeyBlockMask:], KeyHeaderSize)

	pKeyDataA := (a&KeyBlockMask)-akh.KeyAlignedSize()
	pKeyDataB := (b&KeyBlockMask)-bkh.KeyAlignedSize()

	return bytes.Compare(KeyBlocks[0][pKeyDataA:pKeyDataA+uint64(akh.KeySize())], KeyBlocks[0][pKeyDataB:pKeyDataB+uint64(bkh.KeySize())])
}

func CompareKey(a uint64, key []byte) int {
	akh := newKeyHeader(KeyBlocks[0][a:], KeyHeaderSize)
	pKeyDataA := a-akh.KeyAlignedSize()

	return bytes.Compare(KeyBlocks[0][pKeyDataA:pKeyDataA+uint64(akh.KeySize())], key)
}

func (kbh KeyBlockHeader) ScanFromBack(key []byte) []byte {

	keyPointer := uint64(KeyBlockSize-1)

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