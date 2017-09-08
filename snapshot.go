package kvbos

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func (kvb *KVBos) Snapshot(db string) {
	// TODO: Prevent resaving same block
	for i := uint64(0x0000000000000000); i>>ValueBlockShift <= kvb.ValuePointer>>ValueBlockShift; i += ValueBlockSize {
		filename := fmt.Sprintf("%s-val-0x%016x", db, i)
		ioutil.WriteFile(filename, kvb.ValueBlocks[i>>ValueBlockShift][:], 0644)
	}

	for i := uint64(0xffffffffffffffff) - KeyBlockMask; i>>KeyBlockShift >= kvb.KeyPointer>>KeyBlockShift; i -= KeyBlockSize {
		filename := fmt.Sprintf("%s-key-0x%016x", db, i)
		ioutil.WriteFile(filename, kvb.KeyBlocks[kvb.getKeyBlockIndex(i)][:], 0644)
	}
}

func Load(db string, startBlock uint64) {
	// derive KeyPointer & ValuePointer from existing blocks

	// consider adding CRC to KeyStruct to detect tampering/bad stores

	filename := fmt.Sprintf("%s-key-0x%016x", db, startBlock)

	//var err error
	keyBlock, _ := ioutil.ReadFile(filename)

	endBlock := startBlock + uint64(len(keyBlock)-1)

	fmt.Printf("endBlock: %016x\n", endBlock)

	header := newKeyBlockHeader(keyBlock)

	fmt.Println(header.Entries())

	freeBlock, valuePointer := header.GetFreeAddress(endBlock, uint64(len(keyBlock)))
	fmt.Printf("   freeBlock: %016x\n", freeBlock)
	fmt.Printf("valuePointer: %016x\n", valuePointer)
}

