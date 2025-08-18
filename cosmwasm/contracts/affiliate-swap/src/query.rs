use cosmwasm_std::{Deps, StdResult};

use crate::msg::ConfigResponse;
use crate::state::CONFIG;

pub fn query_config(deps: Deps) -> StdResult<ConfigResponse> {
    let cfg = CONFIG.load(deps.storage)?;
    Ok(ConfigResponse {
        owner: cfg.owner.into_string(),
        affiliate_addr: cfg.affiliate_addr.into_string(),
        affiliate_bps: cfg.affiliate_bps,
    })
}
