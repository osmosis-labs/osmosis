use cosmwasm_std::Deps;

use crate::{
    packet::Packet,
    state::{flow::FlowType, storage::ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM},
    ContractError,
};

pub fn check_restricted_denoms(
    deps: Deps,
    packet: &Packet,
    direction: &FlowType,
) -> Result<(), ContractError> {
    // we are only limiting out-flow. In-flow is always allowed
    if matches!(direction, FlowType::In) {
        return Ok(());
    }

    let channels = ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
        .load(deps.storage, packet.data.denom.to_string())
        .unwrap_or_default();

    // if no channels are blocked, we can return early
    if channels.is_empty() {
        return Ok(());
    }

    // Only channels in the list are allowed. If the source channel is not in the list, we reject the packet
    if !channels.contains(&packet.source_channel) {
        return Err(ContractError::ChannelBlocked {
            denom: packet.data.denom.clone(),
            channel: packet.source_channel.to_string(),
        });
    }

    Ok(())
}
