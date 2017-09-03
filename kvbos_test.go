package kvbos

import (
	"fmt"
	"testing"
)

func TestKVBos(t *testing.T) {

	kvb := KVBos{}

	kvb.Put([]byte("car-0"), []byte("porsche"))
	kvb.Put([]byte("car-1"), []byte("bmw"))
	kvb.Put([]byte("car-2"), []byte("donkervoort"))
	kvb.Put([]byte("car-3"), []byte("ferrari"))
	kvb.Put([]byte("car-4"), []byte("mclaren"))
	kvb.Put([]byte("car-5"), []byte("audi"))
	kvb.Put([]byte("car-6"), []byte("mercedes"))
	kvb.Put([]byte("car-7"), []byte("jaguar"))

	// empty keys have value size = 0 and current value of value pointer
	kvb.Put([]byte("car-8"), []byte(""))
	kvb.Put([]byte("car-9"), []byte(""))
	// both have same value pointer

	fmt.Println(string(kvb.Get([]byte("car-1"))))
	fmt.Println(string(kvb.Get([]byte("car-4"))))
	fmt.Println(string(kvb.Get([]byte("car-8"))))

	Snapshot("test")
}
