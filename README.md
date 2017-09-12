
# KVBOS

kvbos is a **K**ey **V**alue **B**acked (by) **O**bject **S**torage.

It is a Key Value store relying on NVRAM for storing key/value pairs that is backed by Object Storage.

It is meant for very large datasets. By taking full advantage of emerging NVRAM trends, servers with up to 150 TB of total space can be created today. This will grow over the next one to two years to PB (Petabyte) scale size.

## Key ideas

- Separate keys and values (store in separate objects)
- Keys are stored in an immutable append-only fashion
- Use persistent addresses via 64-bit addresses
- Make snapshotting cheap and fast
- Use same format for in-memory and object storage (no conversions for IO)
- Do not optimize for size, eg avoid varints
- High performance and fully multi-threaded
- Simplictly in use, minimal config (ideally none)
- Fast restarts (esp. to accept writes again)

## Performance

```
BenchmarkPut100b-8   	 5000000	       297 ns/op	 336.61 MB/s
BenchmarkPut200b-8   	 5000000	       310 ns/op	 643.47 MB/s
```

Read from 100M keys (single thread)
```
BenchmarkGet10B-8   	  200000	      5940 ns/op
```

Total creation time using sequential ordering (key-size=8 bytes)

test case | time | IOPS (M) | key store size
----------|----:|------:|------:|
TestCreate10M | 4.0s | 2.50 | 0.3 GB
TestCreate20M | 7.7s |2.59 | 0.6 GB
TestCreate40M | 17.1s | 2.34 | 1.3 GB
TestCreate80M | 31.9s | 2.51 | 2.6 GB
TestCreate120M | 51.0s | 2.36 | 3.8 GB
TestCreate160M | 67.5s | 2.37 | 5.1 GB
TestCreate240M | 102.0s | 2.35 | 7.7 GB
TestCreate320M | 137.4s | 2.33 | 10 GB
TestCreate400M | 169.5s | 2.36 | 13 GB
TestCreate500M | 212.8s | 2.35 | 16 GB
TestCreate750M | 321.3s | 2.33 | 24 GB
TestCreate1000M | 393.1s | 2.54 | 32 GB

## Persistent Memory

Think of kvbos as creating a persistent memory space at a 64-bit address range scale whereby the values are stored uopwards from the bottom of the address space and the keys are stored downwards from the top of the address space.

Keys are grouped into immutable blocks. The upper part of a block contains the keys in the original order as they were added whereby each key has a pointer to its respective value. The bottom part of a key block is a arrays of pointers to its keys that is sorted on the keys so allow for quick (binary) searching of a key.

```
    +---------------------+  0xffffffffffffffff
    |---------------------|
    || Keys in AOF order ||
    ||  w ptrs to Values ||
    ||                   ||
    +---------------------+
    |                     |
    |                     |
    |                     |
    |                     |
    |                     |
    +---------------------+
    || Array of Key ptrs ||
    ||  (sorted)         ||
    ||                   ||
    |---------------------|
    +---------------------+  0xffffffffffff0000
```

Values are also grouped into immutable blocks and are stored in increasing memory order. In case there is not sufficient space in a block to store a new value, the current block will be closed and saved and a new block will be created to store the value.

```
    +---------------------+  0x0000000000010000
    |                     |
    |                     |
    |                     |
    |                     |
    |                     |
    +---------------------+
    || Values in AOF     ||
    ||  order            ||
    ||                   ||
    |---------------------|
    +---------------------+  0x0000000000000000
```

## Data structures

```
struct {
    key          [...]byte  // 64-bit aligned based on keySize
    unused       [...]byte  // 0 or more NULL bytes 
    valuePointer uint64
    valueSize    uint32
    keySize 	 uint16
    crc16        uint16
}
```

```
struct {
    value        [...]byte
}
```

## Snapshotting

Blocks (whether value or key blocks) are stored as is on object storage. Blocks that are not yet full can also be snapshotted and will be overwritten later with an updated version of the same block containing more values.

## Merging blocks

Blocks will be merged together in order to reduce the number of blocks as the number of blocks increases. Two consecutive value blocks will simply be concatenated (this process can be repeated many times).

When key blocks are merged a new combined sorted list of pointers is created so that searches are still fast  within the new key block.

## Trailing out of S3

A read-only copy can trail the database by first finding the 'tail' (lowest address) of the key blocks

## Server 

kvbos is designed as a server, much like redis.

## API Design

kvbos has a simple API inferface
- Put(key []byte, value []byte)
- Get(key []byte) value []byte
- Delete(key []byte)
- IteratorNext(key []byte)
- IteratorPrev(key []byte)

## Delete value

- use special value for ValueSize (eg == 0 or eg == 0xffff)
- use special value for ValuePointer (eg == 0x0000000000000000 or == 0xffffffffffffffff)

