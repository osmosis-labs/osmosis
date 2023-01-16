# osmosis-std

[![osmosis-std on crates.io](https://img.shields.io/crates/v/osmosis-std.svg)](https://crates.io/crates/osmosis-std) [![Docs](https://docs.rs/osmosis-std/badge.svg)](https://docs.rs/osmosis-std)

Osmosis's proto-generated types and helpers for interacting with the appchain. Compatible with CosmWasm contract.

## CosmWasm stargate message and stargate query

You can find all types and querier generated from osmosis's protobuf in their respective module in `osmosis_std`. To understand how each module works, please look at the [osmosis documentation](https://docs.osmosis.zone/osmosis-core/modules).

[Full working example contract can be found here.](https://github.com/osmosis-labs/osmosis-rust/tree/main/examples/cosmwasm/contracts/osmosis-stargate)

### Publishing Osmosis' message from CosmWasm Contract

```rust
use cosmwasm_std::{CosmosMsg, Response, Env};
use osmosis_std::types::osmosis::tokenfactory::v1beta1::MsgCreateDenom;

# type ContractError = cosmwasm_std::StdError;
// ..

pub fn try_create_denom(env: Env, subdenom: String) -> Result<Response, ContractError> {
    let sender = env.contract.address.into();

    // construct message and convet them into cosmos message
    // (notice `CosmosMsg` type and `.into()`)
    let msg_create_denom: CosmosMsg = MsgCreateDenom { sender, subdenom }.into();

    Ok(Response::new()
        .add_message(msg_create_denom)
        .add_attribute("method", "try_create_denom"))
}

```

## Querying Osmosis' module

Each module has their own querier that derived from protobuf service definition that can be found [here](https://github.com/osmosis-labs/osmosis/tree/v12.1.0/proto/osmosis).

To avoid non-determinism in stargate queries, only some of them are whitelisted, you can find the list [here](https://github.com/osmosis-labs/osmosis/blob/v12.1.0/wasmbinding/stargate_whitelist.go).

```rust
use cosmwasm_std::{Deps, Env, StdResult};
use osmosis_std::types::osmosis::tokenfactory::v1beta1::{TokenfactoryQuerier, QueryDenomsFromCreatorResponse};

// ..

fn query_creator_denoms(deps: Deps, env: Env) -> StdResult<QueryDenomsFromCreatorResponse> {
    // create `TokenfactoryQuerier`
    let tokenfactory = TokenfactoryQuerier::new(&deps.querier);

    // `TokenfactoryQuerier` has all the fns for querying the module
    let res = tokenfactory.denoms_from_creator(env.contract.address.into())?;

    Ok(QueryDenomsFromCreatorResponse { denoms: res.denoms })
}
```

## Querying Pool

When querying pool related values, eg. `Gamm::pool`, you might find that return type contains `Any`. It's a cosmos' way to implement polymorphism in protobuf.

https://github.com/osmosis-labs/osmosis/blob/f024498f1e8e0d2a1fe259cd9cc4223803fea0cd/proto/osmosis/gamm/v1beta1/query.proto#L82-L84

```proto
message QueryPoolResponse {
  google.protobuf.Any pool = 1 [ (cosmos_proto.accepts_interface) = "PoolI" ];
}
```

This is needed due to osmosis supporting multiple pool types which will be added in the future.

For that matter, `osmosis-std` provides `TryFrom` trait for all possible `Any` used in all query responses in this crate.

That means the following code works:

```rust
use prost::DecodeError;
use cosmwasm_std::{Deps, StdResult, StdError};
use osmosis_std::types::osmosis::gamm::v1beta1::GammQuerier;

fn query_pool(
    deps: &Deps,
    pool_id: u64,
) -> StdResult<osmosis_std::types::osmosis::gamm::v1beta1::Pool> {
    let res = GammQuerier::new(&deps.querier).pool(pool_id)?;
    res.pool
        .ok_or_else(|| StdError::NotFound {
            kind: "pool".to_string(),
        })?
        .try_into() // convert `Any` to `osmosis_std::types::osmosis::gamm::v1beta1::Pool`
        .map_err(|e: DecodeError| StdError::ParseErr {
            target_type: "osmosis_std::types::osmosis::gamm::v1beta1::Pool".to_string(),
            msg: e.to_string(),
        })
}
```

Or if later you want to support multiple pool type

```rust
use prost::{DecodeError, Message};
use cosmwasm_std::{Deps, StdResult, StdError};
use osmosis_std::types::osmosis::gamm::v1beta1::GammQuerier;

enum Pool {
    Balancer(osmosis_std::types::osmosis::gamm::v1beta1::Pool),
    StableSwap(osmosis_std::types::osmosis::gamm::poolmodels::stableswap::v1beta1::Pool),
}

impl TryFrom<osmosis_std::shim::Any> for Pool {
    type Error = StdError;

    fn try_from(value: osmosis_std::shim::Any) -> Result<Self, Self::Error> {
        if let Ok(pool) = osmosis_std::types::osmosis::gamm::v1beta1::Pool::decode(value.value.as_slice()) {
            return Ok(Pool::Balancer(pool));
        }
        if let Ok(pool) = osmosis_std::types::osmosis::gamm::poolmodels::stableswap::v1beta1::Pool::decode(value.value.as_slice()) {
            return Ok(Pool::StableSwap(pool));
        }

        Err(StdError::ParseErr {
            target_type: "Pool".to_string(),
            msg: "Unmatched pool: must be either `Balancer` or `StableSwap`.".to_string(),
        })
    }
}

fn query_pool(
    deps: &Deps,
    pool_id: u64,
) -> StdResult<Pool> {
    let res = GammQuerier::new(&deps.querier).pool(pool_id)?;
    res.pool
        .ok_or_else(|| StdError::NotFound {
            kind: "pool".to_string(),
        })?
        .try_into() // convert `Any` to `Pool`
}
```

When translate to rust, especially with CosmWasm, it can get tricky if we want to also support json (de)serialization. [It could erase type url information from serialized json as for current implementation.](https://github.com/osmosis-labs/osmosis-rust/issues/43).

## Non-CosmWasm Client

(WIP)
