use cosmwasm_std::{Deps, Order, StdResult, Storage};

use crate::{
    packet::{hash_denom, Packet},
    state::{flow::FlowType, storage::ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM},
    ContractError,
};

/// A restriction list that allows no channel. IBC channel ids are always
/// `channel-N`, so this entry can never match a packet's source channel.
/// Written only by the migration when two aliases of one token disagreed, so
/// that the fail-closed outcome lives under a single canonical key that
/// governance can query, replace or unset as one restriction.
pub const NO_CHANNEL_ALLOWED: &str = "no-channel";

/// The key a denom's restriction is stored under: the denom as it exists on
/// this chain. A packet-form denom (transfer/<channel>/<base>) becomes its
/// ibc/HASH; anything else is used as-is. Both spellings of the same token
/// therefore resolve to one key, so an entry can never be shadowed by its alias.
pub fn restriction_key(denom: &str) -> String {
    if denom.starts_with("transfer/") {
        hash_denom(denom)
    } else {
        denom.to_string()
    }
}

/// The channels `denom` is restricted to, or None when it is unrestricted.
///
/// Only the canonical key is consulted. The 0.2.0 migration moved every
/// packet-form entry to its canonical key (see canonicalise_restrictions) and
/// Set and Unset only write canonical keys since, so no other entry can exist.
/// An empty list never restricts; before 0.2.0 one could be stored and meant
/// nothing, and the migration drops them.
pub fn effective_restriction(storage: &dyn Storage, denom: &str) -> StdResult<Option<Vec<String>>> {
    Ok(ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
        .may_load(storage, restriction_key(denom))?
        .filter(|channels| !channels.is_empty()))
}

/// What canonicalise_restrictions did, for the migration event
#[derive(Debug, Default, PartialEq, Eq)]
pub struct Canonicalised {
    /// legacy packet-form entries moved (or merged) under their canonical key
    pub moved: usize,
    /// empty entries removed; they never restricted anything
    pub dropped: usize,
    /// legacy and canonical entries that restricted to disjoint channel sets.
    /// Collapsed into one canonical entry allowing no channel, so the token
    /// stays blocked everywhere (as both entries together already blocked it)
    /// until governance replaces or unsets that single entry.
    pub conflicting: usize,
}

/// Moves every restriction written under a packet-form key to its canonical
/// key and drops empty entries, so that afterwards each token has at most one
/// entry and Set, Unset and the query all act on the whole restriction. When
/// both spellings restrict, the result is their intersection, so neither can
/// relax the other; a disjoint pair collapses to NO_CHANNEL_ALLOWED. Run once
/// by the 0.2.0 migration. The map is written by governance only (one entry on
/// mainnet), so a full pass over it is bounded in practice.
pub fn canonicalise_restrictions(storage: &mut dyn Storage) -> StdResult<Canonicalised> {
    let entries: Vec<(String, Vec<String>)> = ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
        .range(storage, None, None, Order::Ascending)
        .collect::<StdResult<_>>()?;
    let mut result = Canonicalised::default();
    for (key, channels) in entries {
        if channels.is_empty() {
            ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.remove(storage, key);
            result.dropped += 1;
            continue;
        }
        let canonical = restriction_key(&key);
        if canonical == key {
            continue;
        }
        let existing = ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .may_load(storage, canonical.clone())?
            .filter(|existing| !existing.is_empty());
        let merged: Vec<String> = match existing {
            Some(existing) => channels
                .into_iter()
                .filter(|channel| existing.contains(channel))
                .collect(),
            None => channels,
        };
        if merged.is_empty() {
            let blocked = vec![NO_CHANNEL_ALLOWED.to_string()];
            ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.save(storage, canonical, &blocked)?;
            ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.remove(storage, key);
            result.conflicting += 1;
            continue;
        }
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.save(storage, canonical, &merged)?;
        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM.remove(storage, key);
        result.moved += 1;
    }
    Ok(result)
}

pub fn check_restricted_denoms(
    deps: Deps,
    packet: &Packet,
    direction: &FlowType,
) -> Result<(), ContractError> {
    // we are only limiting out-flow. In-flow is always allowed
    if matches!(direction, FlowType::In) {
        return Ok(());
    }

    // On a send the packet carries the packet-form denom; its canonical key
    // is derived from it.
    let Some(channels) = effective_restriction(deps.storage, &packet.data.denom)? else {
        return Ok(());
    };

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
    fn test_out_flow_restriction_keyed_by_ibc_hash_matches_packet_denom() {
        // A restriction stored under the ibc/HASH spelling must apply to the packet,
        // which carries the packet-form denom on sends.
        let mut deps = mock_dependencies();
        let packet_denom = "transfer/channel-6897/usat";
        let packet = Packet::mock(
            "channel-0".to_string(),
            "dest_channel".to_string(),
            packet_denom.to_string(),
            Uint256::from(100u128),
        );

        ACCEPTED_CHANNELS_FOR_RESTRICTED_DENOM
            .save(
                deps.as_mut().storage,
                hash_denom(packet_denom),
                &vec!["channel-6897".to_string()],
            )
            .unwrap();

        let result = check_restricted_denoms(deps.as_ref(), &packet, &FlowType::Out);
        assert!(matches!(result, Err(ContractError::ChannelBlocked { .. })));

        // and the origin channel is still allowed
        let packet = Packet::mock(
            "channel-6897".to_string(),
            "dest_channel".to_string(),
            packet_denom.to_string(),
            Uint256::from(100u128),
        );
        assert!(check_restricted_denoms(deps.as_ref(), &packet, &FlowType::Out).is_ok());
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
