use cosmwasm_std::{Addr, DepsMut, Response, Timestamp};
use crate::ContractError;
use crate::msg::{Channel, QuotaMsg};
use crate::state::{CHANNEL_FLOWS, ChannelFlow, Flow, GOVMODULE, IBCMODULE};

pub fn add_new_channels(
    deps: DepsMut,
    channels: Vec<Channel>,
    now: Timestamp,
) -> Result<(), ContractError> {
    for channel in channels {
        CHANNEL_FLOWS.save(
            deps.storage,
            &channel.name,
            &channel
                .quotas
                .iter()
                .map(|q| ChannelFlow {
                    quota: q.into(),
                    flow: Flow::new(0_u128, 0_u128, now, q.duration),
                })
                .collect(),
        )?
    }
    Ok(())
}

pub fn try_add_channel(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    quotas: Vec<QuotaMsg>,
    now: Timestamp,
) -> Result<Response, ContractError> {
    let ibc_module = IBCMODULE.load(deps.storage)?;
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != ibc_module && sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }
    add_new_channels(
        deps,
        vec![Channel {
            name: channel_id.to_string(),
            quotas,
        }],
        now,
    )?;

    Ok(Response::new()
        .add_attribute("method", "try_add_channel")
        .add_attribute("channel_id", channel_id))
}

pub fn try_remove_channel(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
) -> Result<Response, ContractError> {
    let ibc_module = IBCMODULE.load(deps.storage)?;
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != ibc_module && sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }
    CHANNEL_FLOWS.remove(deps.storage, &channel_id);
    Ok(Response::new()
        .add_attribute("method", "try_remove_channel")
        .add_attribute("channel_id", channel_id))
}

pub fn try_reset_channel_quota(
    deps: DepsMut,
    sender: Addr,
    channel_id: String,
    quota_id: String,
    now: Timestamp,
) -> Result<Response, ContractError> {
    let gov_module = GOVMODULE.load(deps.storage)?;
    if sender != gov_module {
        return Err(ContractError::Unauthorized {});
    }

    CHANNEL_FLOWS.update(
        deps.storage,
        &channel_id.clone(),
        |maybe_flows| match maybe_flows {
            None => Err(ContractError::QuotaNotFound {
                quota_id,
                channel_id: channel_id.clone(),
            }),
            Some(mut flows) => {
                flows.iter_mut().for_each(|channel| {
                    if channel.quota.name == channel_id.as_ref() {
                        channel.flow.expire(now, channel.quota.duration)
                    }
                });
                Ok(flows)
            }
        },
    )?;

    Ok(Response::new()
        .add_attribute("method", "try_reset_channel")
        .add_attribute("channel_id", channel_id))
}
