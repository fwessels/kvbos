package kvbos

import (
	"testing"
	_ "fmt"
)

func TestKVBos(t *testing.T) {

	kvb := KVBos{}

	kvb.Put([]byte("car-0"), []byte("porsche"))
	kvb.Put([]byte("car-1"), []byte("bmw"))
	kvb.Put([]byte("car-2"), []byte("audi"))
	kvb.Put([]byte("car-3"), []byte("donkervoort"))
	kvb.Put([]byte("car-4"), []byte("ferrari"))
	kvb.Put([]byte("car-5"), []byte("mclaren"))

	//fmt.Println(string(kvb.Get([]byte("car-1"))))

	Snapshot("test")
}
