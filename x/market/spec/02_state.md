<!--
order: 2
-->

# State

## OsmosisPoolDelta

Market module provides swap functionality based on constant product mechanism. Osmo pool have to keep its delta to track the currency demands for swap spread. Luna pool can be retrieved from Osmo pool delta with following equation:

```go
OsmoPool := BasePool + delta
LunaPool := (BasePool * BasePool) / OsmoPool
```

> Note that the all pool holds decimal unit of `usdr` amount, so delta is also `usdr` unit.

- OsmosisPoolDelta: `0x01 -> amino(OsmosisPoolDelta)`

```go
type OsmosisPoolDelta sdk.Dec // the gap between the OsmoPool and the BasePool
```
