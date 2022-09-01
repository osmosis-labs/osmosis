use cosmwasm_std::{DepsMut, Response, Timestamp};

use crate::{
    state::{FlowType, Path, RateLimit, RATE_LIMIT_TRACKERS},
    ContractError,
};

/// This function checks the rate limit and, if successful, stores the updated data about the value
/// that has been transfered through the channel for a specific denom.
/// If the period for a RateLimit has ended, the Flow information is reset.
///
/// The channel_value is the current value of the denom for the the channel as
/// calculated by the caller. This should be the total supply of a denom
pub fn try_transfer(
    deps: DepsMut,
    path: &Path,
    channel_value: u128,
    funds: u128,
    direction: FlowType,
    now: Timestamp,
) -> Result<Response, ContractError> {
    // Sudo call. Only go modules should be allowed to access this
    let trackers = RATE_LIMIT_TRACKERS.may_load(deps.storage, path.into())?;

    let configured = match trackers {
        None => false,
        Some(ref x) if x.is_empty() => false,
        _ => true,
    };

    if !configured {
        // No Quota configured for the current path. Allowing all messages.
        return Ok(Response::new()
            .add_attribute("method", "try_transfer")
            .add_attribute("channel_id", path.channel.to_string())
            .add_attribute("denom", path.denom.to_string())
            .add_attribute("quota", "none"));
    }

    let mut rate_limits = trackers.unwrap();

    // If any of the RateLimits fails, allow_transfer() will return
    // ContractError::RateLimitExceded, which we'll propagate out
    let results: Vec<RateLimit> = rate_limits
        .iter_mut()
        .map(|limit| limit.allow_transfer(path, &direction, funds, channel_value, now))
        .collect::<Result<_, ContractError>>()?;

    RATE_LIMIT_TRACKERS.save(deps.storage, path.into(), &results)?;

    let response = Response::new()
        .add_attribute("method", "try_transfer")
        .add_attribute("channel_id", path.channel.to_string())
        .add_attribute("denom", path.denom.to_string());

    // Adds the attributes for each path to the response. In prod, the
    // addtribute add_rate_limit_attributes is a noop
    results.iter().fold(Ok(response), |acc, result| {
        Ok(add_rate_limit_attributes(acc?, result))
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
