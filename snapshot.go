package kvbos

import (
	"fmt"
	"io/ioutil"
)

func Snapshot(db string) {
	for i := uint64(0x0000000000000000); i>>ValueBlockShift <= ValuePointer>>ValueBlockShift; i += ValueBlockSize {
		filename := fmt.Sprintf("%s-val-0x%016x", db, i)
		ioutil.WriteFile(filename, ValueBlocks[i>>ValueBlockShift][:], 0644)
	}

	for i := uint64(0xffffffffffffffff) - KeyBlockMask; i>>KeyBlockShift >= KeyPointer>>KeyBlockShift; i -= KeyBlockSize {
		filename := fmt.Sprintf("%s-key-0x%016x", db, i)
		ioutil.WriteFile(filename, KeyBlocks[getKeyBlockIndex(i)][:], 0644)
	}
}
