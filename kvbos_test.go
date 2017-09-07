package kvbos

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"sync"
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

func testCreate(t *testing.T, entries uint64) {

	kvb, keyCounter := NewKVBos(), uint64(0)

	wg := sync.WaitGroup{}
	for gr := 0; gr < 1; gr++ {
		wg.Add(1)
		go func() {

			key := make([]byte, 8)
			value := make([]byte, 1)
			if _, err := io.ReadFull(rand.Reader, value); err != nil {
				t.Fatalf("Failed to generate random value: %v", err)
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
}

func TestCreate10M(t *testing.T) { testCreate(t, 10*million) }
func TestCreate20M(t *testing.T) { testCreate(t, 20*million) }
func TestCreate40M(t *testing.T) { testCreate(t, 40*million) }
func TestCreate80M(t *testing.T) { testCreate(t, 80*million) }
func TestCreate160M(t *testing.T) { testCreate(t, 160*million) }
func TestCreate240M(t *testing.T) { testCreate(t, 240*million) }
func TestCreate320M(t *testing.T) { testCreate(t, 320*million) }
func TestCreate400M(t *testing.T) { testCreate(t, 400*million) }
func TestCreate500M(t *testing.T) { testCreate(t, 500*million) }
func TestCreate750M(t *testing.T) { testCreate(t, 750*million) }
func TestCreate1000M(t *testing.T) { testCreate(t, 1000*million) }

func benchmarkPuts(b *testing.B, valSize int64) {

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


func BenchmarkPuts20B(b *testing.B) { benchmarkPuts(b, 20) }
func BenchmarkPuts100B(b *testing.B) { benchmarkPuts(b, 100) }
func BenchmarkPuts200B(b *testing.B) { benchmarkPuts(b, 200) }

func benchmarkGets(b *testing.B, valSize int64) {

}