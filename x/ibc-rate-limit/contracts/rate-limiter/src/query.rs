use cosmwasm_std::{to_binary, Binary, Deps, StdResult};

use crate::state::{Path, RATE_LIMIT_TRACKERS};

pub fn get_quotas(
    deps: Deps,
    channel_id: impl Into<String>,
    denom: impl Into<String>,
) -> StdResult<Binary> {
    let path = Path::new(channel_id, denom);
    to_binary(&RATE_LIMIT_TRACKERS.load(deps.storage, path.into())?)
}
