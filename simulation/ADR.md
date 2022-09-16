# Simulator ADR

(NOTE these are some notes for myself, that I moved from paper to laptop. I didn't actually write out the relevant parts for anyone else to have a better time in reviewing (the target architecture and why))

We have 5 correlated testing goals, that I'd like the simulator refactor to simultaneously achieve.
On top of this, there are improved dev UX / debugging goals, that any refactor should make easy as well.
We detail what we wish to achieve below, then details of what we do today, and finally a blurb on how we get to where we want to go.

## Testing feature goals

### Build a complex state machine

One goal of the state machine is to construct very complex state machine states.
The only way for us to build state machine states that we don't apriori anticipate,
is to randomly generate many state transitions.

State transitions in the SDK come from:

* txs
* Beginning a block
* Ending a block

So we seek to compose those to achieve this simulation.

### Fuzz test individual SDK Messages

We want to be able to use golang fuzz testing tools to thoroughly test edge case behavior of given message type.

### Run memoryless state consistency tests

On complex states, check that all of state is "consistent" acccording to various checks.
(Known as "invariants" in the current SDK architecture)

### Make memory-ful property tests

Simulator takes as input property tests
A property test can specify a list of actions it wants to run before & after
And this function can maintain its own local state/memory (Not State machine state)

(CFMM k value)

```go
func CheckCfmmK() {

}
```

### Make dependent, multi-msg

Q: Do we consider these property tests or Actions
Its altering state (2 txs), so seems like an Action
But if it fails on the important part (if outcoins > inCoins), we want the simulator to fail with an informative error.

(Join <> Exit invariants)

```go
shares := SimulateJoinPoolMessage(inCoins)
outCoins := ExitPool(shares)
assert outCoins <= inCoins
```

## Dev UX goals

### Simple API

Have minimal overhead for randomizing a message, e.g. Just having:

```golang
func MakeRandomizedJoinPoolMsg(sim simulator.Helper) sdk.Msg {
    pool_id := sim.FuzzLessThan(k.GetNextPoolId(ctx))
    sender := sim.FuzzAddrWithDenoms(k.GetPool(ctx, pool_id).Assets())
    token_in_maxs := sim.FuzzTokensSubset(sender, k.GetPool(pool_id).Assets().Denoms())
    share_out_amount := gamm.EstimateJoinPoolShareOut(ctx, pool_id, token_in_maxs)
    share_out_amount = sim.FuzzEqualInt(share_out_amount)
    
    return &MsgJoinPool{
            sender,
            pool_id,
            token_in_maxs,
            share_out_amount
    }
}
```

### Unscrew up genesis handling

Genesis handling is a tremendous mess right now in the simulator. I've spent many hours in the refactor, and I still have trouble tracking wtf it was doing. The old genesis handling / simState config should be entirely scrapped.

### Weight management

Weight management should be done globally at the simulator, with a config. Not per module.
At most each module can 'optionally' provide a 'weight-hint' from an enum, rather than a number.

### Observability into simulation run

We should make all logs go into a standard database, have an easy way of seeing the history
and seeing how many message attempts failed

### Replay

* Be able to replay old simulation messages (so going off the messages in sequence)
* Be able to run blocks from a real network `osmosisd export-history-for-simulator`

## What exists in the SDK to date

The SDK to date only has tooling for constructing complex state machines, and memoryless checks (dubbed invariants)
Moreover, this is done with fairly painful dev UX's.

## Architecture

We solve building a complex state machine+ 
