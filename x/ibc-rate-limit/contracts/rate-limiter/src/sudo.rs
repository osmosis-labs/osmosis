use cosmwasm_std::{DepsMut, Response, Storage, Timestamp, Uint256};

use crate::{
    blocking::check_restricted_denoms,
    packet::Packet,
    state::{flow::FlowType, path::Path, rate_limit::RateLimit, storage::RATE_LIMIT_TRACKERS},
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

    try_transfer(deps, path, channel_value, funds, direction, now)
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

/// This function checks the rate limit and, if successful, stores the updated data about the value
/// that has been transfered through the channel for a specific denom.
/// If the period for a RateLimit has ended, the Flow information is reset.
///
/// The channel_value is the current value of the denom for the the channel as
/// calculated by the caller. This should be the total supply of a denom
pub fn try_transfer(
    deps: DepsMut,
    path: &Path,
    channel_value: Uint256,
    funds: Uint256,
    direction: FlowType,
    now: Timestamp,
) -> Result<Response, ContractError> {
    // Sudo call. Only go modules should be allowed to access this

    let any_path = Path::new("any", path.denom.clone());
    let (mut trackers, mut any_trackers) = load_trackers(deps.storage, path)?;

    if trackers.is_empty() && any_trackers.is_empty() {
        // No Quota configured for the current path. Allowing all messages.
        return Ok(not_configured_response(path));
    }

    // If any of the RateLimits fails, allow_transfer() will return
    // ContractError::RateLimitExceded, which we'll propagate out
    let results: Vec<RateLimit> = trackers
        .iter_mut()
        .map(|limit| limit.allow_transfer(path, &direction, funds, channel_value, now))
        .collect::<Result<_, ContractError>>()?;

    let any_results: Vec<RateLimit> = any_trackers
        .iter_mut()
        .map(|limit| limit.allow_transfer(path, &direction, funds, channel_value, now))
        .collect::<Result<_, ContractError>>()?;

    RATE_LIMIT_TRACKERS.save(deps.storage, path.into(), &results)?;
    RATE_LIMIT_TRACKERS.save(deps.storage, any_path.into(), &any_results)?;

    let response = Response::new()
        .add_attribute("method", "try_transfer")
        .add_attribute("channel_id", path.channel.to_string())
        .add_attribute("denom", path.denom.to_string());

    // Adds the attributes for each path to the response. In prod, the
    // addtribute add_rate_limit_attributes is a noop
    let response: Result<Response, ContractError> =
        any_results.iter().try_fold(response, |acc, result| {
            Ok(add_rate_limit_attributes(acc, result))
        });

    results.iter().try_fold(response?, |acc, result| {
        Ok(add_rate_limit_attributes(acc, result))
    })
}

// #[cfg(any(feature = "verbose_responses", test))]
fn add_rate_limit_attributes(response: Response, result: &RateLimit) -> Response {
    let (used_in, used_out) = result.flow.balance();
    let (max_in, max_out) = result.quota.capacity();
    // These attributes are only added during testing. That way we avoid
    // calculating these again on prod.
    response
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
        )
}

// Leaving the attributes in until we can conditionally compile the contract
// for the go tests in CI: https://github.com/mandrean/cw-optimizoor/issues/19
//
// #[cfg(not(any(feature = "verbose_responses", test)))]
// fn add_rate_limit_attributes(response: Response, _result: &RateLimit) -> Response {
//     response
// }

// This function manually injects an inflow. This is used when reverting a
// packet that failed ack or timed-out.
pub fn undo_send(deps: DepsMut, packet: Packet) -> Result<Response, ContractError> {
    // Sudo call. Only go modules should be allowed to access this
    let (channel_id, denom) = packet.path_data(&FlowType::Out); // Sends have direction out.
    #[allow(clippy::needless_borrow)]
    let path = &Path::new(channel_id, &denom);
    let any_path = Path::new("any", &denom);
    let funds = packet.get_funds();

    let mut any_trackers = RATE_LIMIT_TRACKERS
        .may_load(deps.storage, any_path.clone().into())?
        .unwrap_or_default();
    let mut trackers = RATE_LIMIT_TRACKERS
        .may_load(deps.storage, path.into())?
        .unwrap_or_default();

    let not_configured = trackers.is_empty() && any_trackers.is_empty();

    if not_configured {
        // No Quota configured for the current path. Allowing all messages.
        return Ok(Response::new()
            .add_attribute("method", "try_transfer")
            .add_attribute("channel_id", path.channel.to_string())
            .add_attribute("denom", path.denom.to_string())
            .add_attribute("quota", "none"));
    }

    // We force update the flow to remove a failed send
    let results: Vec<RateLimit> = trackers
        .iter_mut()
        .map(|limit| {
            limit.flow.undo_flow(FlowType::Out, funds);
            limit.to_owned()
        })
        .collect();
    let any_results: Vec<RateLimit> = any_trackers
        .iter_mut()
        .map(|limit| {
            limit.flow.undo_flow(FlowType::Out, funds);
            limit.to_owned()
        })
        .collect();

    RATE_LIMIT_TRACKERS.save(deps.storage, path.into(), &results)?;
    RATE_LIMIT_TRACKERS.save(deps.storage, any_path.into(), &any_results)?;

    Ok(Response::new()
        .add_attribute("method", "undo_send")
        .add_attribute("channel_id", path.channel.to_string())
        .add_attribute("denom", path.denom.to_string())
        .add_attribute("any_channel", (!any_trackers.is_empty()).to_string()))
}
