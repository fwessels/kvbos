
# KVBOS

kvbos is a **K**ey **V**alue **B**acked (on) **O**bject **S**torage.

It is a Key Value store relying on storing key/value pairs in NVRAM that is backed by Object Storage.

It is meant for very large datasets. By taking full advantage of emerging NVRAM trends, servers with up to 150 TB of total space can be created today. This will grow over the next one to two years to PB (Petabyte) scale size.

## Persistent Memory

Think of kvbos as creating a persistent memory space at a 64-bit address range scale.

Think of a 64-bit address space whereby the values are stored from te bottom of the address space and the keys are stored from the top of the address space.

Values are grouped into immutable blocks (default 4 MB). Values are stored in increasing memory order (either aligned or not). In case there is not sufficient space in a block to store a new value, the current block will be closed and saved and a new block will be created to store the value.

Keys are also grouped into immutable blocks of size 512 KB. The upper part of a block contains the keys in unsorted order whereby each key has a pointer to its respective value. The bottom part of a key block is a sorted list of pointers to its keys.

## Grouping key blocks

**After merging the sorted list of (pointers to) keys will typically contain entries that point to keys outside of its own block.** But this is perfectly fine as we are using persistent pointers.

Once we have too many blocks, it is possible to create a key block without any keys itself but just pointers, this way it is possbile to prevent too many 'hops' between key blocks in order to (binary) search for a key.

## Snapshotting

Value blocks are stored as is on object storage. Blocks that are not yet full can also be snapshotted and will be overwritten later with an updated version of the same block containing more values.

Key blocks

## Trailing out of S3

A read-only copy can trail the database by first finding the 'tail' (lowest address) of the key blocks

## Server 

kvbos is designed as a server, much like redis.

## API Design

kvbos has a simple API inferface

## Delete value

- use special value for ValueSize (eg == 0 or eg == 0xffff)
- use special value for ValuePointer (eg == 0x0000000000000000 or == 0xffffffffffffffff)

## Comparison to other KV Stores

In order to better understand where kvbos stands, here is a comparison to redis and rocksDB

|      | redis  | rocksDB    | kvbos
|------|--------|------------|----------
| type | server | embeddable | server
| storage | RAM | disk | NVRAM |
| max size | RAM Size | disk | NVRAM Size |
| persistent | disk | disk | Object storage |

## Miscellaenous

- Use HTTP range GETs for values that are purged from memory

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

## Data structure

```
struct {
    valuePointer uint64
    valueSize    uint32
    keySize      uint32
    key          []byte
}
```

```
struct {
    value  []byte
}
```

or

```
struct {
    valueSize uint32
    keySize   uint32
    value     []byte
    key       []byte
}
```


## Limitations

- Values larger than the Value Block size cannot be stored
- 
