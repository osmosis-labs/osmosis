use cosmwasm_std::{Addr, Deps};
use itertools::{self, Itertools};
use registry::Registry;

use crate::state::CONFIG;
use crate::ContractError;

pub fn check_is_contract_governor(deps: Deps, sender: Addr) -> Result<(), ContractError> {
    let config = CONFIG.load(deps.storage).unwrap();
    if config.governor != sender {
        Err(ContractError::Unauthorized {})
    } else {
        Ok(())
    }
}

/// If the specified receiver is just a bech32 addr, validate that the receiver
/// address is a valid address for the destination chain. This will prevent IBC
/// transfers from failing after forwarding
fn validate_bech32_receiver(deps: Deps, receiver: &str) -> Result<(String, Addr), ContractError> {
    let Ok((prefix, _, _)) = bech32::decode(receiver) else {
        return Err(ContractError::InvalidReceiver { receiver: receiver.to_string() })
    };

    let registry = Registry::default(deps);
    let chain = registry.get_chain_for_bech32_prefix(&prefix)?;

    Ok((chain, Addr::unchecked(receiver)))
}

// If the specified receiver contains the chain name, we get the chain's channel
// from the registry and check that the receiver's prefix match the chain
fn validate_chain_receiver(deps: Deps, receiver: &str) -> Result<(String, Addr), ContractError> {
    let Some((chain, addr)) = receiver.split('/').collect_tuple() else {
        return Err(ContractError::InvalidReceiver { receiver: receiver.to_string() })
    };

    // validate that the prefix of the receiver matches the chain
    let registry = Registry::default(deps);
    let (prefix, _, _) = registry.decode_bech32_addr(receiver)?;
    let prefix_for_chain = registry.get_bech32_prefix(chain)?;

    if prefix_for_chain != prefix {
        return Err(ContractError::InvalidReceiverChainPrefix {
            prefix,
            chain: chain.to_string(),
            chain_prefix: prefix_for_chain,
        });
    }

    Ok((chain.to_string(), Addr::unchecked(addr)))
}

/// The receiver can be specified explicitly (ibc:channel-n/osmo1...) or in a
/// simplified way (osmo1...).
///
/// The explicit way will allow senders to use any channel/addr combination they
/// want at the risk of more complexity and transaction failures if not properly
/// specified.
///
/// The simplified way will check in the channel registry to extract the
/// appropriate channel for the addr
pub fn validate_receiver(deps: Deps, receiver: &str) -> Result<(String, Addr), ContractError> {
    if receiver.contains('/') {
        validate_chain_receiver(deps, receiver)
    } else {
        validate_bech32_receiver(deps, receiver)
    }
}

pub fn ensure_map(json: &serde_cw_value::Value) -> Result<(), ContractError> {
    match json {
        serde_cw_value::Value::Map(_) => Ok(()),
        _ => Err(ContractError::InvalidJson {
            error: format!("invalid json: expected an object"),
            json: registry::utils::stringify(json)?,
        }),
    }
}

pub fn ensure_key_missing(
    json_object: &serde_cw_value::Value,
    key: &str,
) -> Result<(), ContractError> {
    ensure_map(json_object)?;
    let serde_cw_value::Value::Map(m) = json_object else {
        unreachable!()
    };

    if m.get(&serde_cw_value::Value::String(key.to_string()))
        .is_some()
    {
        Err(ContractError::InvalidJson {
            error: format!("invalid json: {key} key not allowed"),
            json: registry::utils::stringify(json_object)?,
        })
    } else {
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::state::Config;
    use cosmwasm_std::testing::mock_dependencies;

    #[test]
    fn test_check_is_contract_governor() {
        let mut deps = mock_dependencies();
        let config = Config {
            governor: Addr::unchecked("governor"),
            swap_contract: Addr::unchecked("governor"),
        };
        CONFIG.save(deps.as_mut().storage, &config).unwrap();
        let sender = Addr::unchecked("governor");
        let res = check_is_contract_governor(deps.as_ref(), sender);
        assert!(res.is_ok());
        let sender = Addr::unchecked("someone_else");
        let res = check_is_contract_governor(deps.as_ref(), sender);
        assert!(res.is_err());
    }
}
