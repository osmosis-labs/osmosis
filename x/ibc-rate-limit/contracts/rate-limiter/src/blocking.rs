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

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::mock_dependencies;
    use cosmwasm_std::Uint256;

    #[test]
    fn test_in_flow_allowed() {
        let deps = mock_dependencies();
        let packet = Packet::mock(
            "src_channel".to_string(),
            "dest_channel".to_string(),
            "denom1".to_string(),
            Uint256::from(100u128),
        );
        let flow_type = FlowType::In;

        let result = check_restricted_denoms(deps.as_ref(), &packet, &flow_type);
        assert!(result.is_ok());
    }

    #[test]
    fn test_out_flow_unrestricted_denom() {
        let deps = mock_dependencies();
        let packet = Packet::mock(
            "src_channel".to_string(),
            "dest_channel".to_string(),
            "denom2".to_string(),
            Uint256::from(100u128),
        );
        let flow_type = FlowType::Out;

        // denom2 is not in the restricted list
        let result = check_restricted_denoms(deps.as_ref(), &packet, &flow_type);
        assert!(result.is_ok());
    }

    #[test]
    fn test_out_flow_restricted_denom_allowed_channel() {
        let mut deps = mock_dependencies();
        let packet = Packet::mock(
            "src_channel_allowed".to_string(),
            "dest_channel".to_string(),
            "denom1".to_string(),
            Uint256::from(100u128),
        );
        let flow_type = FlowType::Out;

        // Add denom1 to restricted list with allowed channels
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .save(
                deps.as_mut().storage,
                "denom1".to_string(),
                &vec!["src_channel_allowed".to_string()],
            )
            .unwrap();

        let result = check_restricted_denoms(deps.as_ref(), &packet, &flow_type);
        assert!(result.is_ok());
    }

    #[test]
    fn test_out_flow_restricted_denom_blocked_channel() {
        let mut deps = mock_dependencies();
        let packet = Packet::mock(
            "src_channel_blocked".to_string(),
            "dest_channel".to_string(),
            "denom1".to_string(),
            Uint256::from(100u128),
        );
        let flow_type = FlowType::Out;

        // Add denom1 to restricted list with allowed channels
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .save(
                deps.as_mut().storage,
                "denom1".to_string(),
                &vec!["src_channel_allowed".to_string()],
            )
            .unwrap();

        let result = check_restricted_denoms(deps.as_ref(), &packet, &flow_type);
        assert!(result.is_err());

        if let Err(ContractError::ChannelBlocked { denom, channel }) = result {
            assert_eq!(denom, "denom1".to_string());
            assert_eq!(channel, "src_channel_blocked".to_string());
        } else {
            panic!("Expected ChannelBlocked error");
        }
    }

    #[test]
    fn test_out_flow_restricted_denom_empty_channel_list() {
        let mut deps = mock_dependencies();
        let packet = Packet::mock(
            "src_channel_blocked".to_string(),
            "dest_channel".to_string(),
            "denom1".to_string(),
            Uint256::from(100u128),
        );
        let flow_type = FlowType::Out;

        // Add denom1 to restricted list but with an empty allowed channels list
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .save(deps.as_mut().storage, "denom1".to_string(), &vec![])
            .unwrap();

        let result = check_restricted_denoms(deps.as_ref(), &packet, &flow_type);
        assert!(result.is_ok());
    }
}
