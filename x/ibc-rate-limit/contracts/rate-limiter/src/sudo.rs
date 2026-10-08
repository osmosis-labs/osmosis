use cosmwasm_std::{DepsMut, Order, Response, StdResult, Storage, Timestamp, Uint256};

use crate::{
    blocking::check_restricted_denoms,
    packet::Packet,
    state::{
        flow::FlowType,
        path::Path,
        pending_send::PendingSend,
        rate_limit::RateLimit,
        storage::{PENDING_SENDS, PENDING_SENDS_BY_EXPIRY, RATE_LIMIT_TRACKERS},
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

    let (response, trackers, any_trackers) = try_transfer(
        deps.storage,
        path,
        trackers,
        any_trackers,
        channel_value,
        funds,
        &direction,
        now,
    )?;

    if !matches!(direction, FlowType::Out) {
        return Ok(response);
    }

    // A send that was counted is remembered under its packet content, so that
    // the acknowledgement or timeout the chain reports later can settle it
    // (see undo_send). Receives need no record: nothing is undone for them.
    let (recorded, evicted) =
        record_pending_send(deps.storage, now, &packet, &trackers, &any_trackers)?;
    Ok(response
        .add_attribute("recorded", recorded.to_string())
        .add_attribute("evicted_stale", evicted.to_string()))
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
/// `trackers` and `any_trackers` are the quotas already loaded by the caller
/// for the path and for the "any" channel of its denom; at least one is
/// non-empty. The channel_value is the current value of the denom for the
/// channel as calculated by the caller. This should be the total supply of a denom.
///
/// Returns the response together with the trackers as they were saved, so the
/// caller can record the windows the transfer was counted in.
#[allow(clippy::too_many_arguments)]
pub fn try_transfer(
    storage: &mut dyn Storage,
    path: &Path,
    mut trackers: Vec<RateLimit>,
    mut any_trackers: Vec<RateLimit>,
    channel_value: Uint256,
    funds: Uint256,
    direction: &FlowType,
    now: Timestamp,
) -> Result<(Response, Vec<RateLimit>, Vec<RateLimit>), ContractError> {
    // Sudo call. Only go modules should be allowed to access this

    let any_path = Path::new("any", path.denom.clone());

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

    // Adds the per-quota usage attributes for each path to the response
    let response = any_results
        .iter()
        .try_fold(response, add_rate_limit_attributes)?;
    let response = results
        .iter()
        .try_fold(response, add_rate_limit_attributes)?;
    Ok((response, results, any_results))
}

// #[cfg(any(feature = "verbose_responses", test))]
fn add_rate_limit_attributes(
    response: Response,
    result: &RateLimit,
) -> Result<Response, ContractError> {
    let (used_in, used_out) = result.flow.balance();
    let (max_in, max_out) = result.quota.capacity()?;
    // Emitted on every quoted transfer; the cfg gate that was meant to keep
    // them out of production builds is disabled (see the note below this
    // function), so keep this set small.
    Ok(response
        .add_attribute(
            format!("{}_used_in", result.quota.name),
            used_in.to_string(),
        )
        .add_attribute(
            format!("{}_used_out", result.quota.name),
            used_out.to_string(),
        )
        .add_attribute(format!("{}_max_in", result.quota.name), max_in.to_string())
        .add_attribute(
            format!("{}_max_out", result.quota.name),
            max_out.to_string(),
        )
        .add_attribute(
            format!("{}_period_end", result.quota.name),
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

/// How many expired pending-send records on a channel one counted send may
/// evict. The index orders a channel's records by the end of their retention,
/// so the front of the range is what can go first. Clearing a few per send
/// keeps the map bounded by roughly one retention period of traffic per
/// channel without anyone having to call PurgeStaleSends; a backlog shrinks
/// by this many entries per new send.
const STALE_EVICTIONS_PER_RECORD: usize = 5;

// Remembers a send that was just counted against one or more quotas, keyed by
// the packet content the chain hands back unchanged on an acknowledgement or
// timeout (see Packet::send_key). The record holds the window each quota was
// in, so the later settlement can tell whether those windows are still the
// active ones.
//
// Only one record exists per content key. A second identical send while a
// record is pending is counted like any other send but not recorded, so it
// cannot be refunded; an acknowledgement for that content always settles the
// oldest send. The record is retained until an identical packet could no
// longer be sent (PendingSend::retention_end), so a later identical send can
// never be recorded while an earlier one might still be acknowledged.
//
// Cost per record: roughly 200 bytes of key and value plus about 50 bytes per
// quota on the path, plus a small index entry. Records are removed by
// UndoSend, by the eviction below and by PurgeStaleSends. A record whose
// packet was acknowledged successfully is simply left to expire: the chain
// deletes the packet commitment on a success ack, so no later acknowledgement
// or timeout can reach UndoSend for it.
//
// Returns whether the send was recorded and how many expired records were
// evicted.
fn record_pending_send(
    storage: &mut dyn Storage,
    now: Timestamp,
    packet: &Packet,
    trackers: &[RateLimit],
    any_trackers: &[RateLimit],
) -> Result<(bool, usize), ContractError> {
    let channel = packet.source_channel.clone();
    let send_key = packet.send_key();
    let key = (channel.clone(), send_key.clone());
    if PENDING_SENDS.has(storage, key.clone()) {
        let evicted = evict_stale_sends(storage, &channel, now)?;
        return Ok((false, evicted));
    }

    let channel_windows = windows_of(trackers);
    let any_windows = windows_of(any_trackers);
    let latest_period_end = channel_windows
        .iter()
        .chain(any_windows.iter())
        .map(|(_, period_end)| *period_end)
        .max()
        .unwrap_or_else(|| Timestamp::from_nanos(0));
    let record = PendingSend {
        funds: packet.get_funds(),
        channel_windows,
        any_windows,
        retain_until: PendingSend::retention_end(latest_period_end, packet),
    };
    let retain_until = record.retain_until.nanos();

    PENDING_SENDS.save(storage, key, &record)?;
    PENDING_SENDS_BY_EXPIRY.save(storage, (channel.clone(), retain_until, send_key), &())?;

    let evicted = evict_stale_sends(storage, &channel, now)?;
    Ok((true, evicted))
}

/// Removes up to STALE_EVICTIONS_PER_RECORD of the records on `channel` whose
/// retention ended first. Returns how many were removed.
fn evict_stale_sends(
    storage: &mut dyn Storage,
    channel: &str,
    now: Timestamp,
) -> Result<usize, ContractError> {
    let oldest: Vec<(u64, String)> = PENDING_SENDS_BY_EXPIRY
        .sub_prefix(channel.to_string())
        .range(storage, None, None, Order::Ascending)
        .take(STALE_EVICTIONS_PER_RECORD)
        .map(|item| item.map(|(key, _)| key))
        .collect::<StdResult<_>>()?;
    let mut evicted = 0;
    for (retain_until, send_key) in oldest {
        if retain_until < now.nanos() {
            evicted += remove_stale_record(storage, channel, retain_until, &send_key, now)?;
        }
    }
    Ok(evicted)
}

/// Drops the record stored under (`channel`, `send_key`) if its retention has
/// ended, together with the index entry for `retain_until`. An index entry
/// without a record, or one whose record has a different retention (both
/// impossible unless storage was edited by hand), is dropped on its own.
/// Returns how many records were removed.
pub(crate) fn remove_stale_record(
    storage: &mut dyn Storage,
    channel: &str,
    retain_until: u64,
    send_key: &str,
    now: Timestamp,
) -> Result<usize, ContractError> {
    let key = (channel.to_string(), send_key.to_string());
    let removed = match PENDING_SENDS.may_load(storage, key.clone())? {
        Some(record) if record.retain_until.nanos() == retain_until && record.is_expired(now) => {
            PENDING_SENDS.remove(storage, key);
            1
        }
        _ => 0,
    };
    PENDING_SENDS_BY_EXPIRY.remove(
        storage,
        (channel.to_string(), retain_until, send_key.to_string()),
    );
    Ok(removed)
}

// The chain calls this when a sent packet fails (error acknowledgement or
// timeout). The send's cost is refunded to each quota it was counted against,
// but only while that quota is still in the window the send was counted in.
// A window that did not count the send is never credited for it, so a refund
// can only ever restore what the same window charged. The record is settled
// either way; a packet with no record refunds nothing.
pub fn undo_send(deps: DepsMut, packet: Packet) -> Result<Response, ContractError> {
    let (channel_id, denom) = packet.path_data(&FlowType::Out); // Sends have direction out.
    let send_key = packet.send_key();
    let key = (packet.source_channel.clone(), send_key.clone());
    let response = Response::new()
        .add_attribute("method", "undo_send")
        .add_attribute("channel_id", channel_id.clone())
        .add_attribute("denom", denom.clone());

    let Some(pending) = PENDING_SENDS.may_load(deps.storage, key.clone())? else {
        // Nothing recorded for this packet: it was never counted, it was a
        // second identical send, or its record was already settled or
        // expired. Nothing to refund.
        return Ok(response.add_attribute("refunded", "0"));
    };
    PENDING_SENDS.remove(deps.storage, key);
    PENDING_SENDS_BY_EXPIRY.remove(
        deps.storage,
        (
            packet.source_channel.clone(),
            pending.retain_until.nanos(),
            send_key,
        ),
    );

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
