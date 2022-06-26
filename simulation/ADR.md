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

(CFMM k value)

### Make dependent, multi-msg property tests

(Join <> Exit invariants)

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

### Weight management

Weight management should be done globally at the simulator, with a config.

### Observability into simulation run

We should make all logs go into a standard database, have an easy way of seeing the history
and seeing how many message attempts failed

## What exists in the SDK to date

The SDK to date only has tooling for constructing complex state machines, and memoryless checks (dubbed invariants)
Moreover, this is done with fairly painful dev UX's.

## Architecture

We solve building a complex state machine+ 