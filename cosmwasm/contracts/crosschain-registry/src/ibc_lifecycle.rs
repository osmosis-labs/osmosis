use cosmwasm_std::{DepsMut, Response};
use registry::Registry;

use crate::{contract::CONTRACT_CHAIN, state::CHAIN_PFM_MAP, ContractError};

pub fn receive_ack(
    deps: DepsMut,
    source_channel: String,
    _sequence: u64,
    _ack: String,
    success: bool,
) -> Result<Response, ContractError> {
    let registry = Registry::default(deps.as_ref());
    let chain = registry.get_connected_chain(CONTRACT_CHAIN, source_channel.as_str())?;
    let mut chain_pfm = CHAIN_PFM_MAP.load(deps.storage, &chain).map_err(|_| {
        ContractError::ValidationNotFound {
            chain: chain.clone(),
        }
    })?;

    if success {
        chain_pfm.acknowledged = true;
        CHAIN_PFM_MAP.save(deps.storage, &chain, &chain_pfm)?;
    } else {
        CHAIN_PFM_MAP.remove(deps.storage, &chain);
    }

    Ok(Response::default())
}

pub fn receive_timeout(
    deps: DepsMut,
    source_channel: String,
    _sequence: u64,
) -> Result<Response, ContractError> {
    let registry = Registry::default(deps.as_ref());
    let chain = registry.get_connected_chain(CONTRACT_CHAIN, source_channel.as_str())?;
    CHAIN_PFM_MAP.remove(deps.storage, &chain);

    Ok(Response::default())
}