func CombineKeyBlocks(db string, startBlockLo, startBlockHi uint64) {

	filenameLo := fmt.Sprintf("%s-key-0x%016x", db, startBlockLo)
	filenameHi := fmt.Sprintf("%s-key-0x%016x", db, startBlockHi)

	//var err error
	keyBlockLo, _ := ioutil.ReadFile(filenameLo)
	keyBlockHi, _ := ioutil.ReadFile(filenameHi)

	endBlockLo := startBlockLo + uint64(len(keyBlockLo)-1)
	endBlockHi := startBlockHi + uint64(len(keyBlockHi)-1)

	headerLo := newKeyBlockHeader(keyBlockLo)
	headerHi := newKeyBlockHeader(keyBlockHi)

	fmt.Println(headerLo.Entries())
	fmt.Println(headerHi.Entries())

	sortedPointersLo := make([]byte, headerLo.Entries()*8)
	sortedPointersHi := make([]byte, headerHi.Entries()*8)

	freeBlockLo, _ := headerLo.GetFreeAddress(endBlockLo, uint64(len(keyBlockLo)))
	freeBlockHi ,_ := headerHi.GetFreeAddress(endBlockHi, uint64(len(keyBlockHi)))
	shift := freeBlockHi - endBlockLo

	// Adjust pointers for low block
	for p := 0; p < len(sortedPointersLo); p += 8 {
		binary.LittleEndian.PutUint64(sortedPointersLo[p:], shift+binary.LittleEndian.Uint64(keyBlockLo[KeyBlockFixedHeaderSize+p:KeyBlockFixedHeaderSize+p+8]))
	}
	// Save pointers for high block
	copy(sortedPointersHi, keyBlockHi[KeyBlockFixedHeaderSize:KeyBlockFixedHeaderSize+headerHi.Entries()*8])

	// Copy keys up
	copy(keyBlockHi, keyBlockLo[uint64(len(keyBlockLo))-shift:])
	copy(keyBlockLo[uint64(len(keyBlockLo))-(endBlockLo-freeBlockLo)+shift:], keyBlockLo[uint64(len(keyBlockLo))-(endBlockLo-freeBlockLo):uint64(len(keyBlockLo))-shift])

	// Zero out remaining keys
	for p := freeBlockLo + 1; p < freeBlockLo+1+shift; p += 8 {
		binary.LittleEndian.PutUint64(keyBlockLo[p&KeyBlockMask:], 0)
	}

	// TODO: combine sorted pointer lists by merge sorting them
	// TODO: determine duplicate entries (same key)
	duplicateEntries := uint64(0)
	copy(keyBlockLo[KeyBlockFixedHeaderSize:], sortedPointersHi)
	copy(keyBlockLo[KeyBlockFixedHeaderSize+len(sortedPointersHi):], sortedPointersLo)

	headerLo.SetEntries(uint64(len(sortedPointersHi)+len(sortedPointersLo))/8 - duplicateEntries)

	ConcatFiles(filenameLo+".merged", keyBlockLo[:], keyBlockHi[:], 0644)

	os.Rename(filenameLo, filenameLo+".bak") // os.Remove(filenameLo)
	os.Rename(filenameLo+".merged", filenameLo)
	os.Rename(filenameHi, filenameHi+".bak") // os.Remove(filenameHi)

	// b l o c k   1
	//$ hexdump -C test-key-0xffffffffffffff00
	//00000000  03 00 00 00 00 00 00 00  70 ff ff ff ff ff ff ff  |........p.......|
	//00000010  58 ff ff ff ff ff ff ff  40 ff ff ff ff ff ff ff  |X.......@.......|
	//00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//00000030  00 00 00 00 00 00 00 00  63 61 72 2d 35 00 00 00  |........car-5...|
	//00000040  39 00 00 00 00 00 00 00  04 00 00 00 05 00 00 00  |9...............|
	//00000050  63 61 72 2d 34 00 00 00  32 00 00 00 00 00 00 00  |car-4...2.......|
	//00000060  07 00 00 00 05 00 00 00  63 61 72 2d 33 00 00 00  |........car-3...|
	//00000070  2b 00 00 00 00 00 00 00  07 00 00 00 05 00 00 00  |+...............|
	//00000080

	// b l o c k   2
	//$ hexdump -C test-key-0xffffffffffffff80
	//00000000  03 00 00 00 00 00 00 00  f0 ff ff ff ff ff ff ff  |................|
	//00000010  d8 ff ff ff ff ff ff ff  c0 ff ff ff ff ff ff ff  |................|
	//00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//00000030  00 00 00 00 00 00 00 00  63 61 72 2d 32 00 00 00  |........car-2...|
	//00000040  20 00 00 00 00 00 00 00  0b 00 00 00 05 00 00 00  | ...............|
	//00000050  63 61 72 2d 31 00 00 00  17 00 00 00 00 00 00 00  |car-1...........|
	//00000060  03 00 00 00 05 00 00 00  63 61 72 2d 30 00 00 00  |........car-0...|
	//00000070  10 00 00 00 00 00 00 00  07 00 00 00 05 00 00 00  |................|
	//00000080

	// c o m b i n e d
	//00000000  06 00 00 00 00 00 00 00  f0 ff ff ff ff ff ff ff  |................|
	//00000010  d8 ff ff ff ff ff ff ff  c0 ff ff ff ff ff ff ff  |................|
	//00000020  a8 ff ff ff ff ff ff ff  90 ff ff ff ff ff ff ff  |........p.......|
	//00000030  78 ff ff ff ff ff ff ff  00 00 00 00 00 00 00 00  |X.......@.......|
	//00000040  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//00000050  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//00000060  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//00000070  63 61 72 2d 35 00 00 00  39 00 00 00 00 00 00 00  |car-5...9.......|
	//00000080  04 00 00 00 05 00 00 00  63 61 72 2d 34 00 00 00  |........car-4...|
	//00000090  32 00 00 00 00 00 00 00  07 00 00 00 05 00 00 00  |2...............|
	//000000a0  63 61 72 2d 33 00 00 00  2b 00 00 00 00 00 00 00  |car-3...+.......|
	//000000b0  07 00 00 00 05 00 00 00  63 61 72 2d 32 00 00 00  |........car-2...|
	//000000c0  20 00 00 00 00 00 00 00  0b 00 00 00 05 00 00 00  | ...............|
	//000000d0  63 61 72 2d 31 00 00 00  17 00 00 00 00 00 00 00  |car-1...........|
	//000000e0  03 00 00 00 05 00 00 00  63 61 72 2d 30 00 00 00  |........car-0...|
	//000000f0  10 00 00 00 00 00 00 00  07 00 00 00 05 00 00 00  |................|
	//00000100
}

func ConcatFiles(filename string, dataLo, dataHi []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(dataLo)
	if err == nil && n < len(dataLo) {
		err = io.ErrShortWrite
	}
	n, err = f.Write(dataHi)
	if err == nil && n < len(dataHi) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func CombineValueBlocks() {
	// Concat two value blocks into larger one and delete old blocks
}

func TrimValueBlock() {
	// Scan from end of the block as to whether value pointer is no longer used (corresponding
	// key deleted) until first active value pointer is found.
	// If scanned all the way to be beginning of the block, delete block altogether
	// Range of keys scanned to be scanned is determined by address range of value block
}

func TrimKeyBlock() {
	// If all keys are tombstoned, consider removing block completely
}
