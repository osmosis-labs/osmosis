```html
<!--
order: 1
-->
```

# Concepts

The purpose of the `pool incentives` module is to distribute incentives
to a pool's LPs. This assumes that pool's follow the interface from the
`x/gamm` module

`Pool incentives` module doesn't directly distribute the rewards to the
LPs. When a pool is created, the `pool incentives` module creates a
`gauge` in the `incentives` module for every lock duration that exists.
Also, the `pool incentives` module takes a part of the minted inflation
from the mint module, and automatically distributes it to the various
selected gauges.
