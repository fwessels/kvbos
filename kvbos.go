package kvbos

type Value struct {
	value []byte
}

type KVBos struct {
}

const (
	ValueBlockShift = 5
	ValueBlockSize  = 1 << ValueBlockShift
	ValueBlockMask  = ValueBlockSize - 1
	KeyBlockShift   = 8
	KeyBlockSize    = 1 << KeyBlockShift
	KeyBlockMask    = KeyBlockSize - 1
	KeyAlign        = 8
)

type ValueBlock [ValueBlockSize]byte
type KeyBlock [KeyBlockSize]byte

var ValueBlocks [10]ValueBlock
var ValuePointer = uint64(0x0000000000000010) // Skip first 0x10 bytes (prevent 'NULL' from being a valid value pointer)

var KeyBlocks [10]KeyBlock
var KeyPointer = uint64(0xffffffffffffffff)

func (kvb *KVBos) Put(key []byte, value []byte) {

	// Copy value into block
	valuePointer := ValuePointer
	valueSize := uint32(len(value))
	keySize := uint32(len(key))
	if valueSize >= ValueBlockSize {
		// TODO: Could "overflow" value into a new block (provided it is smaller than uint32)
		// and would require multiple blocks to read (or multiple HTTP Range GETs)
		panic("Attempting to store item larger than value block size")
	} else if (valuePointer&ValueBlockMask)+uint64(valueSize) >= ValueBlockSize {
		valuePointer = ((valuePointer >> ValueBlockShift) + 1) << ValueBlockShift // advance to beginning of next block
	}
	copy(ValueBlocks[valuePointer>>ValueBlockShift][valuePointer&ValueBlockMask:], value[:])

	ValuePointer = valuePointer + uint64(valueSize)

	kh := newKeyHeader(make([]byte, KeyHeaderSize), KeyHeaderSize)
	kh.SetValuePointer(valuePointer)
	kh.SetValueSize(valueSize)
	kh.SetKeySize(keySize)

	// Copy key header first (has a deterministic size,
	// so we can iterate manually 'downwards' in memory if need be)
	copy(KeyBlocks[0][(KeyPointer&KeyBlockMask)-uint64(KeyHeaderSize-1):], kh[:])

	// Copy key itself
	keyAlignedSize := kh.KeyAlignedSize()
	copy(KeyBlocks[0][(KeyPointer&KeyBlockMask)-uint64(KeyHeaderSize+keyAlignedSize-1):], key[:])

	KeyPointer -= KeyHeaderSize + uint64(keyAlignedSize)

	kbh := newKeyBlockHeader(KeyBlocks[0][:])
	kbh.AddSortedPointer(KeyPointer + 1 + keyAlignedSize)
}

func (kvb *KVBos) Get(key []byte) []byte {

	kbh := newKeyBlockHeader(KeyBlocks[0][:])
	return kbh.Get(key)
}
