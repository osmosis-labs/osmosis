<!--
order: 7
-->

# Queries

```protobuf
  // Returns superfluid asset type
  rpc AssetType(AssetTypeRequest) returns (AssetTypeResponse) {}
  
  // Returns all superfluid asset types
  rpc AllAssets(AllAssetsRequest) returns (AllAssetsResponse) {}
  
  // Returns superfluid asset Multiplier
  rpc AssetMultiplier(AssetMultiplierRequest) returns (AssetMultiplierResponse) {}
  
  // Returns all superfluid intermediary account
  rpc AllIntermediaryAccounts(AllIntermediaryAccountsRequest) returns (AllIntermediaryAccountsResponse) {}
  
  // Returns intermediary account connected to a superfluid staked lock by id
  rpc ConnectedIntermediaryAccount(ConnectedIntermediaryAccountRequest) returns (ConnectedIntermediaryAccountResponse) {}

  // Returns the coins superfluid delegated for a delegator, validator, denom
  // triplet
  rpc SuperfluidDelegationAmount(SuperfluidDelegationAmountRequest) returns (SuperfluidDelegationAmountResponse) {}

  // Returns all the superfluid poistions for a specific delegator
  rpc SuperfluidDelegationsByDelegator(SuperfluidDelegationsByDelegatorRequest) returns (SuperfluidDelegationsByDelegatorResponse) {}

  // Returns all the superfluid positions of a specific denom delegated to one
  // validator
  rpc SuperfluidDelegationsByValidatorDenom(SuperfluidDelegationsByValidatorDenomRequest) returns (SuperfluidDelegationsByValidatorDenomResponse) {}

  // Returns the amount of a specific denom delegated to a specific validator
  // This is labeled an estimate, because the way it calculates the amount can
  // lead rounding errors from the true delegated amount
  rpc EstimateSuperfluidDelegatedAmountByValidatorDenom(EstimateSuperfluidDelegatedAmountByValidatorDenomRequest)returns (EstimateSuperfluidDelegatedAmountByValidatorDenomResponse) {}
```

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

We represent different types of superfluid assets as different enums.  Currently, only enum `1` is actually used.  Enum value `0` is reserved for the Native staking token for
if we deprecate the legacy staking workflow to have native staking also go through the superfluid module. In the future, more enums will be added.

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

// The Osmo-Equivalent-Multiplier Record for epoch N refers to the osmo worth we
// treat an LP share as having, for all of epoch N. Eventually this is intended
// to be set as the Time-weighted-average-osmo-backing for the entire duration
// of epoch N-1. (Thereby locking whats in use for epoch N as based on the prior
// epochs rewards) However for now, this is not the TWAP but instead the spot
// price at the boundary.  For different types of assets in the future, it could
// change.