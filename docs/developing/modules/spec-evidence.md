# Evidence

::: warning NOTE
Osmosis's evidence module inherits from Cosmos SDK's [`evidence`](https://docs.cosmos.network/master/modules/evidence/) module. This document is a stub, and covers mainly important Osmosis-specific notes about how it is used.
:::

The evidence module allows arbitrary evidence of misbehavior, such as equivocation and counterfactual signing, to be submitted and handled.

Typically, standard evidence handling expects the underlying consensus engine, Tendermint, to automatically submit evidence when it is discovered by allowing clients and foreign chains to submit more complex evidence directly. The evidence module operates differently.

All concrete evidence types must implement the `Evidence` interface contract. First, submitted `Evidence` is routed through the evidence module's `Router`, where it attempts to find a corresponding registered `Handler` for that specific `Evidence` type. Each `Evidence` type must have a `Handler` registered with the evidence module's keeper for it to be successfully routed and executed.

Each corresponding handler must also fulfill the `Handler` interface contract. The `Handler` for a given `Evidence` type can perform any arbitrary state transitions such as slashing, jailing, and tombstoning.

## Concepts

### Evidence

Any concrete type of evidence submitted to the  module must fulfill the following `Evidence` contract. Not all concrete types of evidence will fulfill this contract in the same way, and some data might be entirely irrelevant to certain types of evidence. An additional `ValidatorEvidence`, which extends `Evidence`, has also been created to define a contract for evidence against malicious validators.

```
// Evidence defines the contract which concrete evidence types of misbehavior
// must implement.
type Evidence interface {
	proto.Message

	Route() string
	Type() string
	String() string
	Hash() tmbytes.HexBytes
	ValidateBasic() error

	// Height at which the infraction occurred
	GetHeight() int64
}

// ValidatorEvidence extends Evidence interface to define contract
// for evidence against malicious validators
type ValidatorEvidence interface {
	Evidence

	// The consensus address of the malicious validator at time of infraction
	GetConsensusAddress() sdk.ConsAddress

	// The total power of the malicious validator at time of infraction
	GetValidatorPower() int64

	// The total validator set power at time of infraction
	GetTotalPower() int64
}
```

### Registration and handling

First, the evidence module must know about all the types of evidence it is expected to handle. Register the `Route` method in the `Evidence` contract with a `Router` as defined below. The `Router` accepts `Evidence` and attempts to find the corresponding `Handler` for the `Evidence` via the `Route` method.

```
type Router interface {
  AddRoute(r string, h Handler) Router
  HasRoute(r string) bool
  GetRoute(path string) Handler
  Seal()
  Sealed() bool
}
```

As defined below, the `Handler` is responsible for executing the entirety of the business logic for handling `Evidence`. Doing so typically includes validating the evidence, both stateless checks via `ValidateBasic` and stateful checks via any keepers provided to the `Handler`. Additionally, the `Handler` may also perform capabilities, such as slashing and jailing a validator. All `Evidence` handled by the `Handler` must be persisted.

```
// Handler defines an agnostic Evidence handler. The handler is responsible
// for executing all corresponding business logic necessary for verifying the
// evidence as valid. In addition, the Handler may execute any necessary
// slashing and potential jailing.
type Handler func(sdk.Context, Evidence) error
```

### State

The evidence module only stores valid submitted `Evidence` in state. The evidence state is also stored and exported in the evidence module's `GenesisState`.

```
// GenesisState defines the evidence module's genesis state.
message GenesisState {
  // evidence defines all the evidence at genesis.
  repeated google.protobuf.Any evidence = 1;
}
```

## Messages

### MsgSubmitEvidence

Evidence is submitted through a `MsgSubmitEvidence` message:

```
// MsgSubmitEvidence represents a message that supports submitting arbitrary
// Evidence of misbehavior such as equivocation or counterfactual signing.
message MsgSubmitEvidence {
  string              submitter = 1;
  google.protobuf.Any evidence  = 2;
}
```

The `Evidence` of a `MsgSubmitEvidence` message must have a corresponding `Handler` registered with the evidence module's `Router` to be processed and routed correctly.

Given the `Evidence` is registered with a corresponding `Handler`, it is processed as follows:

```
func SubmitEvidence(ctx Context, evidence Evidence) error {
  if _, ok := GetEvidence(ctx, evidence.Hash()); ok {
    return sdkerrors.Wrap(types.ErrEvidenceExists, evidence.Hash().String())
  }
  if !router.HasRoute(evidence.Route()) {
    return sdkerrors.Wrap(types.ErrNoEvidenceHandlerExists, evidence.Route())
  }

  handler := router.GetRoute(evidence.Route())
  if err := handler(ctx, evidence); err != nil {
    return sdkerrors.Wrap(types.ErrInvalidEvidence, err.Error())
  }

  ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitEvidence,
			sdk.NewAttribute(types.AttributeKeyEvidenceHash, evidence.Hash().String()),
		),
	)

  SetEvidence(ctx, evidence)
  return nil
}
```

Valid submitted `Evidence` of the same type must not already exist. The `Evidence` is routed to the `Handler` and executed. If no error in handling the Evidence occurs, an event is emitted, and it is persisted to state.

### Events

The evidence module emits the following handler events:

#### MsgSubmitEvidence

| Type            | Attribute Key | Attribute Value |
| --------------- | ------------- | --------------- |
| submit_evidence | evidence_hash | {evidenceHash}  |
| message         | module        | evidence        |
| message         | sender        | {senderAddress} |
| message         | action        | submit_evidence |

### BeginBlock

#### Evidence handling

Tendermint blocks can include
[Evidence](https://github.com/tendermint/tendermint/blob/master/docs/spec/blockchain/blockchain.md#evidence) that indicates whether a validator acted maliciously. The relevant information is forwarded to the application as ABCI Evidence in `abci.RequestBeginBlock` so that the validator can be punished accordingly.

#### Equivocation

Currently, the SDK handles two types of evidence inside the ABCI `BeginBlock`:

- `DuplicateVoteEvidence`,
- `LightClientAttackEvidence`.

The evidence module handles these two evidence types the same way. First, the SDK converts the Tendermint concrete evidence type to a SDK `Evidence` interface by using `Equivocation` as the concrete type.

```proto
// Equivocation implements the Evidence interface.
message Equivocation {
  int64                     height            = 1;
  google.protobuf.Timestamp time              = 2;
  int64                     power             = 3;
  string                    consensus_address = 4;
}
```

For an `Equivocation` submitted in `block` to be valid, it must meet the following requirement:

`Evidence.Timestamp >= block.Timestamp - MaxEvidenceAge`

where:

- `Evidence.Timestamp` is the timestamp in the block at height `Evidence.Height`.
- `block.Timestamp` is the current block timestamp.

If valid `Equivocation` evidence is included in a block, the validator's stake is
reduced by `SlashFractionDoubleSign`, as defined by the [slashing module](spec-slashing.md). The reduction is implemented at the point when the infraction occurred instead of when the evidence was discovered.
The stake that contributed to the infraction is slashed, even if it has been redelegated or started unbonding.

Additionally, the validator is permanently jailed and tombstoned so that the validator cannot re-enter the validator set again.

::: details `Equivocation` evidence handling code

```go
func (k Keeper) HandleEquivocationEvidence(ctx sdk.Context, evidence *types.Equivocation) {
	logger := k.Logger(ctx)
	consAddr := evidence.GetConsensusAddress()

	if _, err := k.slashingKeeper.GetPubkey(ctx, consAddr.Bytes()); err != nil {
		// Ignore evidence that cannot be handled.
		//
		// NOTE: We used to panic with:
		// `panic(fmt.Sprintf("Validator consensus-address %v not found", consAddr))`,
		// but this couples the expectations of the app to both Tendermint and
		// the simulator.  Both are expected to provide the full range of
		// allowable but none of the disallowed evidence types.  Instead of
		// getting this coordination right, it is easier to relax the
		// constraints and ignore evidence that cannot be handled.
		return
	}

	// calculate the age of the evidence
	infractionHeight := evidence.GetHeight()
	infractionTime := evidence.GetTime()
	ageDuration := ctx.BlockHeader().Time.Sub(infractionTime)
	ageBlocks := ctx.BlockHeader().Height - infractionHeight

	// Reject evidence if the double-sign is too old. Evidence is considered stale
	// if the difference in time and number of blocks is greater than the allowed
	// parameters defined.
	cp := ctx.ConsensusParams()
	if cp != nil && cp.Evidence != nil {
		if ageDuration > cp.Evidence.MaxAgeDuration && ageBlocks > cp.Evidence.MaxAgeNumBlocks {
			logger.Info(
				"ignored equivocation; evidence too old",
				"validator", consAddr,
				"infraction_height", infractionHeight,
				"max_age_num_blocks", cp.Evidence.MaxAgeNumBlocks,
				"infraction_time", infractionTime,
				"max_age_duration", cp.Evidence.MaxAgeDuration,
			)
			return
		}
	}

	validator := k.stakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if validator == nil || validator.IsUnbonded() {
		// Defensive: Simulation doesn't take unbonding periods into account, and
		// Tendermint might break this assumption at some point.
		return
	}

	if ok := k.slashingKeeper.HasValidatorSigningInfo(ctx, consAddr); !ok {
		panic(fmt.Sprintf("expected signing info for validator %s but not found", consAddr))
	}

	// ignore if the validator is already tombstoned
	if k.slashingKeeper.IsTombstoned(ctx, consAddr) {
		logger.Info(
			"ignored equivocation; validator already tombstoned",
			"validator", consAddr,
			"infraction_height", infractionHeight,
			"infraction_time", infractionTime,
		)
		return
	}

	logger.Info(
		"confirmed equivocation",
		"validator", consAddr,
		"infraction_height", infractionHeight,
		"infraction_time", infractionTime,
	)

	// We need to retrieve the stake distribution which signed the block, so we
	// subtract ValidatorUpdateDelay from the evidence height.
	// Note, that this *can* result in a negative "distributionHeight", up to
	// -ValidatorUpdateDelay, i.e. at the end of the
	// pre-genesis block (none) = at the beginning of the genesis block.
	// That's fine since this is just used to filter unbonding delegations & redelegations.
	distributionHeight := infractionHeight - sdk.ValidatorUpdateDelay

	// Slash validator. The `power` is the int64 power of the validator as provided
	// to/by Tendermint. This value is validator.Tokens as sent to Tendermint via
	// ABCI, and now received as evidence. The fraction is passed in to separately
	// to slash unbonding and rebonding delegations.
	k.slashingKeeper.Slash(
		ctx,
		consAddr,
		k.slashingKeeper.SlashFractionDoubleSign(ctx),
		evidence.GetValidatorPower(), distributionHeight,
	)

	// Jail the validator if not already jailed. This will begin unbonding the
	// validator if not already unbonding (tombstoned).
	if !validator.IsJailed() {
		k.slashingKeeper.Jail(ctx, consAddr)
	}

	k.slashingKeeper.JailUntil(ctx, consAddr, types.DoubleSignJailEndTime)
	k.slashingKeeper.Tombstone(ctx, consAddr)
}
```
:::

The slashing, jailing, and tombstoning calls are delegated through the slashing module, which emits informative events and finally delegates calls to the staking module. For more information about slashing and jailing, see [transitions](spec-staking.md#transitions).
