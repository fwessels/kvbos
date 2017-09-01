package kvbos

import (
	_ "fmt"
	"testing"
)

func TestKVBos(t *testing.T) {

	kvb := KVBos{}

	kvb.Put([]byte("car-0"), []byte("porsche"))
	kvb.Put([]byte("car-1"), []byte("bmw"))
	kvb.Put([]byte("car-2"), []byte("donkervoort"))
	//kvb.Put([]byte("car-3"), []byte("ferrari"))
	//kvb.Put([]byte("car-4"), []byte("mclaren"))
	//kvb.Put([]byte("car-5"), []byte("audi"))
	//
	//fmt.Println(string(kvb.Get([]byte("car-1"))))
	//fmt.Println(string(kvb.Get([]byte("car-4"))))

	Snapshot("test")
}
