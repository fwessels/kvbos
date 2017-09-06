package kvbos

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
)

func TestKVBos(t *testing.T) {

	kvb := NewKVBos()

	kvb.Put([]byte("car-00"), []byte("porsche"))
	kvb.Put([]byte("car-01"), []byte("bmw"))
	kvb.Put([]byte("car-02"), []byte("donkervoort"))
	kvb.Put([]byte("car-03"), []byte("ferrari"))
	kvb.Put([]byte("car-04"), []byte("mclaren"))
	kvb.Put([]byte("car-05"), []byte("audi"))
	kvb.Put([]byte("car-06"), []byte("mercedes"))
	kvb.Put([]byte("car-07"), []byte("jaguar"))
	//
	// empty keys have value size = 0 and current value of value pointer
	kvb.Put([]byte("car-08"), []byte(""))
	kvb.Put([]byte("car-09"), []byte(""))
	// both have same value pointer

	kvb.Put([]byte("car-10"), []byte("pagani"))
	kvb.Put([]byte("car-11"), []byte("astonmartin"))
	kvb.Put([]byte("car-12"), []byte("lamborghini"))
	kvb.Put([]byte("car-13"), []byte("lotus"))

	fmt.Println(string(kvb.Get([]byte("car-01"))))
	fmt.Println(string(kvb.Get([]byte("car-04"))))
	fmt.Println(string(kvb.Get([]byte("car-08"))))
	fmt.Println(string(kvb.Get([]byte("car-12"))))

	kvb.Snapshot("test")
}

var keyCounter = uint64(128)

func benchmarkPuts(b *testing.B, valSize int64) {

	key := make([]byte, 8)

	value := make([]byte, valSize)
	if _, err := io.ReadFull(rand.Reader, value); err != nil {
		b.Fatalf("Failed to generate random value: %v", err)
	}

	b.SetBytes(valSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		cntr := atomic.AddUint64(&keyCounter, 256)

		binary.BigEndian.PutUint64(key[0:], cntr) // Use big endian for sequential ordering

		kvbBenchmark.Put(key, value)
	}

	//kvbBenchmark.Snapshot("benchmark")
}

var kvbBenchmark = NewKVBos()

func BenchmarkPuts800b(b *testing.B) {
	benchmarkPuts(b, 200)
}
