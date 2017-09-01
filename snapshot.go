package kvbos

import (
	"io/ioutil"
	"fmt"
)

func Snapshot() {

	for i, vb := range ValueBlocks {
		filename := fmt.Sprintf("value-%016x", i * ValueBlockSize)
		ioutil.WriteFile(filename, vb[:], 0644)
	}
}