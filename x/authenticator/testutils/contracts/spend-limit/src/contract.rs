#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    to_json_binary, Addr, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdError, StdResult,
};
use osmosis_std::types::osmosis::poolmanager::v1beta1::SwapAmountInRoute;

use crate::authenticate::sudo_authenticate;
use crate::confirm_execution::sudo_confirm_execution;
use crate::msg::{InstantiateMsg, QueryMsg, SpendLimitDataResponse, SudoMsg};
use crate::state::{Denom, Path, TrackedDenom, SPEND_LIMITS, TRACKED_DENOMS};
use crate::track::sudo_track;
use crate::ContractError;

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, StdError> {
    // Create mock data for Denom and Path
    let osmo_denom: Denom = "uosmo".to_string();
    let osmo_usdc_path: Path = vec![SwapAmountInRoute {
        pool_id: 1,
        token_out_denom: "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
            .to_string(),
    }];

    let atom_denom: Denom =
        "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2".to_string();
    let atom_usdc_path: Path = vec![SwapAmountInRoute {
        pool_id: 2,
        token_out_denom: "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
            .to_string(),
    }];

    // Create a TrackedDenom instance using the InstantiateMsg data
    let tracked_denom = TrackedDenom {
        denom: osmo_denom.clone(),
        path: osmo_usdc_path.clone(),
    };
    let atom_tracked_denom = TrackedDenom {
        denom: atom_denom.clone(),
        path: atom_usdc_path.clone(),
    };

    // Store the TrackedDenom in the map
    TRACKED_DENOMS.save(deps.storage, osmo_denom, &tracked_denom)?;
    TRACKED_DENOMS.save(deps.storage, atom_denom, &atom_tracked_denom)?;
    Ok(Response::new())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, env: Env, msg: SudoMsg) -> Result<Response, ContractError> {
    match msg {
        SudoMsg::Authenticate(auth_request) => sudo_authenticate(deps, env, auth_request),
        SudoMsg::Track(track_request) => sudo_track(deps, env, track_request),
        SudoMsg::ConfirmExecution(confirm_execution_request) => {
            sudo_confirm_execution(deps, env, confirm_execution_request)
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetSpendLimitData { account } => {
            to_json_binary(&query_spend_limit(deps, account)?)
        }
    }
}

pub fn query_spend_limit(deps: Deps, account: Addr) -> StdResult<SpendLimitDataResponse> {
    let spend_limit_data = SPEND_LIMITS.load(deps.storage, account.to_string())?;
    return Ok(SpendLimitDataResponse { spend_limit_data });
}
