
# KVBOS

kvbos is a **K**ey **V**alue **B**acked (on) **O**bject **S**torage.

It is a Key Value store relying on storing key/value pairs in NVRAM that is backed by Object Storage.

It is meant for very large datasets. By taking full advantage of emerging NVRAM trends, servers with up to 150 TB of total space can be created today. This will grow over the next one to two years to PB (Petabyte) scale size.

## Persistent Memory

Think of kvbos as creating a persistent memory space at a 64-bit address range scale.

Think of a 64-bit address space whereby the values are stored from te bottom of the address space and the keys are stored from the top of the address space.

Values are grouped into immutable 4 MB blocks. Values are stored one after the other (either aligned or not). In case there is not sufficient space in the block to store a value, the current block will be closed and saved and a new block will be created to store the value.

Keys are also grouped into immutable blocks of size 512 KB.

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
