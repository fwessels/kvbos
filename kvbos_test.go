package kvbos

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"sync"
	mr "math/rand"
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

const million = 1000000

func testCreate(entries uint64, valSize int64) (*KVBos, uint64) {

	kvb, keyCounter := NewKVBos(), uint64(0)

	wg := sync.WaitGroup{}
	for gr := 0; gr < 1; gr++ {
		wg.Add(1)
		go func() {

			key := make([]byte, 8)
			value := make([]byte, valSize)
			if _, err := io.ReadFull(rand.Reader, value); err != nil {
				panic("Failed to generate random value")
			}
			for {
				cntr := atomic.AddUint64(&keyCounter, 1)
				if cntr >= entries {
					break
				}

				binary.BigEndian.PutUint64(key[0:], cntr) // Use big endian for sequential ordering

				kvb.Put(key, value)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	return kvb, atomic.LoadUint64(&keyCounter)
}

func TestCreate10M(t *testing.T) { testCreate(10*million, 1) }
func TestCreate20M(t *testing.T) { testCreate(20*million, 1) }
func TestCreate40M(t *testing.T) { testCreate(40*million, 1) }
func TestCreate80M(t *testing.T) { testCreate(80*million, 1) }
func TestCreate160M(t *testing.T) { testCreate(160*million, 1) }
func TestCreate240M(t *testing.T) { testCreate(240*million, 1) }
func TestCreate320M(t *testing.T) { testCreate(320*million, 1) }
func TestCreate400M(t *testing.T) { testCreate(400*million, 1) }
func TestCreate500M(t *testing.T) { testCreate(500*million, 1) }
func TestCreate750M(t *testing.T) { testCreate(750*million, 1) }
func TestCreate1000M(t *testing.T) { testCreate(1000*million, 1) }

func benchmarkPut(b *testing.B, valSize int64) {

	keyCounter := uint64(0)

	kvbBenchmark := NewKVBos()

	key := make([]byte, 8)

	value := make([]byte, valSize)
	if _, err := io.ReadFull(rand.Reader, value); err != nil {
		b.Fatalf("Failed to generate random value: %v", err)
	}

	b.SetBytes(valSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		cntr := atomic.AddUint64(&keyCounter, 1)

		binary.BigEndian.PutUint64(key[0:], cntr) // Use big endian for sequential ordering

		kvbBenchmark.Put(key, value)
	}

	//kvbBenchmark.Snapshot("benchmark")
}

func BenchmarkPut20B(b *testing.B) { benchmarkPut(b, 20) }
func BenchmarkPut100B(b *testing.B) { benchmarkPut(b, 100) }
func BenchmarkPut200B(b *testing.B) { benchmarkPut(b, 200) }

var kvbGets *KVBos
var kvbGetsEntries uint64

func benchmarkGet(b *testing.B, valSize int64) {

	if kvbGets == nil {
		fmt.Println("Create test db ...")
		kvbGets, kvbGetsEntries = testCreate(100*million, valSize)
		fmt.Println("Done creating test db")
		fmt.Println("  key blocks =", kvbGets.GetKeyBlocks())
		fmt.Println("     entries =", kvbGetsEntries)
	}

	key := make([]byte, 8)

	b.SetBytes(valSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		cntr := mr.Int63n(int64(kvbGetsEntries))
		binary.BigEndian.PutUint64(key[0:], uint64(cntr)) // Use big endian for sequential ordering

		kvbGets.Get(key)
	}
}

func BenchmarkGet10B(b *testing.B) { benchmarkGet(b, 10) }
