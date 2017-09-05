package kvbos

import (
	"fmt"
	"testing"
	"io"
	"crypto/rand"
	"encoding/binary"
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

func benchmarkPuts(b *testing.B, valSize int64) {

	kvb := KVBos{}

	keyCounter := uint64(1)
	key := make([]byte, 8)

	value := make([]byte, valSize)
	if _, err := io.ReadFull(rand.Reader, value); err != nil {
		b.Fatalf("Failed to generate random value: %v", err)
	}

	b.SetBytes(valSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		keyCounter++
		binary.LittleEndian.PutUint64(key[0:], keyCounter)

		kvb.Put(key, value)
	}
}

func BenchmarkPuts800b(b *testing.B) {
	benchmarkPuts(b, 800)
}