#[cfg(not(feature = "library"))]
use cosmwasm_std::{Coin, Deps};
use cosmwasm_std::{Decimal, StdResult, Timestamp, Uint128};

use osmosis_std::shim::Timestamp as OsmosisTimestamp;
use osmosis_std::types::osmosis::poolmanager::v1beta1::SwapAmountInRoute;
use osmosis_std::types::osmosis::twap::v1beta1::TwapQuerier;

use crate::ContractError;

pub fn calculate_price_from_route(
    deps: Deps,
    input_token: Coin,
    now: Timestamp,
    window: Option<u64>,
    percentage_impact: Decimal,
    route: Vec<SwapAmountInRoute>,
) -> Result<Coin, ContractError> {
    if route.is_empty() {
        return Err(ContractError::InvalidPoolRoute {
            reason: format!("Route must not be empty"),
        });
    }

    let output_denom = route
        .last()
        .ok_or(ContractError::InvalidPoolRoute {
            reason: "route must have at least one element".to_string(),
        })?
        .token_out_denom
        .clone();

    let percentage = percentage_impact / Uint128::new(100);

    let mut twap_price: Decimal = Decimal::one();

    // When swapping from input to output, we need to quote the price in the input token
    // For example when seling osmo to buy atom:
    //  price of <out> is X<in> (i.e.: price of atom is Xosmo)
    let mut sell_denom = input_token.denom;

    // if duration is not provided, default to 1h
    let start_time = now.minus_seconds(window.unwrap_or(3600));
    let start_time = OsmosisTimestamp {
        seconds: start_time.seconds() as i64,
        nanos: 0_i32,
    };

    let end_time = OsmosisTimestamp {
        seconds: now.seconds() as i64,
        nanos: 0_i32,
    };

    deps.api.debug(&format!("twap_price: {twap_price}"));

    for route_part in route {
        deps.api
            .debug(&format!("route part: {sell_denom:?} {route_part:?}"));

        let twap = TwapQuerier::new(&deps.querier)
            .arithmetic_twap(
                route_part.pool_id,
                sell_denom.clone(),                 // base_asset
                route_part.token_out_denom.clone(), // quote_asset
                Some(start_time.clone()),
                Some(end_time.clone()),
            )
            .map_err(|_e| ContractError::TwapNotFound {
                denom: route_part.token_out_denom.clone(),
                sell_denom,
                pool_id: route_part.pool_id,
            })?
            .arithmetic_twap;

        deps.api.debug(&format!("twap = {twap}"));

        let twap: Decimal = twap
            .parse()
            .map_err(|_e| ContractError::InvalidTwapString { twap })?;

        twap_price =
            twap_price
                .checked_mul(twap)
                .map_err(|_e| ContractError::InvalidTwapOperation {
                    operation: format!("{twap_price} * {twap}"),
                })?;

        // the current output is the input for the next route_part
        sell_denom = route_part.token_out_denom;
        deps.api.debug(&format!("twap_price: {twap_price}"));
    }

    twap_price = twap_price - twap_price * percentage;
    deps.api.debug(&format!(
        "twap_price minus {percentage_impact}%: {twap_price}"
    ));

    let min_out: Uint128 = input_token.amount * twap_price;
    deps.api.debug(&format!("min: {min_out}"));

    Ok(Coin::new(min_out.into(), output_denom))
}
