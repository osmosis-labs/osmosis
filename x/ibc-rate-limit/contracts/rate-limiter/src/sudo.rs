use cosmwasm_std::{DepsMut, Order, Response, StdResult, Storage, Timestamp, Uint256};

use crate::{
    blocking::check_restricted_denoms,
    packet::Packet,
    state::{
        flow::FlowType,
        path::Path,
        pending_send::PendingSend,
        rate_limit::RateLimit,
        storage::{PENDING_SENDS, RATE_LIMIT_TRACKERS},
    },
    ContractError,
};

// This function will process a packet and extract the paths information, funds,
// and channel value from it. This is will have to interact with the chain via grpc queries to properly
// obtain this information.
//
// For backwards compatibility, we're teporarily letting the chain override the
// denom and channel value, but these should go away in favour of the contract
// extracting these from the packet
pub fn process_packet(
    deps: DepsMut,
    packet: Packet,
    direction: FlowType,
    now: Timestamp,
    #[cfg(test)] channel_value_mock: Option<Uint256>,
) -> Result<Response, ContractError> {
    check_restricted_denoms(deps.as_ref(), &packet, &direction)?;

    let (channel_id, denom) = packet.path_data(&direction);
    #[allow(clippy::needless_borrow)]
    let path = &Path::new(channel_id, denom);
    let funds = packet.get_funds();

    // Look for quotas before touching the chain. The channel value is a bank
    // SupplyOf query, and most paths have no quota at all, so this avoids a
    // stargate query on every unquoted transfer and means a failing supply
    // query can only ever affect a path that is actually rate limited.
    let (trackers, any_trackers) = load_trackers(deps.storage, path)?;
    if trackers.is_empty() && any_trackers.is_empty() {
        return Ok(not_configured_response(path));
    }

    #[cfg(test)]
    // When testing we override the channel value with the mock since we can't get it from the chain
    let channel_value = match channel_value_mock {
        Some(channel_value) => channel_value,
        None => packet.channel_value(deps.as_ref(), &direction)?, // This should almost never be used, but left for completeness in case we want to send an empty channel_value from the test
    };

    #[cfg(not(test))]
    let channel_value = packet.channel_value(deps.as_ref(), &direction)?;

    try_transfer(deps.storage, path, channel_value, funds, &direction, now)
}

/// Loads the rate limits configured for a path and for the "any" channel of
/// its denom. Missing entries come back as empty vectors.
fn load_trackers(
    storage: &dyn Storage,
    path: &Path,
) -> Result<(Vec<RateLimit>, Vec<RateLimit>), ContractError> {
    let any_path = Path::new("any", path.denom.clone());
    let any_trackers = RATE_LIMIT_TRACKERS
        .may_load(storage, any_path.into())?
        .unwrap_or_default();
    let trackers = RATE_LIMIT_TRACKERS
        .may_load(storage, path.into())?
        .unwrap_or_default();
    Ok((trackers, any_trackers))
}

/// The response for a packet on a path with no quota configured. Everything is allowed.
fn not_configured_response(path: &Path) -> Response {
    Response::new()
        .add_attribute("method", "try_transfer")
        .add_attribute("channel_id", path.channel.to_string())
        .add_attribute("denom", path.denom.to_string())
        .add_attribute("quota", "none")
}

/// The (quota name, window end) pairs of a set of trackers after a transfer
fn windows_of(trackers: &[RateLimit]) -> Vec<(String, Timestamp)> {
    trackers
        .iter()
        .map(|limit| (limit.quota.name.clone(), limit.flow.period_end))
        .collect()
}

/// This function checks the rate limit and, if successful, stores the updated data about the value
/// that has been transfered through the channel for a specific denom.
/// If the period for a RateLimit has ended, the Flow information is reset.
///
/// The channel_value is the current value of the denom for the the channel as
/// calculated by the caller. This should be the total supply of a denom
pub fn try_transfer(
    storage: &mut dyn Storage,
    path: &Path,
    channel_value: Uint256,
    funds: Uint256,
    direction: &FlowType,
    now: Timestamp,
) -> Result<Response, ContractError> {
    // Sudo call. Only go modules should be allowed to access this

    let any_path = Path::new("any", path.denom.clone());
    let (mut trackers, mut any_trackers) = load_trackers(storage, path)?;

    if trackers.is_empty() && any_trackers.is_empty() {
        // No Quota configured for the current path. Allowing all messages.
        return Ok(not_configured_response(path));
    }

    // If any of the RateLimits fails, allow_transfer() will return
    // ContractError::RateLimitExceded, which we'll propagate out
    let results: Vec<RateLimit> = trackers
        .iter_mut()
        .map(|limit| limit.allow_transfer(path, direction, funds, channel_value, now))
        .collect::<Result<_, ContractError>>()?;

    let any_results: Vec<RateLimit> = any_trackers
        .iter_mut()
        .map(|limit| limit.allow_transfer(path, direction, funds, channel_value, now))
        .collect::<Result<_, ContractError>>()?;

    // Only write back the side that actually has quotas. Writing an empty
    // vector for the other side would leave a permanent entry for every
    // (channel, denom) pair ever seen, one per denom a sender cares to mint.
    if !results.is_empty() {
        RATE_LIMIT_TRACKERS.save(storage, path.into(), &results)?;
    }
    if !any_results.is_empty() {
        RATE_LIMIT_TRACKERS.save(storage, any_path.into(), &any_results)?;
    }

    let response = Response::new()
        .add_attribute("method", "try_transfer")
        .add_attribute("channel_id", path.channel.to_string())
        .add_attribute("denom", path.denom.to_string());

    // Adds the attributes for each path to the response. In prod, the
    // addtribute add_rate_limit_attributes is a noop
    let response = any_results
        .iter()
        .try_fold(response, add_rate_limit_attributes)?;
    results.iter().try_fold(response, add_rate_limit_attributes)
}

