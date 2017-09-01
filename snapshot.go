package kvbos

import (
	"fmt"
	"io/ioutil"
)

func Snapshot(db string) {

	for i, vb := range ValueBlocks {
		if uint64(i)*ValueBlockSize >= ValuePointer {
			break
		}
		filename := fmt.Sprintf("%s-value-0x%016x", db, i*ValueBlockSize)
		ioutil.WriteFile(filename, vb[:], 0644)
	}

	for i, kb := range KeyBlocks {
		if i == 1 {
			break
		}
		filename := fmt.Sprintf("%s-key-0x%016x", db, uint64(0xffffffffffffffff)-uint64(i+1)*KeyBlockSize+1)
		ioutil.WriteFile(filename, kb[:], 0644)

	}
}
