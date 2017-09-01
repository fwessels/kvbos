
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
values/0x000000000000000000
values/0x000000000004000000
values/0x000000000008000000

keys/0xffffffffffffff8000
keys/0xfffffffffffff00000
keys/0xfffffffffffff08000
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
