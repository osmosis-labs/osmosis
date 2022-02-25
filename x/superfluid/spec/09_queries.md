<!--
order: 9
-->

# Queries

## Params

```protobuf
message ParamsRequest {};

message ParamsResponse {
  // params defines the parameters of the module.
  Params params = 1 [ (gogoproto.nullable) = false ];
}

message Params {
  sdk.Dec minimum_risk_factor = 1; // serialized as string
}
```

The params query returns the params for the superfluid module.  This currently contains:
- `MinimumRiskFactor` which is an sdk.Dec that represents the discount to apply to all superfluid staked modules when calcultating their staking power.  For example, if a specific denom has an OSMO equivalent value of 100 OSMO, but the the `MinimumRiskFactor` param is 0.05, then the denom will only get 95 OSMO worth of staking power when staked.

## AssetType

```protobuf
message AssetTypeRequest { 
    string denom = 1;
};

message AssetTypeResponse {
    SuperfluidAssetType asset_type = 1;
};

enum SuperfluidAssetType {
  SuperfluidAssetTypeNative = 0;
  SuperfluidAssetTypeLPShare = 1;
}
```

The AssetType query returns what type of superfluid asset a denom is.  AssetTypes are meant for when
we support more types of assets for superfluid staking than just LP shares.  Each AssetType has a different
algorithm used to get its "Osmo equivalent value".

We represent different types of superfluid assets as different enums.  Currently, only enum `1` is actually used.  Enum value `0` is reserved for the Native staking token for if we deprecate the legacy staking workflow to have native staking also go through the superfluid module. In the future, more enums will be added.

If this query errors, that means that a denom is not allowed to be used for superfluid staking.

## AllAssets

```protobuf
message AllAssetsRequest {};

message AllAssetsResponse {
  repeated SuperfluidAsset assets = 1 [ (gogoproto.nullable) = false ];
};

message SuperfluidAsset {
  string denom = 1;
  SuperfluidAssetType asset_type = 2;
}
```

This parameterless query returns a list of all the superfluid staking compatible assets. The return value includes a list of SuperfluidAssets, which are pairs of `denom` with `SuperfluidAssetType` which was described in the previous section.

This query does not currently support pagination, but may in the future.

## AssetMultiplier

```protobuf
message AssetMultiplierRequest {
    string denom = 1;
};

message AssetMultiplierResponse {
  OsmoEquivalentMultiplierRecord osmo_equivalent_multiplier = 1;
};

message OsmoEquivalentMultiplierRecord {
  int64 epoch_number = 1;
  string denom = 2;
  string multiplier = 3;
}
```

This query allows you to find the multiplier factor on a specific denom. The Osmo-Equivalent-Multiplier Record for epoch N refers to the osmo worth we treat a denom as having, for all of epoch N.  For now, this is the spot price at the last epoch boundary, and this is reset every epoch.  We currently don't store historical multipliers, so the epoch parameter is kind of meaningless for now.

To calculate the staking power of the denom, one needs to multiply the amount of the denom with `OsmoEquivalentMultipler` from this query with the `MinimumRiskFactor` from the Params query endpoint.

`staking_power = amount * OsmoEquivalentMultipler * MinimumRiskFactor`

## ConnectedIntermediaryAccount

```protobuf
message ConnectedIntermediaryAccountRequest {
  uint64 lock_id = 1;
}

message ConnectedIntermediaryAccountResponse {
  SuperfluidIntermediaryAccountInfo account = 1;
}

message SuperfluidIntermediaryAccount {
  string denom = 1;
  string val_addr = 2;
  uint64 gauge_id = 3; // perpetual gauge for rewards distribution
}
```

Every superfluid denom and validator pair has an associated "intermediary account", which does the actual delegation.  This query helps find the superfluid intermediary account for any superfluid position.

That `lock_id` parameter passed in is the underlying lock id for the superfluid, NOT the synthetic lock id.

This query can be used to find the validator a superfluid lock is delegated to.  The `gauge_id` also refers to the perpetual gauge that is used to pay out the superfluid positions associated with this intermediary account.

## AllIntermediaryAccounts

```protobuf
message AllIntermediaryAccountsRequest {
  cosmos.base.query.v1beta1.PageRequest pagination = 1;
};

message AllIntermediaryAccountsResponse {
  repeated SuperfluidIntermediaryAccountInfo accounts = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
};
```

This query returns a list of all superfluid intermediary accounts.  It supports pagination.

## SuperfluidDelegationAmount

```protobuf
message SuperfluidDelegationAmountRequest {
  string delegator_address = 1;
  string validator_address = 2;
  string denom = 3;
}

message SuperfluidDelegationAmountResponse {
  repeated cosmos.base.v1beta1.Coin amount = 1 [];
}
```

This query returns the amount of underlying denom (i.e. lp share) for a triplet of delegator, validator, and denom.


## SuperfluidDelegationAmount

```protobuf
message SuperfluidDelegationAmountRequest {
  string delegator_address = 1;
  string validator_address = 2;
  string denom = 3;
}

message SuperfluidDelegationAmountResponse {
  repeated cosmos.base.v1beta1.Coin amount = 1 [];
}
```

This query returns the amount of underlying denom (i.e. lp share) for a triplet of delegator, validator, and denom.

## SuperfluidDelegationsByDelegator

```protobuf
message SuperfluidDelegationsByDelegatorRequest {
  string delegator_address = 1;
}

message SuperfluidDelegationsByDelegatorResponse {
  repeated SuperfluidDelegationRecord superfluid_delegation_records = 1;
  repeated cosmos.base.v1beta1.Coin total_delegated_coins = 2;
}

message SuperfluidDelegationRecord {
  string delegator_address = 1;
  string validator_address = 2;
  cosmos.base.v1beta1.Coin delegation_amount = 3;
}
```

This query returns a list of all the superfluid delegations of a specific delegator.  The return value includes, the validator delgated to and the delegated coins (both denom and amount).

The return value of the query also includes the `total_delegated_coins` which is the sum of all the delegations of that validator.

This query does require iteration that is linear with the number of delegations a delegator has made, but for now until we support many superfluid denoms, should be relatively bounded.  Once that increases, we will need to support pagination.

## SuperfluidDelegationsByValidatorDenom

```protobuf
message SuperfluidDelegationsByValidatorDenomRequest {
  string validator_address = 1;
  string denom = 2;
}

message SuperfluidDelegationsByValidatorDenomResponse {
  repeated SuperfluidDelegationRecord superfluid_delegation_records = 1;
}
```

This query returns a list of all superfluid delegations that are with a validator / superfluid denom pair.  This query requires a lot of iteration and should be used sparingly.  We will need to add pagination to make this usable.

## EstimateSuperfluidDelegatedAmountByValidatorDenom

```protobuf
message EstimateSuperfluidDelegatedAmountByValidatorDenomRequest {
  string validator_address = 1;
  string denom = 2;
}

message EstimateSuperfluidDelegatedAmountByValidatorDenomResponse {
  repeated cosmos.base.v1beta1.Coin total_delegated_coins = 1;
}
```

This query returns the total amount of delegated coins for a validator / superfluid denom pair.  This query does NOT involve iteration, so should be used instead of the above `SuperfluidDelegationsByValidatorDenom` whenever possible.  It is called an "Estimate" because it can have some slight rounding errors, due to conversions between sdk.Dec and sdk.Int", but for the most part it should be very close to the sum of the results of the previous query.