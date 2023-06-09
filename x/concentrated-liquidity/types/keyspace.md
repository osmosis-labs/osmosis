# CL keyspace

This document defines the key space format used in CL, and guarantees that are expected to hold.
It was created to more easily reason about key space bugs that we were facing.

If || is seen outside of a `backtick` quotation, it means raw appending.

Also all of this really shows we vastly overpay for complexity in the statemachine, and that collections work in the SDK can really simplify this code.

## 0x01 - Pool Tick storage

If a key exists in state, that begins with `0x01`, it is expected that it is of the form:
`0x01` || `8 byte big endian encoding of pool ID` || `9 byte signed tick encoding`

We are expected to be able to iterate over all ticks in a pool, from most negative to most positive.

## 0x02 - Indexed position ID storage

If a key exists in state, that begins with `0x02`, it is expected that it is of the form:
`0x02/` || `Hexadecimal encoding of an address` || `/` || `string encoding of pool ID` || `/` || `string encoding of position ID`

- We are expected to be able to safely iterate over all positions for an address
- We are expected to be able to safely iterate over all positions for an address, pool_ID pairs.


## 0x08 - Position ID storage

If a key exists in state, that begins with `0x08`, it is expected that it is of the form:
`0x08` || `var-length, base10 string encoding of position ID`


## 0x0D - Position to Lock map

If a key exists in state, that begins with `0x0D`, it is expected that it is of the form:
`0x0D` || `var-length, base10 string encoding of position ID`

## 0x0E - Full range liquidity of every pool

If a key exists in state, that begins with `0x0E`, it is expected that it is of the form:
`0x0E` || `var-length, base10 string encoding of pool ID`


## 0x10 - Lock to Position map

If a key exists in state, that begins with `0x10`, it is expected that it is of the form:
`0x10` || `var-length, base10 string encoding of lock ID`