// #[cfg(any(feature = "verbose_responses", test))]
fn add_rate_limit_attributes(
    response: Response,
    result: &RateLimit,
) -> Result<Response, ContractError> {
    // The two bounds measure different things: the percentage bound is
    // compared against the net flow, the absolute bound against the gross
    // flow. Each is reported next to the figure it is compared against so a
    // monitor can compute either utilisation without mixing them up.
    let (used_in, used_out) = result.flow.balance();
    let (max_in, max_out) = result.quota.capacity()?;
    let name = &result.quota.name;
    let or_none = |bound: Option<Uint256>| match bound {
        Some(bound) => bound.to_string(),
        None => "none".to_string(),
    };
    // These attributes are only added during testing. That way we avoid
    // calculating these again on prod.
    Ok(response
        .add_attribute(format!("{name}_used_in"), used_in.to_string())
        .add_attribute(format!("{name}_used_out"), used_out.to_string())
        .add_attribute(format!("{name}_max_in"), or_none(max_in.percentage))
        .add_attribute(format!("{name}_max_out"), or_none(max_out.percentage))
        .add_attribute(format!("{name}_gross_in"), result.flow.inflow.to_string())
        .add_attribute(format!("{name}_gross_out"), result.flow.outflow.to_string())
        .add_attribute(format!("{name}_max_absolute_in"), or_none(max_in.absolute))
        .add_attribute(
            format!("{name}_max_absolute_out"),
            or_none(max_out.absolute),
        )
        .add_attribute(
            format!("{name}_period_end"),
            result.flow.period_end.to_string(),
        ))
}

// Leaving the attributes in until we can conditionally compile the contract
// for the go tests in CI: https://github.com/mandrean/cw-optimizoor/issues/19
//
// #[cfg(not(any(feature = "verbose_responses", test)))]
// fn add_rate_limit_attributes(response: Response, _result: &RateLimit) -> Response {
//     response
// }

/// How many of the oldest pending-send records on a channel one RecordSend may
/// evict when they are stale. Sequences grow monotonically per channel, so the
/// front of a channel's range holds its oldest records. Scanning a few per
/// send keeps the map bounded by roughly one quota window of traffic per
/// channel without anyone having to call PurgeStaleSends; a backlog shrinks by
/// this many entries per new send.
const STALE_EVICTIONS_PER_RECORD: usize = 5;

// The chain calls this right after a send passed the rate limit and the
// packet's sequence is known: SendPacket charges the quota first, the packet
// is committed, and then the sequence is associated here. Splitting the two
// means a rejected send is never committed, and a failure to record leaves
// the send charged, which is the conservative outcome. The record remembers
// the window each quota was in so a later failure refunds only while those
// windows are still active.
//
// Cost per record: roughly 150 bytes of key and value plus about 50 bytes per
// quota on the path, so a path with four quotas is around 350 bytes. Records
// are removed by ConfirmSend, UndoSend, the eviction below and PurgeStaleSends.
pub fn record_send(
    deps: DepsMut,
    now: Timestamp,
    packet: Packet,
) -> Result<Response, ContractError> {
    let (channel_id, denom) = packet.path_data(&FlowType::Out); // Sends have direction out.
    let response = Response::new()
        .add_attribute("method", "record_send")
        .add_attribute("channel_id", channel_id.clone())
        .add_attribute("denom", denom.clone());

    // A chain that has not learned the sequence (before the upgrade that
    // passes it) sends 0. Nothing is recorded and a failed send keeps its cost.
    if packet.sequence == 0 {
        return Ok(response.add_attribute("recorded", "false"));
    }

    let path = Path::new(channel_id, denom);
    let (trackers, any_trackers) = load_trackers(deps.storage, &path)?;
    if trackers.is_empty() && any_trackers.is_empty() {
        // Nothing was charged for this send, so there is nothing to refund later
        return Ok(response.add_attribute("recorded", "false"));
    }

    let record = PendingSend {
        funds: packet.get_funds(),
        channel_windows: windows_of(&trackers),
        any_windows: windows_of(&any_trackers),
    };
    let key = (packet.source_channel.clone(), packet.sequence);
    PENDING_SENDS.save(deps.storage, key, &record)?;

    let evicted = evict_stale_sends(deps.storage, &packet.source_channel, now)?;

    Ok(response
        .add_attribute("recorded", "true")
        .add_attribute("evicted_stale", evicted.to_string()))
}

