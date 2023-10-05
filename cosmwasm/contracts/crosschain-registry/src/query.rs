use crate::state::{
    CHAIN_PFM_MAP, CHAIN_TO_BECH32_PREFIX_MAP, CHAIN_TO_BECH32_PREFIX_REVERSE_MAP,
    CHAIN_TO_CHAIN_CHANNEL_MAP, CHANNEL_ON_CHAIN_CHAIN_MAP, DENOM_ALIAS_MAP,
    DENOM_ALIAS_REVERSE_MAP,
};

use cosmwasm_std::{Deps, StdError};
use registry::proto::{DenomTrace, QueryDenomTraceRequest};

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
            "Chain {chain_name} to bech32 prefix mapping is disabled"
        )));
    }

    Ok(chain_to_bech32_prefix_map.value)
}

pub fn query_chain_name_from_bech32_prefix(deps: Deps, prefix: String) -> Result<String, StdError> {
    let chains = CHAIN_TO_BECH32_PREFIX_REVERSE_MAP.load(deps.storage, &prefix)?;
    if chains.len() > 1 {
        return Err(StdError::generic_err(format!(
            "Bech32 prefix {prefix} is not unique"
        )));
    }

    match chains.first() {
        Some(chain) => Ok(chain.to_string()),
        None => Err(StdError::generic_err(format!(
            "Bech32 prefix {prefix} is not found"
        ))),
    }
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
            "Channel from {source_chain} to {destination_chain} mapping is disabled"
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
            "Destination chain from channel {on_chain} on source chain {via_channel} mapping is disabled"
        )));
    }

    Ok(chain.value)
}

pub fn query_chain_has_pfm(deps: Deps, chain: String) -> bool {
    let chain = chain.to_lowercase();
    if let Ok(chain_pfm) = CHAIN_PFM_MAP.load(deps.storage, &chain) {
        deps.api.debug(&format!("{chain_pfm:?}"));
        chain_pfm.is_validated()
    } else {
        false
    }
}

pub fn query_denom_path_for_alias(deps: Deps, alias: &str) -> Result<String, StdError> {
    let path = DENOM_ALIAS_REVERSE_MAP
        .load(deps.storage, alias)
        .map_err(|_| StdError::GenericErr {
            msg: format!("alias {alias} cannot be found"),
        })?;

    if !path.enabled {
        return Err(StdError::GenericErr {
            msg: format!("alias {alias} is disabled"),
        });
    }

    Ok(path.value)
}

pub fn query_alias_for_denom_path(deps: Deps, denom_path: &str) -> Result<String, StdError> {
    let alias = DENOM_ALIAS_MAP.load(deps.storage, denom_path)?;

    if !alias.enabled {
        return Err(StdError::generic_err(format!(
            "alias for path {denom_path} is disabled"
        )));
    }

    Ok(alias.value)
}
