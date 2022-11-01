# IBC Rate Limit

The IBC Rate Limit module is responsible for adding a governance-configurable rate limit to IBC transfers.
This is a safety control, intended to protect assets on osmosis in event of:

* a bug/hack on osmosis
* a bug/hack on the counter-party chain
* a bug/hack in IBC itself

This is done in exchange for a potential (one-way) bridge liveness tradeoff, in periods of high deposits or withdrawals.

The architecture of this package is a minimal go package which implements an [IBC Middleware](https://github.com/cosmos/ibc-go/blob/f57170b1d4dd202a3c6c1c61dcf302b6a9546405/docs/ibc/middleware/develop.md) that wraps the [ICS20 transfer](https://ibc.cosmos.network/main/apps/transfer/overview.html) app, and calls into a cosmwasm contract.
The cosmwasm contract then has all of the actual IBC rate limiting logic.
The Cosmwasm code can be found in the [`contracts`](./contracts/) package, with bytecode findable in the [`bytecode`](./bytecode/) folder. The cosmwasm VM usage allows Osmosis chain governance to choose to change this safety control with no hard forks, via a parameter change proposal, a great mitigation for faster threat adaptavity.

The status of the module is being in a state suitable for some initial governance settable rate limits for high value bridged assets.
Its not in its long term / end state for all channels by any means, but does act as a strong protection we
can instantiate today for high value IBC connections.

## Motivation

The motivation of IBC-rate-limit comes from the empirical observations of blockchain bridge hacks that a rate limit would have massively reduced the stolen amount of assets in:

- [Polynetwork Bridge Hack ($611 million)](https://rekt.news/polynetwork-rekt/)
- [BNB Bridge Hack ($586 million)](https://rekt.news/bnb-bridge-rekt/)
- [Wormhole Bridge Hack ($326 million)](https://rekt.news/wormhole-rekt/)
- [Nomad Bridge Hack ($190 million)](https://rekt.news/nomad-rekt/)
- [Harmony Bridge Hack ($100 million)](https://rekt.news/harmony-rekt/) - (Would require rate limit + monitoring)
- [Dragonberry IBC bug](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702) (can't yet disclose amount at risk, but was saved due to being found first by altruistic Osmosis core developers)

In the presence of a software bug on Osmosis, IBC itself, or on a counterparty chain, we would like to prevent the bridge from being fully depegged.
This stems from the idea that a 30% asset depeg is ~infinitely better than a 100% depeg.
Its _crazy_ that today these complex bridged assets can instantly go to 0 in event of bug.
The goal of a rate limit is to raise an alert that something has potentially gone wrong, allowing validators and developers to have time to analyze, react, and protect larger portions of user funds.

The thesis of this is that, it is worthwile to sacrifice liveness in the case of legitimate demand to send extreme amounts of funds, to prevent the terrible long-tail full fund risks.
Rate limits aren't the end-all of safety controls, they're merely the simplest automated one. More should be explored and added onto IBC!

## Rate limit types

We express rate limits in time-based periods.
This means, we set rate limits for (say) 6-hour, daily, and weekly intervals.
The rate limit for a given time period stores the relevant amount of assets at the start of the rate limit.
Rate limits are then defined on percentage terms of the asset.
The time windows for rate limits are currently _not_ rolling, they have discrete start/end times.

We allow setting separate rate limits for the inflow and outflow of assets.
We do all of our rate limits based on the _net flow_ of assets on a channel pair. This prevents DOS issues, of someone repeatedly sending assets back and forth, to trigger rate limits and break liveness.

We currently envision creating two kinds of rate limits:

* Per denomination rate limits
   - allows safety statements like "Only 30% of Stars on Osmosis can flow out in one day" or "The amount of Atom on Osmosis can at most double per day".
* Per channel rate limits
   - Limit the total inflow and outflow on a given IBC channel, based on "USDC" equivalent, using Osmosis as the price oracle.

We currently only implement per denomination rate limits for non-native assets. We do not yet implement channel based rate limits.

Currently these rate limits automatically "expire" at the end of the quota duration. TODO: Think of better designs here. E.g. can we have a constant number of subsequent quotas start filled? Or perhaps harmonically decreasing amounts of next few quotas pre-filled? Halted until DAO override seems not-great.

## Instantiating rate limits

Today all rate limit quotas must be set manually by governance.
In the future, we should design towards some conservative rate limit to add as a safety-backstop automatically for channels.
Ideas for how this could look:

* One month after a channel has been created, automatically add in some USDC-based rate limit
* One month after governance incentivizes an asset, add on a per-denomination rate limit.

Definitely needs far more ideation and iteration!

## Parameterizing the rate limit

One element is we don't want any rate limit timespan thats too short, e.g. not enough time for humans to react to. So we wouldn't want a 1 hour rate limit, unless we think that if its hit, it could be assessed within an hour.

### Handling rate limit boundaries

We want to be safe against the case where say we have a daily rate limit ending at a given time, and an adversary attempts to attack near the boundary window.
We would not like them to be able to "double extract funds" by timing their extraction near a window boundary.

Admittedly, not a lot of thought has been put into how to deal with this well.
Right now we envision simply handling this by saying if you want a quota of duration D, instead include two quotas of duration D, but offset by `D/2` from each other.

Ideally we can change windows to be more 'rolling' in the future, to avoid this overhead and more cleanly handle the problem. (Perhaps rolling ~1 hour at a time)

### Inflow parameterization

The "Inflow" side of a rate limit is essentially protection against unforeseen bug on a counterparty chain.
This can be quite conservative (e.g. bridged amount doubling in one week). This covers a few cases:

* Counter-party chain B having a token theft attack
   - TODO: description of how this looks
* Counter-party chain B runaway mint
   - TODO: description of how this looks
* IBC theft
   - TODO: description of how this looks

It does get more complex when the counterparty chain is itself a DEX, but this is still much more protection than nothing.

### Outflow parameterization

The "Outflow" side of a rate limit is protection against a bug on Osmosis OR IBC.
This has potential for much more user-frustrating issues, if set too low.
E.g. if theres some event that causes many people to suddenly withdraw many STARS or many USDC.

So this parameterization has to contend with being a tradeoff of withdrawal liveness in high volatility periods vs being a crucial safety rail, in event of on-Osmosis bug.

TODO: Better fill out

### Example suggested parameterization

## Code structure

As mentioned at the beginning of the README, the go code is a relatively minimal ICS 20 wrapper, that dispatches relevant calls to a cosmwasm contract that implements the rate limiting functionality.

### Go Middleware

To achieve this, the middleware  needs to implement  the `porttypes.Middleware` interface and the
`porttypes.ICS4Wrapper` interface. This allows the middleware to send and receive IBC messages by wrapping 
any IBC module, and be used as an ICS4 wrapper by a transfer module (for sending packets or writing acknowledgements).

Of those interfaces, just the following methods have custom logic:

* `ICS4Wrapper.SendPacket` forwards to contract, with intent of tracking of value sent via an ibc channel 
* `Middleware.OnRecvPacket` forwards to contract, with intent of tracking of value received via an ibc channel 
* `Middleware.OnAcknowledgementPacket` forwards to contract, with intent of undoing the tracking of a sent packet if the acknowledgment is not a success
* `OnTimeoutPacket` forwards to contract, with intent of undoing the tracking of a sent packet if the packet times out (is not relayed)

All other methods from those interfaces are passthroughs to the underlying implementations.

#### Parameters

The middleware uses the following parameters:

| Key             | Type   |
|-----------------|--------|
| ContractAddress | string |

1. **ContractAddress** -
   The contract address is the address of an instantiated version of the contract provided under `./contracts/`

### Cosmwasm Contract Concepts

Something to keep in mind with all of the code, is that we have to reason separately about every item in the following matrix:

|     Native Token     |     Non-Native Token     |
|----------------------|--------------------------|
| Send Native Token    | Send Non-Native Token    |
| Receive Native Token | Receive Non-Native Token |
| Timeout Native Send  | Timeout Non-native Send  |

(Error ACK can reuse the same code as timeout)

TODO: Spend more time on sudo messages in the following description. We need to better describe how we map the quota concepts onto the code.
Need to describe how we get the quota beginning balance, and that its different for sends and receives.
Explain intracacies of tracking that a timeout and/or ErrorAck must appear from the same quota, else we ignore its update to the quotas.


The tracking contract uses the following concepts

1. **RateLimit** - tracks the value flow transferred and the quota for a path.
2. **Path** - is a (denom, channel) pair.
3. **Flow** - tracks the value that has moved through a path during the current time window.
4. **Quota** - is the percentage of the denom's total value that can be transferred through the path in a given period of time (duration)

#### Messages

The contract specifies the following messages:

##### Query

* GetQuotas - Returns the quotas for a path

##### Exec

* AddPath - Adds a list of quotas for a path
* RemovePath - Removes a path
* ResetPathQuota - If a rate limit has been reached, the contract's governance address can reset the quota so that transfers are allowed again

##### Sudo

Sudo messages can only be executed by the chain.

* SendPacket - Increments the amount used out of the send quota and checks that the send is allowed. If it isn't, it will return a RateLimitExceeded error
* RecvPacket - Increments the amount used out of the receive quota and checks that the receive is allowed. If it isn't, it will return a RateLimitExceeded error
* UndoSend - If a send has failed, the undo message is used to remove its cost from the send quota

### Integration

The rate limit middleware wraps the `transferIBCModule` and is added as the entry route for IBC transfers.

The module is also provided to the underlying `transferIBCModule` as its `ICS4Wrapper`; previously, this would have 
pointed to a channel, which also implements the `ICS4Wrapper` interface.

This integration can be seen in [osmosis/app/keepers/keepers.go](https://github.com/osmosis-labs/osmosis/blob/main/app/keepers/keepers.go)

## Testing strategy

TODO: Fill this out

## Known Future work

Items that've been highlighted above:

* Making automated rate limits get added for channels, instead of manual configuration only
* Improving parameterization strategies / data analysis
* Adding the USDC based rate limits
* We need better strategies for how rate limits "expire".

Not yet highlighted

* Making monitoring tooling to know when approaching rate limiting and when they're hit
* Making tooling to easily give us summaries we can use, to reason about "bug or not bug" in event of rate limit being hit
* Enabling ways to pre-declare large transfers so as to not hit rate limits.
   * Perhaps you can on-chain declare intent to send these assets with a large delay, that raises monitoring but bypasses rate limits?
   * Maybe contract-based tooling to split up the transfer suffices?
* Strategies to account for high volatility periods without hitting rate limits
   * Can imagine "Hop network" style markets emerging
   * Could imagine tieng it into looking at AMM volatility, or off-chain oracles
      * but these are both things we should be wary of security bugs in.
      * Maybe [constraint based programming with tracking of provenance](https://youtu.be/HB5TrK7A4pI?t=2852) as a solution
* Analyze changing denom-based rate limits, to just overall withdrawal amount for Osmosis