## Comparison to other KV Stores

In order to better understand where kvbos stands, here is a comparison to redis and rocksDB

|            | redis    | rocksDB    | kvbos
|------------|----------|------------|----------
| type       | server   | embeddable | server
| storage    | RAM      | disk       | NVRAM |
| max size   | RAM Size | disk       | NVRAM Size |
| persistent | disk     | disk       | Object storage |

### Bulk Load of keys in Sequential Order (1B)					

|            | minutes | MB/sec | ops/sec (K) | total data size
|------------|--------:|-------:|----------:|------------:
| kvbos      | 8       | 1560   | 2000      |  777 GB 
| rocksdb    | 36	   | 370    | 463       |  760 GB
| leveldb    | 91      | 146    | 183       |  760 GB

## Miscellaenous

- Use HTTP range GETs for values that are purged from memory
- It is up to the client to compress data or not (after all, the client knows best when to do this)

## Block storage

```
 32 Sep  1 19:00 test-val-0x0000000000000000
 32 Sep  1 19:00 test-val-0x0000000000000020
 32 Sep  1 19:00 test-val-0x0000000000000040
128 Sep  1 19:00 test-key-0xfffffffffffffe80
128 Sep  1 19:00 test-key-0xffffffffffffff00
128 Sep  1 19:00 test-key-0xffffffffffffff80
```

## Dump value blocks 

```
$ hexdump -C test-val-0x0000000000000000 
00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000010  70 6f 72 73 63 68 65 62  6d 77 00 00 00 00 00 00  |porschebmw......|
00000020
$ hexdump -C test-val-0x0000000000000020 
00000000  64 6f 6e 6b 65 72 76 6f  6f 72 74 66 65 72 72 61  |donkervoortferra|
00000010  72 69 6d 63 6c 61 72 65  6e 61 75 64 69 00 00 00  |rimclarenaudi...|
00000020
$ hexdump -C test-val-0x0000000000000040 
00000000  6d 65 72 63 65 64 65 73  6a 61 67 75 61 72 00 00  |mercedesjaguar..|
00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000020
```

## Dump key blocks 

```
$ hexdump -C test-key-0xffffffffffffff80 
00000000  03 00 00 00 00 00 00 00  f0 ff ff ff ff ff ff ff  |................|
00000010  d8 ff ff ff ff ff ff ff  c0 ff ff ff ff ff ff ff  |................|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000030  00 00 00 00 00 00 00 00  63 61 72 2d 32 00 00 00  |........car-2...|
00000040  20 00 00 00 00 00 00 00  0b 00 00 00 05 00 00 00  | ...............|
00000050  63 61 72 2d 31 00 00 00  17 00 00 00 00 00 00 00  |car-1...........|
00000060  03 00 00 00 05 00 00 00  63 61 72 2d 30 00 00 00  |........car-0...|
00000070  10 00 00 00 00 00 00 00  07 00 00 00 05 00 00 00  |................|
00000080
$ hexdump -C test-key-0xffffffffffffff00 
00000000  03 00 00 00 00 00 00 00  70 ff ff ff ff ff ff ff  |........p.......|
00000010  58 ff ff ff ff ff ff ff  40 ff ff ff ff ff ff ff  |X.......@.......|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
00000030  00 00 00 00 00 00 00 00  63 61 72 2d 35 00 00 00  |........car-5...|
00000040  39 00 00 00 00 00 00 00  04 00 00 00 05 00 00 00  |9...............|
00000050  63 61 72 2d 34 00 00 00  32 00 00 00 00 00 00 00  |car-4...2.......|
00000060  07 00 00 00 05 00 00 00  63 61 72 2d 33 00 00 00  |........car-3...|
00000070  2b 00 00 00 00 00 00 00  07 00 00 00 05 00 00 00  |+...............|
00000080
$ hexdump -C test-key-0xfffffffffffffe80 
00000000  02 00 00 00 00 00 00 00  f0 fe ff ff ff ff ff ff  |................|
00000010  d8 fe ff ff ff ff ff ff  00 00 00 00 00 00 00 00  |................|
00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
*
00000050  63 61 72 2d 37 00 00 00  48 00 00 00 00 00 00 00  |car-7...H.......|
00000060  06 00 00 00 05 00 00 00  63 61 72 2d 36 00 00 00  |........car-6...|
00000070  40 00 00 00 00 00 00 00  08 00 00 00 05 00 00 00  |@...............|
00000080
```

## Ideas

Split between
- kvbos-engine (core embeddable engine)
- kvbos (Redis like interface)

## Limitations

- Value size is limited to 32-bits
- Key size is limited to 16-bits
- Values larger than the Value Block size cannot be stored
