use crate::helpers::*;
use crate::state::{
    CHAIN_TO_BECH32_PREFIX_MAP, CHAIN_TO_CHAIN_CHANNEL_MAP, CHANNEL_ON_CHAIN_CHAIN_MAP,
};

use cosmwasm_std::{Deps, StdError};

pub fn query_denom_trace_from_ibc_denom(
    deps: Deps,
    ibc_denom: String,
) -> Result<DenomTrace, StdError> {
    let res = QueryDenomTraceRequest { hash: ibc_denom }.query(&deps.querier)?;

    match res.denom_trace {
        Some(denom_trace) => Ok(denom_trace),
        None => Err(StdError::generic_err("No denom trace found")),
    }
}

pub fn query_bech32_prefix_from_chain_name(
    deps: Deps,
    chain_name: String,
) -> Result<String, StdError> {
    let chain_to_bech32_prefix_map = CHAIN_TO_BECH32_PREFIX_MAP.load(deps.storage, &chain_name)?;

    if !chain_to_bech32_prefix_map.enabled {
        return Err(StdError::generic_err(format!(
            "Chain {} to bech32 prefix mapping is disabled",
            chain_name
        )));
    }

    Ok(chain_to_bech32_prefix_map.value)
}

pub fn query_channel_from_chain_pair(
    deps: Deps,
    source_chain: String,
    destination_chain: String,
) -> Result<String, StdError> {
    let channel = CHAIN_TO_CHAIN_CHANNEL_MAP.load(
        deps.storage,
        (
            &source_chain.to_lowercase(),
            &destination_chain.to_lowercase(),
        ),
    )?;

    if !channel.enabled {
        return Err(StdError::generic_err(format!(
            "Channel from {} to {} mapping is disabled",
            source_chain, destination_chain
        )));
    }

    Ok(channel.value)
}

pub fn query_chain_from_channel_chain_pair(
    deps: Deps,
    on_chain: String,
    via_channel: String,
) -> Result<String, StdError> {
    let chain = CHANNEL_ON_CHAIN_CHAIN_MAP.load(
        deps.storage,
        (&via_channel.to_lowercase(), &on_chain.to_lowercase()),
    )?;

    if !chain.enabled {
        return Err(StdError::generic_err(format!(
            "Destination chain from channel {} on source chain {} mapping is disabled",
            on_chain, via_channel
        )));
    }

    Ok(chain.value)
}
