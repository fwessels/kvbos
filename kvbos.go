package kvbos

import (
	"sync"
)

const (
	ValueBlockShift = /*8//*/ 26
	KeyBlockShift   = /*8//*/ 19+1+3+1+1+1+1+1+1
)

type KVBos struct {
	ValueBlocks  [28]ValueBlock
	ValuePointer uint64

	KeyBlockWarm KeyBlock
	KeyBlocksCold []KeyBlockCold // TODO: Protect cold blocks with atomic.Value (ReadMostly)
	KeyPointer   uint64
	KeyLock      sync.Mutex

	DBName string

}

type KeyBlockCold struct {
	mask  uint64
	block []byte
}

type Value struct {
	value []byte
}

func NewKVBos(dbname string) *KVBos {
	return &KVBos{
		ValuePointer: uint64(0x0000000000000010), // Skip first 0x10 bytes (prevent 'NULL' from being a valid value pointer),
		KeyPointer:   uint64(0xffffffffffffffff), // Start at the end of the address range
		DBName:       dbname}
}

const (
	ValueBlockSize = 1 << ValueBlockShift
	ValueBlockMask = ValueBlockSize - 1
	KeyBlockSize   = 1 << KeyBlockShift
	KeyBlockMask   = KeyBlockSize - 1
	KeyAlign       = 8
)

type ValueBlock [ValueBlockSize]byte
type KeyBlock [KeyBlockSize]byte

var KeyBlockEmpty KeyBlock

// Put
func (kvb *KVBos) Put(key []byte, value []byte) {

	valuePointer := kvb.putAtomic(key, value)

	// Copy value into value block
	copy(kvb.ValueBlocks[valuePointer>>ValueBlockShift][valuePointer&ValueBlockMask:], value)
}

// putAtomic -- atomic part of Put() operation
// KeyPointer and ValuePointer are always adjusted in tandem, ie
// the lowest (=latest) KeyPointer is associated with the highest
// (=latest) ValuePointer
func (kvb *KVBos) putAtomic(key []byte, value []byte) uint64 {

	kvb.KeyLock.Lock()

	// Prepare value pointer
	valuePointer := kvb.ValuePointer
	valueSize := uint32(len(value))
	keySize := uint16(len(key))
	if valueSize >= ValueBlockSize {
		// TODO: Could "overflow" value into a new block (provided it is smaller than uint32)
		// and would require multiple blocks to read (or multiple HTTP Range GETs)
		panic("Attempting to store item larger than value block size")
	} else if (valuePointer&ValueBlockMask)+uint64(valueSize) >= ValueBlockSize {
		//fmt.Println("New value block")
		valuePointer = ((valuePointer >> ValueBlockShift) + 1) << ValueBlockShift // advance to beginning of next block
	}

	kvb.ValuePointer = valuePointer + uint64(valueSize)

	kh := newKeyHeader(make([]byte, KeyHeaderSize), KeyHeaderSize)
	kh.SetValuePointer(valuePointer)
	kh.SetValueSize(valueSize)
	kh.SetKeySize(keySize)
	kh.SetCrc16(0x1234)

	// Compute the boundaries from the beginning and end of the key block
	lowWaterMark := uint64(KeyBlockFixedHeaderSize + newKeyBlockHeader(kvb.KeyBlockWarm[:]).Entries()*8)
	highWaterMark := kvb.KeyPointer&KeyBlockMask + 1
	if KeyHeaderSize+kh.KeyAlignedSize()+8 > highWaterMark-lowWaterMark {
		// Not enough space left to store this key, so advance to the next key block
		kvb.Snapshot()

		// Add block to cold blocks
		block := make([]byte, len(kvb.KeyBlockWarm))
		copy(block[:], kvb.KeyBlockWarm[:])
		kvb.KeyBlocksCold = append(kvb.KeyBlocksCold, KeyBlockCold{mask: KeyBlockMask, block: block})

		// Advance pointer
		kvb.KeyPointer = (((kvb.KeyPointer >> KeyBlockShift) - 1) << KeyBlockShift) + KeyBlockMask

		mergeSize := uint64(1 << (KeyBlockShift + 1))
		for kp := kvb.KeyPointer + 1; kp+(mergeSize>>1) != 0x0000000000000000; mergeSize *= 2  { // Until there are blocks to be merged

			if (kp+mergeSize)&(mergeSize-1) != 0x0 {
				break // Break out when upper half does not match size of lower half
			}
			CombineKeyBlocks(kvb.DBName, kp, kp+(mergeSize>>1))
		}

		// clear block for next iteration
		copy(kvb.KeyBlockWarm[:], KeyBlockEmpty[:])
	}

	// Copy key header first (has a deterministic size,
	// so we can iterate manually 'downwards' in memory if need be)
	copy(kvb.KeyBlockWarm[(kvb.KeyPointer&KeyBlockMask)-uint64(KeyHeaderSize-1):], kh[:])

	// Copy key itself
	keyAlignedSize := kh.KeyAlignedSize()
	copy(kvb.KeyBlockWarm[(kvb.KeyPointer&KeyBlockMask)-uint64(KeyHeaderSize+keyAlignedSize-1):], key[:])

	kvb.KeyPointer -= KeyHeaderSize + uint64(keyAlignedSize)

	kbh := newKeyBlockHeader(kvb.KeyBlockWarm[:])
	kbh.AddSortedPointer(kvb.KeyPointer+1+keyAlignedSize, kvb)

	kvb.KeyLock.Unlock()

	return valuePointer
}

// Get
func (kvb *KVBos) Get(key []byte) []byte {

	// Search active key block first (sorted pointer map may be modified by writes)
	val, found := kvb.getAtomic(key)
	if found {
		return val
	}

	for _, kbl := range kvb.KeyBlocksCold {
		kbh := newKeyBlockHeader(kbl.block)
		val, found = kbh.Get(key, kvb, kbl.mask)
		if found {
			return val
		}
	}

	return []byte{}
}

// getAtomic -- atomic part of Get() operation
func (kvb *KVBos) getAtomic(key []byte) (val []byte, found bool) {

	kvb.KeyLock.Lock()
	kbh := newKeyBlockHeader(kvb.KeyBlockWarm[:])
	val, found = kbh.Get(key, kvb, KeyBlockMask)
	kvb.KeyLock.Unlock()

	return
}

// Delete
func (kvb *KVBos) Delete(key []byte) {
	// Add key with value pointer = NULL/0x0 and value size = 0
}
