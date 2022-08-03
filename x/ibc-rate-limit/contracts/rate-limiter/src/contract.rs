#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg};
use crate::state::{Flow, FlowType, FLOW, IBCMODULE, QUOTA};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:rate-limiter";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    IBCMODULE.save(deps.storage, &info.sender)?;

    for (channel, quota) in msg.channel_quotas {
        QUOTA.save(deps.storage, channel.clone(), &quota.into())?;
        FLOW.save(deps.storage, channel, &Flow::new(0_u128, 0_u128))?;
    }

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("ibc_module", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    let ibc_module = IBCMODULE.load(deps.storage)?;
    if info.sender != ibc_module {
        return Err(ContractError::Unauthorized {});
    }
    match msg {
        ExecuteMsg::SendPacket {
            channel_id,
            channel_value,
            funds,
        } => try_transfer(deps, channel_id, channel_value, funds, FlowType::In),
        ExecuteMsg::RecvPacket {
            channel_id,
            channel_value,
            funds,
        } => try_transfer(deps, channel_id, channel_value, funds, FlowType::Out),
        ExecuteMsg::AddChannel {} => todo!(),
        ExecuteMsg::RemoveChannel {} => todo!(),
    }
}

pub fn try_transfer(
    deps: DepsMut,
    channel_id: String,
    channel_value: u128,
    funds: u128,
    direction: FlowType,
) -> Result<Response, ContractError> {
    let quota = QUOTA.load(deps.storage, channel_id.clone())?;
    let max = quota.apply(&channel_value);
    let mut flow = FLOW.load(deps.storage, channel_id.clone())?;
    flow.add_flow(direction, funds);
    if flow.volume() > max {
        return Err(ContractError::RateLimitExceded {
            channel: channel_id.clone(),
            reset: flow.period_end,
        });
    }

    FLOW.update(
        deps.storage,
        channel_id.clone(),
        |_| -> Result<_, ContractError> { Ok(flow) },
    )?;

    Ok(Response::new()
        .add_attribute("method", "try_transfer")
        .add_attribute("channel_id", channel_id)
        .add_attribute("used", flow.volume().to_string())
        .add_attribute("max", max.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: ExecuteMsg) -> StdResult<Binary> {
    todo!()
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{coins, Addr, Attribute};

    const CREATOR_ADDR: &str = "IBC_MODULE";

    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg {
            ibc_module: Addr::unchecked(CREATOR_ADDR),
            channel_quotas: vec![],
        };
        let info = mock_info(CREATOR_ADDR, &coins(1000, "nosmo"));

        // we can just call .unwrap() to assert this was a success
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        // TODO: Check initialization values are correct
    }

    #[test]
    fn permissions() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg {
            ibc_module: Addr::unchecked(CREATOR_ADDR),
            channel_quotas: vec![("channel".to_string(), 10)],
        };
        let info = mock_info(CREATOR_ADDR, &coins(1000, "nosmo"));
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        // beneficiary can release it
        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };

        // This succeeds
        execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let info = mock_info("SomeoneElse", &coins(1000, "nosmo"));

        // beneficiary can release it
        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };
        let err = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap_err();
        assert!(matches!(err, ContractError::Unauthorized { .. }));
    }

    #[test]
    fn use_allowance() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg {
            ibc_module: Addr::unchecked(CREATOR_ADDR),
            channel_quotas: vec![("channel".to_string(), 10)],
        };
        let info = mock_info(CREATOR_ADDR, &coins(1000, "nosmo"));
        let _res = instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        // beneficiary can release it
        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };
        let res = execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        let Attribute { key, value } = &res.attributes[2];
        assert_eq!(key, "used");
        assert_eq!(value, "300");

        let msg = ExecuteMsg::SendPacket {
            channel_id: "channel".to_string(),
            channel_value: 3_000,
            funds: 300,
        };
        let err = execute(deps.as_mut(), mock_env(), info, msg).unwrap_err();
        assert!(matches!(err, ContractError::RateLimitExceded { .. }));
        //assert_eq!(18, value.count);
    }
}
