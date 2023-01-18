use cosmwasm_std::{Deps, StdResult};

use crate::msg::{GetOwnerResponse, GetRouteResponse};
use crate::state::{ROUTING_TABLE, STATE};

pub fn query_owner(deps: Deps) -> StdResult<GetOwnerResponse> {
    let state = STATE.load(deps.storage)?;

    Ok(GetOwnerResponse {
        owner: state.owner.into_string(),
    })
}

pub fn query_route(
    deps: Deps,
    input_token: &str,
    output_token: &str,
) -> StdResult<GetRouteResponse> {
    let route = ROUTING_TABLE.load(deps.storage, (input_token, output_token))?;

    Ok(GetRouteResponse { pool_route: route })
}
