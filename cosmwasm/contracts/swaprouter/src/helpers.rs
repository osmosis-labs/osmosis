use std::ops::{Div, Mul};

use cosmwasm_std::{Addr, Coin, Decimal, Deps, Timestamp, Uint128};
use osmosis_std::shim::Timestamp as OsmosisTimestamp;
use osmosis_std::types::osmosis::gamm::v1beta1::{
    MsgSwapExactAmountIn, QueryTotalPoolLiquidityRequest, SwapAmountInRoute,
};
use osmosis_std::types::osmosis::twap::v1beta1::TwapQuerier;

use crate::{
    state::{ROUTING_TABLE, STATE},
    ContractError,
};

pub fn check_is_contract_owner(deps: Deps, sender: Addr) -> Result<(), ContractError> {
    let config = STATE.load(deps.storage).unwrap();
    println!("here");
    println!("{}, {}", config.owner, sender);
    if config.owner != sender {
        Err(ContractError::Unauthorized {})
    } else {
        Ok(())
    }
}

pub fn validate_pool_route(
    deps: Deps,
    input_denom: String,
    output_denom: String,
    pool_route: Vec<SwapAmountInRoute>,
) -> Result<(), ContractError> {
    let mut current_denom = input_denom;

    // make sure that this route actually works
    for route_part in &pool_route {
        let liquidity = QueryTotalPoolLiquidityRequest {
            pool_id: route_part.pool_id,
        }
        .query(&deps.querier)
        .map_err(|_e| ContractError::QueryError {
            val: format!("Couldn't query liquidity for pool {}", route_part.pool_id),
        })?
        .liquidity;

        if !liquidity.iter().any(|coin| coin.denom == current_denom) {
            return Result::Err(ContractError::InvalidPoolRoute {
                reason: format!(
                    "denom {} is not in pool id {}",
                    current_denom, route_part.pool_id
                ),
            });
        }

        if !liquidity
            .iter()
            .any(|coin| coin.denom == route_part.token_out_denom)
        {
            return Result::Err(ContractError::InvalidPoolRoute {
                reason: format!(
                    "denom {} is not in pool id {}",
                    current_denom, route_part.pool_id
                ),
            });
        }

        current_denom = route_part.token_out_denom.clone();
    }

    // make sure the final route output asset is the same as the expected output_denom
    if current_denom != output_denom {
        return Result::Err(ContractError::InvalidPoolRoute {
            reason: "last denom doesn't match".to_string(),
        });
    }

    Ok(())
}

pub fn generate_swap_msg(
    deps: Deps,
    sender: Addr,
    input_token: Coin,
    min_output_token: Coin,
) -> Result<MsgSwapExactAmountIn, ContractError> {
    // get trade route
    let route = ROUTING_TABLE.load(deps.storage, (&input_token.denom, &min_output_token.denom))?;
    Ok(MsgSwapExactAmountIn {
        sender: sender.into_string(),
        routes: route,
        token_in: Some(input_token.into()),
        token_out_min_amount: min_output_token.amount.to_string(),
    })
}

pub fn calculate_min_output_from_twap(
    deps: Deps,
    input_token: Coin,
    output_denom: String,
    now: Timestamp,
    window: Option<u64>,
    percentage_impact: Decimal,
) -> Result<Coin, ContractError> {
    // get trade route
    let route = ROUTING_TABLE
        .load(deps.storage, (&input_token.denom, &output_denom))
        .unwrap_or_default();
    if route.is_empty() {
        return Err(ContractError::InvalidPoolRoute {
            reason: format!("No route foung for {} -> {output_denom}", input_token.denom),
        });
    }

    let percentage = percentage_impact.div(Uint128::new(100));

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

    // deps.api.debug(&format!("twap_price: {twap_price}"));

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

    twap_price = twap_price - twap_price.mul(percentage);
    // deps.api.debug(&format!(
    //     "twap_price minus {percentage_impact}%: {twap_price}"
    // ));

    let min_out: Uint128 = input_token.amount.mul(twap_price);
    // deps.api.debug(&format!("min: {min_out}"));

    Ok(Coin::new(min_out.into(), output_denom))
}
