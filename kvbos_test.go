package kvbos

import (
	"testing"
	"fmt"
)

func TestKVBos(t *testing.T) {

	kvb := KVBos{}

	kvb.Put([]byte("car-0"), []byte("porsche"))
	kvb.Put([]byte("car-1"), []byte("bmw"))
	kvb.Put([]byte("car-2"), []byte("audi"))

	fmt.Println(string(kvb.Get([]byte("car-1"))))

	Snapshot()
}
