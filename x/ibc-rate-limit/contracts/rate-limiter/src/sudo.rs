use cosmwasm_std::{DepsMut, Order, Response, StdResult, Storage, Timestamp, Uint256};

use crate::{
    blocking::check_restricted_denoms,
    packet::Packet,
    state::{
        flow::FlowType,
        path::Path,
        pending_send::{same_windows, split_ids, PendingSend, WindowId},
        rate_limit::RateLimit,
        storage::{MIGRATED_AT, PENDING_SENDS, PENDING_SENDS_BY_EXPIRY, RATE_LIMIT_TRACKERS},
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

    // A send identical to one still pending from an earlier window is refused
    // before anything is counted or queried (see record_pending_send).
    if matches!(direction, FlowType::Out) {
        reject_cross_window_duplicate(deps.storage, now, &packet, &trackers, &any_trackers)?;
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
    let (pending_count, evicted) =
        record_pending_send(deps.storage, now, &packet, &trackers, &any_trackers)?;
    Ok(response
        .add_attribute("recorded", "true")
        .add_attribute("pending_count", pending_count.to_string())
        .add_attribute("evicted_stale", evicted.to_string()))
}

/// The window each tracker will be in once a transfer at `now` is applied:
/// the current one, or a fresh one when the current one has ended, exactly
/// as Flow::apply_transfer resets it (which also advances the generation).
fn upcoming_window_ids(trackers: &[RateLimit], now: Timestamp) -> Vec<(String, Timestamp, u64)> {
    trackers
        .iter()
        .map(|limit| {
            let (period_end, generation) = if limit.flow.is_expired(now) {
                (
                    now.plus_seconds(limit.quota.duration),
                    limit.flow.generation.saturating_add(1),
                )
            } else {
                (limit.flow.period_end, limit.flow.generation)
            };
            (limit.quota.name.clone(), period_end, generation)
        })
        .collect()
}

/// Refuses a send whose content matches a retained record from different
/// windows. Such a send could not share the record (its refund would land in
/// the wrong window) and must not be counted without one (an unrefundable
/// charge would be netted off as receive capacity), so it is rejected before
/// any state changes. It can only happen when an identical packet with the
/// same, still valid timeout is sent again after a window has rolled over,
/// and lasts as long as the earlier record is retained.
fn reject_cross_window_duplicate(
    storage: &dyn Storage,
    now: Timestamp,
    packet: &Packet,
    trackers: &[RateLimit],
    any_trackers: &[RateLimit],
) -> Result<(), ContractError> {
    let key = (packet.source_channel.clone(), packet.send_key());
    let Some(existing) = PENDING_SENDS.may_load(storage, key)? else {
        return Ok(());
    };
    if existing.is_expired(now) {
        return Ok(());
    }
    if same_windows(&existing.channel_ids(), &upcoming_window_ids(trackers, now))
        && same_windows(&existing.any_ids(), &upcoming_window_ids(any_trackers, now))
    {
        return Ok(());
    }
    Err(ContractError::DuplicatePendingSend {
        channel: packet.source_channel.clone(),
        denom: packet.data.denom.clone(),
    })
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

/// The (quota name, window end, generation) of each tracker after a transfer
fn window_ids(trackers: &[RateLimit]) -> Vec<(String, Timestamp, u64)> {
    trackers
        .iter()
        .map(|limit| {
            (
                limit.quota.name.clone(),
                limit.flow.period_end,
                limit.flow.generation,
            )
        })
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
    // The two bounds measure different things: the percentage bound is
    // compared against the net flow, the absolute bound against the gross
    // flow. Each is reported next to the figure it is compared against so a
    // monitor can compute either utilisation without mixing them up.
    let (used_in, used_out) = result.flow.balance();
    let (max_in, max_out) = result.quota.capacity()?;
    // Emitted on every quoted transfer; the cfg gate that was meant to keep
    // them out of production builds is disabled (see the note below this
    // function), so keep this set small.
    let name = &result.quota.name;
    let or_none = |bound: Option<Uint256>| match bound {
        Some(bound) => bound.to_string(),
        None => "none".to_string(),
    };
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
// Every counted send is recorded, because percentage bounds net the two
// directions: a charge that could never be refunded would be netted off as
// receive capacity once the tokens came back. One record exists per content
// key; identical sends counted in the same windows share it and raise its
// count, and each settlement refunds one of them. An identical send that
// would be counted in different windows while the record is retained is
// refused before anything is counted (reject_cross_window_duplicate). The
// record is retained until the packet's timeout has passed by a clock-skew
// margin (PendingSend::retention_end). The chain accepts a send whenever its
// timeout is ahead of the counterparty light client's latest timestamp, which
// trails block time, so a send whose timeout is already behind block time is
// possible and is recorded like any other; its timeout is refunded in its own
// window or not at all.
//
// Cost per record: roughly 200 bytes of key and value plus about 50 bytes per
// quota on the path, plus a small index entry. Records are removed by
// UndoSend, by the eviction below and by PurgeStaleSends. A record whose
// packet was acknowledged successfully is simply left to expire: the chain
// deletes the packet commitment on a success ack, so no later acknowledgement
// or timeout can reach UndoSend for it.
//
// Returns how many identical sends the record now stands for and how many
// expired records were evicted.
fn record_pending_send(
    storage: &mut dyn Storage,
    now: Timestamp,
    packet: &Packet,
    trackers: &[RateLimit],
    any_trackers: &[RateLimit],
) -> Result<(u32, usize), ContractError> {
    let channel = packet.source_channel.clone();
    let send_key = packet.send_key();
    let key = (channel.clone(), send_key.clone());
    let evicted = evict_stale_sends(storage, &channel, now)?;

    let channel_ids = window_ids(trackers);
    let any_ids = window_ids(any_trackers);
    match PENDING_SENDS.may_load(storage, key.clone())? {
        Some(existing) if !existing.is_expired(now) => {
            // reject_cross_window_duplicate ran before the transfer was
            // counted, so the windows match; a mismatch here would be a bug,
            // and the transfer must then fail rather than go unrecorded.
            if !same_windows(&existing.channel_ids(), &channel_ids)
                || !same_windows(&existing.any_ids(), &any_ids)
            {
                return Err(ContractError::DuplicatePendingSend {
                    channel,
                    denom: packet.data.denom.clone(),
                });
            }
            // A count that cannot grow must fail the transfer rather than
            // leave a counted send unrecorded
            let count = existing.count.checked_add(1).ok_or_else(|| {
                ContractError::Overflow("too many identical sends pending".to_string())
            })?;
            let record = PendingSend { count, ..existing };
            PENDING_SENDS.save(storage, key, &record)?;
            return Ok((record.count, evicted));
        }
        Some(expired) => {
            // Retention ended but eviction had not reached it yet
            PENDING_SENDS_BY_EXPIRY.remove(
                storage,
                (
                    channel.clone(),
                    expired.retain_until.nanos(),
                    send_key.clone(),
                ),
            );
        }
        None => {}
    }

    let latest_period_end = channel_ids
        .iter()
        .chain(any_ids.iter())
        .map(|(_, period_end, _)| *period_end)
        .max()
        .unwrap_or_else(|| Timestamp::from_nanos(0));
    let (channel_windows, channel_generations) = split_ids(channel_ids);
    let (any_windows, any_generations) = split_ids(any_ids);
    let record = PendingSend {
        funds: packet.get_funds(),
        channel_windows,
        any_windows,
        retain_until: PendingSend::retention_end(latest_period_end, packet),
        count: 1,
        channel_generations,
        any_generations,
    };
    let retain_until = record.retain_until.nanos();

    PENDING_SENDS.save(storage, key, &record)?;
    PENDING_SENDS_BY_EXPIRY.save(storage, (channel, retain_until, send_key), &())?;
    Ok((1, evicted))
}

/// Removes up to STALE_EVICTIONS_PER_RECORD of the records on `channel` whose
/// retention ended first. Returns how many were removed.
fn evict_stale_sends(
    storage: &mut dyn Storage,
    channel: &str,
    now: Timestamp,
) -> Result<usize, ContractError> {
    // The index is ordered by retention end, so the first entry that has not
    // expired means none after it has either, and the read stops there
    let expired: Vec<(u64, String)> = PENDING_SENDS_BY_EXPIRY
        .sub_prefix(channel.to_string())
        .range(storage, None, None, Order::Ascending)
        .take(STALE_EVICTIONS_PER_RECORD)
        .map(|item| item.map(|(key, _)| key))
        .take_while(|item| !matches!(item, Ok((retain_until, _)) if *retain_until >= now.nanos()))
        .collect::<StdResult<_>>()?;
    let mut evicted = 0;
    for (retain_until, send_key) in expired {
        evicted += remove_stale_record(storage, channel, retain_until, &send_key, now)?;
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
// can only ever restore what the same window charged. One of the identical
// sends the record stands for is settled either way; a packet with no record
// refunds nothing.
pub fn undo_send(deps: DepsMut, packet: Packet) -> Result<Response, ContractError> {
    let (channel_id, denom) = packet.path_data(&FlowType::Out); // Sends have direction out.
    let send_key = packet.send_key();
    let key = (packet.source_channel.clone(), send_key.clone());
    let response = Response::new()
        .add_attribute("method", "undo_send")
        .add_attribute("channel_id", channel_id.clone())
        .add_attribute("denom", denom.clone());

    let path = Path::new(channel_id, denom.clone());
    let any_path = Path::new("any", denom);

    let Some(pending) = PENDING_SENDS.may_load(deps.storage, key.clone())? else {
        // Nothing recorded for this packet: it was counted before the
        // migration to 0.2.0, or its record was already settled or expired.
        // Only the former can still hold a charge in an active window, so
        // that case is refunded the way earlier versions did.
        let refunded = refund_legacy_send(deps.storage, &path, &any_path, packet.get_funds())?;
        return Ok(response
            .add_attribute("refunded", refunded.to_string())
            .add_attribute("pending_count", "0"));
    };

    let (mut trackers, mut any_trackers) = load_trackers(deps.storage, &path)?;

    let refunded_channel = refund_same_window(&mut trackers, &pending.channel_ids(), pending.funds);
    let refunded_any = refund_same_window(&mut any_trackers, &pending.any_ids(), pending.funds);

    if refunded_channel > 0 {
        RATE_LIMIT_TRACKERS.save(deps.storage, path.into(), &trackers)?;
    }
    if refunded_any > 0 {
        RATE_LIMIT_TRACKERS.save(deps.storage, any_path.into(), &any_trackers)?;
    }

    // Settle the record last. The chain today swallows an error from this
    // call without rolling the contract's writes back, so nothing that can
    // fail may run between settling the record and crediting the refund.
    let remaining = pending.count.saturating_sub(1);
    if remaining > 0 {
        let record = PendingSend {
            count: remaining,
            ..pending
        };
        PENDING_SENDS.save(deps.storage, key, &record)?;
    } else {
        PENDING_SENDS.remove(deps.storage, key);
        PENDING_SENDS_BY_EXPIRY.remove(
            deps.storage,
            (
                packet.source_channel.clone(),
                pending.retain_until.nanos(),
                send_key,
            ),
        );
    }

    Ok(response
        .add_attribute("refunded", (refunded_channel + refunded_any).to_string())
        .add_attribute("pending_count", remaining.to_string()))
}

/// Versions before 0.2.0 kept no pending-send records and refunded every
/// failed send to the quota's current window. A send counted before the
/// migration therefore has no record, and its failure is still refunded that
/// way, but only to a window that began before the migration and so may hold
/// the charge. Once every window has rolled over since, nothing is refunded
/// without a record. Returns how many trackers were refunded.
fn refund_legacy_send(
    storage: &mut dyn Storage,
    path: &Path,
    any_path: &Path,
    funds: Uint256,
) -> Result<usize, ContractError> {
    let Some(migrated_at) = MIGRATED_AT.may_load(storage)? else {
        return Ok(0);
    };
    let (mut trackers, mut any_trackers) = load_trackers(storage, path)?;
    let began_before_migration = |limit: &RateLimit| {
        let window_start = limit
            .flow
            .period_end
            .nanos()
            .saturating_sub(limit.quota.duration.saturating_mul(1_000_000_000));
        window_start <= migrated_at.nanos()
    };
    let mut refunded = 0;
    for (limits, path) in [(&mut trackers, path), (&mut any_trackers, any_path)] {
        let mut touched = 0;
        for limit in limits.iter_mut() {
            if began_before_migration(limit) {
                limit.flow.undo_flow(FlowType::Out, funds);
                touched += 1;
            }
        }
        if touched > 0 {
            RATE_LIMIT_TRACKERS.save(storage, path.clone().into(), limits)?;
            refunded += touched;
        }
    }
    Ok(refunded)
}

/// Refunds `funds` to every tracker whose quota is still in the window the
/// send was recorded against. Returns how many trackers were refunded.
fn refund_same_window(trackers: &mut [RateLimit], windows: &[WindowId], funds: Uint256) -> usize {
    let mut refunded = 0;
    for limit in trackers.iter_mut() {
        let same_window = |(name, period_end, generation): &WindowId| {
            *name == limit.quota.name
                && *period_end == limit.flow.period_end
                && match generation {
                    Some(generation) => *generation == limit.flow.generation,
                    None => true,
                }
        };
        if windows.iter().any(same_window) {
            limit.flow.undo_flow(FlowType::Out, funds);
            refunded += 1;
        }
    }
    refunded
}
