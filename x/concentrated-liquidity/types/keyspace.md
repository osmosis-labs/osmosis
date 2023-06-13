# CL keyspace

This document defines the key space format used in CL, and guarantees that are expected to hold.
It was created to more easily reason about key space bugs that we were facing.

If || is seen outside of a `backtick` quotation, it means raw appending.

We use `string encoding` of an integer to mean the variable length, base10 encoding of an integer into a string.

Also all of this really shows we vastly overpay for complexity in the state machine, and that collections work in the SDK can really simplify this code.

We break this up into parts, the multi-component keys, and the single component keys that don't really have concern.

TODO: How do we train AI tools to do more of this for us, instead if it being manual
- Maybe not worth doing since collections / ORM should delete this code for us.

## multi-component keys

## 0x01 - Pool Tick storage

If a key exists in state, that begins with `0x01`, it is expected that it is of the form:
`0x01` || `8 byte big endian encoding of pool ID` || `9 byte signed tick encoding`

We are expected to be able to iterate over all ticks in a pool, from most negative to most positive.

## 0x02 - Indexed position ID storage

If a key exists in state, that begins with `0x02`, it is expected that it is of the form:
`0x02/` || `Hexadecimal encoding of an address` || `/` || `string encoding of pool ID` || `/` || `string encoding of position ID`

- We are expected to be able to safely iterate over all positions for an address
    - Iterate over `0x02/` || `Hexadecimal encoding of an address` || `/` 
- We are expected to be able to safely iterate over all positions for an address, pool_ID pairs.
    - Iterate over `0x02/` || `Hexadecimal encoding of an address` || `/` || `string encoding of pool ID` || `/` 

Since we encode the address in hexadecimal, and hexadecimal doesn't contain `/`, there is no malleability on the encoding.

However, we are likely better off encoding this address with bech32, which also doesn't contain `/`, but has the benefit of being a more user friendly representation. This is at the expense of being marginally more space inefficient. ("43" bytes vs "40" bytes)

## 0x04 - Incentive records

If a key exists in state, that begins with `0x04`, it is expected that it is of the form:

`0x04|` || `string encoding of pool ID` || `|` || `string encoding of min uptime index` || `|` || `denom` || `|` || `bech32 addr`

- This encoding is safe, because denom cannot contain a `|`, it is restricted to alpha-numeric and `/`.

- We are expected to be able to safely iterate over all positions for a pool ID
    - Iterate over `0x04|` || `string encoding of pool ID` || `|` 
- We are expected to be able to safely iterate over all positions for a pool_ID, uptime index.
    - Iterate over `0x04|` || `string encoding of pool ID` || `|` || `string encoding of min uptime index` || `|` 

## 0x09 - Pool Position ID storage

If a key exists in state, that begins with `0x09`, it is expected that it is of the form:
`0x09` || `big endian encoding of pool ID` || `/` || `big endian encoding of position ID`

It is expected that you can iterate over all position ID's for a given pool.

## 0x0F - Balancer full range map

If a key exists in state, that begins with `0x0F`, it is expected that it is of the form:
`0x0F|` || `str encode cl pool ID` || `|` || `str encode balancer pool ID` || `|` || `str encode uptime index`


## single component keys

## 0x03 - Pool storage

Just stores the pool structs.

`0x03` || `var-length, base10 string encoding of pool ID`

## 0x08 - Position ID storage

If a key exists in state, that begins with `0x08`, it is expected that it is of the form:
`0x08` || `var-length, base10 string encoding of position ID`

## accum - Accumulator storage

This one is a bit complicated. Any key that begins with `accum` belongs to accumulator storage. The accumulator package writes state at the following two key formats:

* `accum/acc/{accumName}`
* `accum/pos/{accumName}||{positionName}`

We really should be prefix separating this state into its own sub-area (or potentially even a different store entirely -- I think its worth doing this prelaunch)

accumName and positionName's are confined to one of the following two:
- Spread rewards
    - accumName = `0x0B/` || `str encode pool ID`
    - positionName = `0x0A/` || `str encode position ID`
- incentive rewards
    - accumName = `0x0C/` || `str encode pool ID` || `/` || `str encode uptime index ID`
    - positionName = `0x08` || `var-length, base10 string encoding of position ID`

## 0x0D - Position to Lock map

If a key exists in state, that begins with `0x0D`, it is expected that it is of the form:
`0x0D` || `var-length, base10 string encoding of position ID`

## 0x0E - Full range liquidity of every pool

If a key exists in state, that begins with `0x0E`, it is expected that it is of the form:
`0x0E` || `var-length, base10 string encoding of pool ID`

## 0x10 - Lock to Position map

If a key exists in state, that begins with `0x10`, it is expected that it is of the form:
`0x10` || `var-length, base10 string encoding of lock ID`