/// Removes up to STALE_EVICTIONS_PER_RECORD of the oldest records on `channel`
/// whose windows have all ended. Returns how many were removed.
fn evict_stale_sends(
    storage: &mut dyn Storage,
    channel: &str,
    now: Timestamp,
) -> Result<usize, ContractError> {
    let oldest: Vec<(u64, bool)> = PENDING_SENDS
        .prefix(channel.to_string())
        .range(storage, None, None, Order::Ascending)
        .take(STALE_EVICTIONS_PER_RECORD)
        .map(|item| item.map(|(sequence, pending)| (sequence, pending.latest_period_end() < now)))
        .collect::<StdResult<_>>()?;
    let mut evicted = 0;
    for (sequence, stale) in oldest {
        if stale {
            PENDING_SENDS.remove(storage, (channel.to_string(), sequence));
            evicted += 1;
        }
    }
    Ok(evicted)
}

// The chain calls this when a sent packet fails (error acknowledgement or
// timeout). The send's cost is refunded to each quota it was counted against,
// but only while that quota is still in the window the send was counted in.
//
// Refunding unconditionally, as versions before 0.2.0 did, let a send made at
// the end of one window be refunded into the next: windows reset lazily on
// the first transfer after they expire, and sending to an invalid receiver is
// enough to trigger the refund, so anyone could double the outflow cap at
// will. Not refunding at all would instead let the same sender consume the
// whole outflow allowance for everyone with sends that cost only gas, since
// the escrowed tokens come straight back. Matching on the recorded window
// closes both.
pub fn undo_send(deps: DepsMut, packet: Packet) -> Result<Response, ContractError> {
    let (channel_id, denom) = packet.path_data(&FlowType::Out); // Sends have direction out.
    let key = (packet.source_channel.clone(), packet.sequence);
    let response = Response::new()
        .add_attribute("method", "undo_send")
        .add_attribute("channel_id", channel_id.clone())
        .add_attribute("denom", denom.clone());

    let Some(pending) = PENDING_SENDS.may_load(deps.storage, key.clone())? else {
        // Nothing recorded: the chain never associated a sequence with this
        // send, or the record was already settled. Nothing to refund.
        return Ok(response.add_attribute("refunded", "0"));
    };
    PENDING_SENDS.remove(deps.storage, key);

    let path = Path::new(channel_id, denom.clone());
    let any_path = Path::new("any", denom);
    let (mut trackers, mut any_trackers) = load_trackers(deps.storage, &path)?;

    let refunded_channel =
        refund_same_window(&mut trackers, &pending.channel_windows, pending.funds);
    let refunded_any = refund_same_window(&mut any_trackers, &pending.any_windows, pending.funds);

    if refunded_channel > 0 {
        RATE_LIMIT_TRACKERS.save(deps.storage, path.into(), &trackers)?;
    }
    if refunded_any > 0 {
        RATE_LIMIT_TRACKERS.save(deps.storage, any_path.into(), &any_trackers)?;
    }

    Ok(response.add_attribute("refunded", (refunded_channel + refunded_any).to_string()))
}

/// Refunds `funds` to every tracker whose quota is still in the window the
/// send was recorded against. Returns how many trackers were refunded.
fn refund_same_window(
    trackers: &mut [RateLimit],
    windows: &[(String, Timestamp)],
    funds: Uint256,
) -> usize {
    let mut refunded = 0;
    for limit in trackers.iter_mut() {
        let same_window = |(name, period_end): &(String, Timestamp)| {
            *name == limit.quota.name && *period_end == limit.flow.period_end
        };
        if windows.iter().any(same_window) {
            limit.flow.undo_flow(FlowType::Out, funds);
            refunded += 1;
        }
    }
    refunded
}

// The chain calls this when a sent packet is acknowledged successfully. The
// send stays counted; only its pending record is settled.
pub fn confirm_send(deps: DepsMut, packet: Packet) -> Result<Response, ContractError> {
    let key = (packet.source_channel.clone(), packet.sequence);
    let settled = PENDING_SENDS.has(deps.storage, key.clone());
    PENDING_SENDS.remove(deps.storage, key);
    Ok(Response::new()
        .add_attribute("method", "confirm_send")
        .add_attribute("channel_id", packet.source_channel)
        .add_attribute("sequence", packet.sequence.to_string())
        .add_attribute("settled", settled.to_string()))
}
