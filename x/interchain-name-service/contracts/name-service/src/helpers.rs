use crate::error::ContractError;
use cosmwasm_std::{Addr, BankMsg, Coin, Response, Timestamp, Uint128};
use cw_utils::Expiration;

const IBC_SUFFIX: &str = ".ibc";
const MIN_NAME_LENGTH: u64 = 3;
const MAX_NAME_LENGTH: u64 = 64;
// There are 31,556,952 seconds in an average Gregoarian year due to
// leap years, end-of-century common years, and leap century years.
const AVERAGE_SECONDS_PER_YEAR: u64 = 31_556_952;

pub fn assert_sent_sufficient_coin(
    sent: &[Coin],
    required: Option<Coin>,
) -> Result<(), ContractError> {
    if let Some(required_coin) = required {
        let required_amount = required_coin.amount.u128();
        if required_amount > 0 {
            let sent_sufficient_funds = sent.iter().any(|coin| {
                // check if a given sent coin matches denom
                // and has sufficient amount
                coin.denom == required_coin.denom && coin.amount.u128() >= required_amount
            });

            if sent_sufficient_funds {
                return Ok(());
            } else {
                return Err(ContractError::InsufficientFundsSent {});
            }
        }
    }
    Ok(())
}

// calculate escrow needed given a transaction price and taxes for a certain number of years
pub fn calculate_required_escrow(
    price: Uint128,
    annual_tax_bps: Uint128,
    years: Uint128,
) -> Uint128 {
    let tax_per_year = annual_tax_bps * price / Uint128::from(10_000 as u128);
    let total_tax = tax_per_year * years;
    price + total_tax
}

// let's not import a regexp library and just do these checks by hand
fn invalid_char(c: char) -> bool {
    let is_valid = c.is_digit(10) || c.is_ascii_lowercase() || (c == '-' || c == '_');
    !is_valid
}

pub fn calculate_expiry(now: Timestamp, years: Uint128) -> Expiration {
    let expiry_ts = now.plus_seconds(AVERAGE_SECONDS_PER_YEAR * years.u128() as u64);
    Expiration::AtTime(expiry_ts)
}

/// validate_name returns an error if the name is invalid
/// (we require 3-64 lowercase ascii letters , numbers, or "-" "_")
/// ends in `.ibc` suffix, no other periods are allowed
pub fn validate_name(name_with_suffix: &str) -> Result<(), ContractError> {
    let length = name_with_suffix.len() as u64;
    let (name, suffix) = {
        let full_length = name_with_suffix.len();
        name_with_suffix.split_at(full_length - IBC_SUFFIX.len())
    };

    if suffix != IBC_SUFFIX {
        return Err(ContractError::NameNeedsSuffix {
            suffix: IBC_SUFFIX.to_string(),
        });
    }

    if (name.len() as u64) < MIN_NAME_LENGTH {
        Err(ContractError::NameTooShort {
            length,
            min_length: MIN_NAME_LENGTH,
        })
    } else if (name.len() as u64) > MAX_NAME_LENGTH {
        Err(ContractError::NameTooLong {
            length,
            max_length: MAX_NAME_LENGTH,
        })
    } else {
        match name.find(invalid_char) {
            None => Ok(()),
            Some(bytepos_invalid_char_start) => {
                let c = name[bytepos_invalid_char_start..].chars().next().unwrap();
                Err(ContractError::InvalidCharacter { c })
            }
        }
    }
}

pub fn send_tokens(to_address: Addr, amount: Vec<Coin>, action: &str) -> Response {
    Response::new()
        .add_message(BankMsg::Send {
            to_address: to_address.clone().into(),
            amount,
        })
        .add_attribute("action", action)
        .add_attribute("to", to_address)
}

#[cfg(test)]
mod test {
    use super::*;
    use cosmwasm_std::{coin, coins};

    #[test]
    fn assert_sent_sufficient_coin_works() {
        match assert_sent_sufficient_coin(&[], Some(coin(0, "token"))) {
            Ok(()) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        };

        match assert_sent_sufficient_coin(&[], Some(coin(5, "token"))) {
            Ok(()) => panic!("Should have raised insufficient funds error"),
            Err(ContractError::InsufficientFundsSent {}) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        };

        match assert_sent_sufficient_coin(&coins(10, "smokin"), Some(coin(5, "token"))) {
            Ok(()) => panic!("Should have raised insufficient funds error"),
            Err(ContractError::InsufficientFundsSent {}) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        };

        match assert_sent_sufficient_coin(&coins(10, "token"), Some(coin(5, "token"))) {
            Ok(()) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        };

        let sent_coins = vec![coin(2, "smokin"), coin(5, "token"), coin(1, "earth")];
        match assert_sent_sufficient_coin(&sent_coins, Some(coin(5, "token"))) {
            Ok(()) => {}
            Err(e) => panic!("Unexpected error: {:?}", e),
        };
    }
}
