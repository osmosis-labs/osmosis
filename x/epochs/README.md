# x/epochs

The epochs module defines on-chain timers, that execute at fixed time intervals.
Other SDK modules can then register logic to be executed at the timer ticks.
We refer to the period in between two timer ticks as an "epoch".

Every timer has a unique identifier.
Every epoch will have a start time, and an end time, where `end time = start time + timer interval`.
On Osmosis mainnet, we only utilize one identifier, with a time interval of `one day`.

