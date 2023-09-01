use crate::helpers::*;
use crate::state::{
    ChainPFM, Config, CHAIN_ADMIN_MAP, CHAIN_MAINTAINER_MAP, CHAIN_PFM_MAP,
    CHAIN_TO_BECH32_PREFIX_MAP, CHAIN_TO_BECH32_PREFIX_REVERSE_MAP, CHAIN_TO_CHAIN_CHANNEL_MAP,
    CHANNEL_ON_CHAIN_CHAIN_MAP, CONFIG, CONTRACT_ALIAS_MAP, DENOM_ALIAS_MAP,
    DENOM_ALIAS_REVERSE_MAP, GLOBAL_ADMIN_MAP,
};
use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, BankMsg, DepsMut, Env, MessageInfo, Response};
use cw_storage_plus::Map;
use registry::msg::Callback;
use registry::{Registry, RegistryError};

use crate::ContractError;

// Enum to represent the operation to be performed
#[cw_serde]
pub enum Operation {
    Set,
    Change,
    Remove,
}

// Enum to represent the operation to be performed (including enable/disable)
#[cw_serde]
pub enum FullOperation {
    Set,
    Change,
    Remove,
    Enable,
    Disable,
}

// Contract Registry

// Struct for input data for a single contract alias
#[cw_serde]
pub struct ContractAliasInput {
    pub operation: Operation,
    pub alias: String,
    pub address: Option<String>,
    pub new_alias: Option<String>,
}

// Struct for input data for a denom alias
#[cw_serde]
pub struct DenomAliasInput {
    pub operation: FullOperation,
    pub alias: String,
    pub full_denom_path: String,
}

// Transfer ownership of this contract
pub fn transfer_ownership(
    deps: DepsMut,
    sender: Addr,
    new_owner: String,
) -> Result<Response, ContractError> {
    // only owner can transfer
    if !is_owner(deps.as_ref(), &sender) {
        return Err(ContractError::Unauthorized {});
    }

    let new_owner = deps.api.addr_validate(&new_owner)?;

    CONFIG.update(
        deps.storage,
        |mut config| -> Result<Config, ContractError> {
            config.owner = new_owner;
            Ok(config)
        },
    )?;

    Ok(Response::new().add_attribute("action", "transfer_ownership"))
}

pub fn propose_pfm(
    ctx: (DepsMut, Env, MessageInfo),
    chain: String,
) -> Result<Response, ContractError> {
    let (deps, env, info) = ctx;

    // enforce lowercase
    let chain = chain.to_lowercase();

    let own_addr = env.contract.address;

    // validation
    let registry = Registry::new(deps.as_ref(), own_addr.to_string())?;
    let coin = cw_utils::one_coin(&info)?;
    let native_chain = registry.get_native_chain(&coin.denom)?;

    if native_chain.as_ref() != chain {
        return Err(ContractError::CoinFromInvalidChain {
            supplied_chain: native_chain.as_ref().to_string(),
            expected_chain: chain,
        });
    }

    // Temporarily check that only the global admin can propose a PFM. This is
    // due to different versions of PFM having different senders. Once all
    // chains are on the latest PFM, we can remove this check and uncomment the
    // code in validate_pfm
    check_action_permission(FullOperation::Set, Permission::GlobalAdmin)?;

    // check if the chain is already registered or is in progress
    if let Some(chain_pfm) = CHAIN_PFM_MAP.may_load(deps.storage, &chain)? {
        if chain_pfm.is_validated() {
            // Only authorized addresses can ask for a validated PFM to be re-checked
            // If sender is the contract governor, then they are authorized to do do this to any chain
            // Otherwise, they must be authorized to do manage the chain they are attempting to modify
            let user_permission =
                check_is_authorized(deps.as_ref(), info.sender.clone(), Some(chain.clone()))?;
            check_action_permission(FullOperation::Change, user_permission)?;
        } else {
            return Err(ContractError::PFMValidationAlreadyInProgress {
                chain: chain.clone(),
            });
        }
    };

    // Store the chain to validate
    CHAIN_PFM_MAP.save(deps.storage, &chain, &ChainPFM::new(info.sender))?;

    // redeclaring (shadowing) registry to avoid issues with the borrow checker
    let registry = Registry::new(deps.as_ref(), own_addr.to_string())?;
    let ibc_transfer = registry.unwrap_coin_into(
        coin,
        own_addr.to_string(),
        None,
        own_addr.to_string(),
        env.block.time,
        format!(r#"{{"ibc_callback":"{own_addr}"}}"#),
        Some(Callback {
            contract: own_addr,
            msg: format!(r#"{{"validate_pfm": {{"chain": "{chain}"}} }}"#).try_into()?,
        }),
        true,
    )?;

    Ok(Response::default().add_message(ibc_transfer))
}

pub fn validate_pfm(
    ctx: (DepsMut, Env, MessageInfo),
    chain: String,
) -> Result<Response, ContractError> {
    let (deps, _env, info) = ctx;

    let chain = chain.to_lowercase();

    // TODO: Uncomment this once all chains are on the latest PFM and we can
    // properly verify the sender. We will also need to modify how
    // derive_wasmhooks_sender works at that point
    //
    // let registry = Registry::default(deps.as_ref());
    // let channel = registry.get_channel(&chain, CONTRACT_CHAIN)?;
    // let own_addr = env.contract.address.as_str();
    // let original_sender = registry.encode_addr_for_chain(own_addr, &chain)?;
    // let expected_sender = registry::derive_wasmhooks_sender(&channel, &original_sender, "osmo")?;
    // if expected_sender != info.sender {
    //     return Err(ContractError::InvalidSender {
    //         expected_sender,
    //         actual_sender: info.sender.into_string(),
    //     });
    // }

    let mut chain_pfm = CHAIN_PFM_MAP.load(deps.storage, &chain).map_err(|_| {
        ContractError::ValidationNotFound {
            chain: chain.clone(),
        }
    })?;

    let initiator = match chain_pfm.initiator {
        Some(initiator) => initiator,
        None => return Err(ContractError::PFMNoInitiator {}),
    };

    let coin = cw_utils::one_coin(&info)?;
    let bank_msg = BankMsg::Send {
        to_address: initiator.to_string(),
        amount: vec![coin],
    };

    chain_pfm.validated = true;
    chain_pfm.initiator = None;

    CHAIN_PFM_MAP.save(deps.storage, &chain, &chain_pfm)?;

    Ok(Response::default().add_message(bank_msg))
}

// Set, change, or remove a contract alias to an address
pub fn contract_alias_operations(
    deps: DepsMut,
    sender: Addr,
    operations: Vec<ContractAliasInput>,
) -> Result<Response, ContractError> {
    // Only contract governor can call contract alias CRUD operations
    check_is_contract_governor(deps.as_ref(), sender)?;

    let response = Response::new();
    for operation in operations {
        match operation.operation {
            Operation::Set => {
                if CONTRACT_ALIAS_MAP.has(deps.storage, &operation.alias) {
                    return Err(ContractError::AliasAlreadyExists {
                        alias: operation.alias,
                    });
                }
                CONTRACT_ALIAS_MAP.save(
                    deps.storage,
                    &operation.alias,
                    &operation.address.ok_or(ContractError::MissingField {
                        field: "address".to_string(),
                    })?,
                )?;
                response
                    .clone()
                    .add_attribute("set_contract_alias", operation.alias.to_string());
            }
            Operation::Change => {
                let address = CONTRACT_ALIAS_MAP
                    .load(deps.storage, &operation.alias)
                    .map_err(|_| RegistryError::AliasDoesNotExist {
                        alias: operation.alias.clone(),
                    })?;
                let new_alias = operation.new_alias.clone().unwrap_or_default().to_string();
                CONTRACT_ALIAS_MAP.save(deps.storage, &new_alias, &address)?;
                CONTRACT_ALIAS_MAP.remove(deps.storage, &operation.alias);
                response
                    .clone()
                    .add_attribute("change_contract_alias", operation.alias.to_string());
            }
            Operation::Remove => {
                CONTRACT_ALIAS_MAP
                    .load(deps.storage, &operation.alias)
                    .map_err(|_| RegistryError::AliasDoesNotExist {
                        alias: operation.alias.clone(),
                    })?;
                CONTRACT_ALIAS_MAP.remove(deps.storage, &operation.alias);
                response
                    .clone()
                    .add_attribute("remove_contract_alias", operation.alias.to_string());
            }
        }
    }
    Ok(response)
}

// Set, Change, Enable, or Disable a denom alias
pub fn denom_alias_operations(
    deps: DepsMut,
    sender: Addr,
    operations: Vec<DenomAliasInput>,
) -> Result<Response, ContractError> {
    // Only contract governor can call denom alias CRUD operations
    let is_owner = is_owner(deps.as_ref(), &sender);
    let is_global_admin = is_global_admin(deps.as_ref(), &sender);

    if !is_owner && !is_global_admin {
        return Err(ContractError::Unauthorized {});
    }

    let mut response = Response::new();
    for operation in operations {
        let denom_alias = normalize_alias(&operation.alias)?;
        let path = operation.full_denom_path;

        match operation.operation {
            FullOperation::Set => {
                if DENOM_ALIAS_MAP.has(deps.storage, &path) {
                    return Err(ContractError::AliasAlreadyExistsFor { base: path });
                }
                // TODO: This check is not enough, as disabled aliases could be
                // re-set. We need to keep track of enabled/disabled in the
                // reverse map as well
                if DENOM_ALIAS_REVERSE_MAP.has(deps.storage, &operation.alias) {
                    return Err(ContractError::AliasAlreadyExists { alias: denom_alias });
                }

                DENOM_ALIAS_MAP.save(deps.storage, &path, &(denom_alias.clone(), true).into())?;
                DENOM_ALIAS_REVERSE_MAP.save(deps.storage, &denom_alias, &(&path, true).into())?;

                response =
                    response.add_attribute("set_denom_alias", format!("{denom_alias} <=> {path}"));
            }
            FullOperation::Change => {
                if !is_owner {
                    // Only the owner can change for security reasons
                    return Err(ContractError::Unauthorized {});
                }

                // Ensure the alias exists
                let map_entry = DENOM_ALIAS_MAP
                    .load(deps.storage, &path)
                    .map_err(|_| ContractError::AliasDoesNotExistFor { base: path.clone() })?;

                let is_enabled = map_entry.enabled;
                let new_alias = normalize_alias(&operation.alias)?;

                if DENOM_ALIAS_REVERSE_MAP.has(deps.storage, &new_alias) {
                    return Err(ContractError::AliasAlreadyExists { alias: new_alias });
                }

                DENOM_ALIAS_MAP.save(deps.storage, &path, &(&new_alias, is_enabled).into())?;
                DENOM_ALIAS_REVERSE_MAP.remove(deps.storage, &map_entry.value);
                DENOM_ALIAS_REVERSE_MAP.save(deps.storage, &new_alias, &(&path, true).into())?;

                response =
                    response.add_attribute("change_denom_alias", format!("{new_alias} <=> {path}"));
            }
            FullOperation::Remove => {
                if !is_owner {
                    // Only the owner can remove for security reasons
                    return Err(ContractError::Unauthorized {});
                }
                let map_entry = DENOM_ALIAS_MAP
                    .load(deps.storage, &path)
                    .map_err(|_| ContractError::AliasDoesNotExistFor { base: path.clone() })?;
                DENOM_ALIAS_MAP.remove(deps.storage, &path);
                DENOM_ALIAS_REVERSE_MAP.remove(deps.storage, &map_entry.value);

                response = response.add_attribute("remove_denom_alias", map_entry.value);
            }
            FullOperation::Enable => {
                let map_entry = DENOM_ALIAS_MAP
                    .load(deps.storage, &path)
                    .map_err(|_| ContractError::AliasDoesNotExistFor { base: path.clone() })?;
                DENOM_ALIAS_MAP.save(
                    deps.storage,
                    &path,
                    &(map_entry.value.clone(), true).into(),
                )?;
                // Add to the enabled alias to the reverse map
                DENOM_ALIAS_REVERSE_MAP.save(
                    deps.storage,
                    &map_entry.value,
                    &(&path, true).into(),
                )?;

                response = response.add_attribute(
                    "enable_denom_alias",
                    format!("{} <=> {path}", map_entry.value),
                );
            }
            FullOperation::Disable => {
                let map_entry = DENOM_ALIAS_MAP
                    .load(deps.storage, &path)
                    .map_err(|_| ContractError::AliasDoesNotExistFor { base: path.clone() })?;
                DENOM_ALIAS_MAP.save(
                    deps.storage,
                    &path,
                    &(map_entry.value.clone(), false).into(),
                )?;
                // Disable the  alias on the reverse map
                DENOM_ALIAS_REVERSE_MAP.save(
                    deps.storage,
                    &map_entry.value,
                    &(&path, false).into(),
                )?;

                response = response
                    .add_attribute("disable_denom_alias", format!("{denom_alias} <=> {path}"));
            }
        }
    }
    Ok(response)
}

// Chain Channel Registry

// Struct for input data for a single connection
#[cw_serde]
pub struct ConnectionInput {
    pub operation: FullOperation,
    pub source_chain: String,
    pub destination_chain: String,
    pub channel_id: Option<String>,
    pub new_source_chain: Option<String>,
    pub new_destination_chain: Option<String>,
    pub new_channel_id: Option<String>,
}

// Set, change, or remove a source chain, destination chain, and channel connection
pub fn connection_operations(
    deps: DepsMut,
    sender: Addr,
    operations: Vec<ConnectionInput>,
) -> Result<Response, ContractError> {
    let response = Response::new();
    for operation in operations {
        let source_chain = operation.source_chain.to_lowercase();
        let destination_chain = operation.destination_chain.to_lowercase();
        let provided_action = operation.operation.clone();

        // Only authorized addresses can call connection CRUD operations
        // If sender is the contract governor, then they are authorized to do CRUD operations on any chain
        // Otherwise, they must be authorized to do CRUD operations on the source_chain they are attempting to modify
        let user_permission =
            check_is_authorized(deps.as_ref(), sender.clone(), Some(source_chain.clone()))?;
        check_action_permission(provided_action, user_permission)?;

        match operation.operation {
            FullOperation::Set => {
                let channel_id = operation
                    .channel_id
                    .ok_or_else(|| ContractError::InvalidInput {
                        message: "channel_id is required for set operation".to_string(),
                    })?
                    .to_lowercase();
                if CHAIN_TO_CHAIN_CHANNEL_MAP.has(deps.storage, (&source_chain, &destination_chain))
                {
                    return Err(ContractError::ChainToChainChannelLinkAlreadyExists {
                        source_chain,
                        destination_chain,
                    });
                }
                CHAIN_TO_CHAIN_CHANNEL_MAP.save(
                    deps.storage,
                    (&source_chain, &destination_chain),
                    &(channel_id.clone(), true).into(),
                )?;
                if CHANNEL_ON_CHAIN_CHAIN_MAP.has(deps.storage, (&channel_id, &source_chain)) {
                    return Err(ContractError::ChannelToChainChainLinkAlreadyExists {
                        channel_id,
                        source_chain,
                    });
                }
                CHANNEL_ON_CHAIN_CHAIN_MAP.save(
                    deps.storage,
                    (&channel_id, &source_chain),
                    &(destination_chain.clone(), true).into(),
                )?;
                response.clone().add_attribute(
                    "set_connection",
                    format!("{source_chain}-{destination_chain}"),
                );
            }
            FullOperation::Change => {
                let chain_to_chain_map = CHAIN_TO_CHAIN_CHANNEL_MAP
                    .load(deps.storage, (&source_chain, &destination_chain))
                    .map_err(|_| RegistryError::ChainChannelLinkDoesNotExist {
                        source_chain: source_chain.clone(),
                        destination_chain: destination_chain.clone(),
                    })?;
                let channel_on_chain_map = CHANNEL_ON_CHAIN_CHAIN_MAP
                    .load(deps.storage, (&chain_to_chain_map.value, &source_chain))
                    .map_err(|_| RegistryError::ChannelDoesNotExistOnChain {
                        channel_id: chain_to_chain_map.value.clone(),
                        source_chain: source_chain.clone(),
                    })?;
                if let Some(new_channel_id) = operation.new_channel_id {
                    let new_channel_id = new_channel_id.to_lowercase();
                    CHAIN_TO_CHAIN_CHANNEL_MAP.save(
                        deps.storage,
                        (&source_chain, &destination_chain),
                        &(new_channel_id.clone(), chain_to_chain_map.enabled).into(),
                    )?;
                    CHANNEL_ON_CHAIN_CHAIN_MAP
                        .remove(deps.storage, (&chain_to_chain_map.value, &source_chain));
                    CHANNEL_ON_CHAIN_CHAIN_MAP.save(
                        deps.storage,
                        (&new_channel_id, &source_chain),
                        &channel_on_chain_map,
                    )?;
                    response.clone().add_attribute(
                        "change_connection",
                        format!("{source_chain}-{destination_chain}"),
                    );
                } else if let Some(new_destination_chain) = operation.new_destination_chain {
                    let new_destination_chain = new_destination_chain.to_lowercase();
                    CHAIN_TO_CHAIN_CHANNEL_MAP
                        .remove(deps.storage, (&source_chain, &destination_chain));
                    CHAIN_TO_CHAIN_CHANNEL_MAP.save(
                        deps.storage,
                        (&source_chain, &new_destination_chain),
                        &chain_to_chain_map,
                    )?;
                    CHANNEL_ON_CHAIN_CHAIN_MAP
                        .remove(deps.storage, (&chain_to_chain_map.value, &source_chain));
                    CHANNEL_ON_CHAIN_CHAIN_MAP.save(
                        deps.storage,
                        (&chain_to_chain_map.value, &source_chain),
                        &(new_destination_chain, channel_on_chain_map.enabled).into(),
                    )?;
                    response.clone().add_attribute(
                        "change_connection",
                        format!("{source_chain}-{destination_chain}"),
                    );
                } else if let Some(new_source_chain) = operation.new_source_chain {
                    let new_source_chain = new_source_chain.to_lowercase();
                    CHAIN_TO_CHAIN_CHANNEL_MAP
                        .remove(deps.storage, (&source_chain, &destination_chain));
                    CHAIN_TO_CHAIN_CHANNEL_MAP.save(
                        deps.storage,
                        (&new_source_chain, &destination_chain),
                        &chain_to_chain_map,
                    )?;
                    CHANNEL_ON_CHAIN_CHAIN_MAP
                        .remove(deps.storage, (&chain_to_chain_map.value, &source_chain));
                    CHANNEL_ON_CHAIN_CHAIN_MAP.save(
                        deps.storage,
                        (&chain_to_chain_map.value, &new_source_chain),
                        &channel_on_chain_map,
                    )?;
                    response.clone().add_attribute(
                        "change_connection",
                        format!("{source_chain}-{destination_chain}"),
                    );
                } else {
                    return Err(ContractError::InvalidInput {
                        message: "Either new_channel_id, new_destination_chain or new_source_chain must be provided for change operation".to_string(),
                    });
                }
                response.clone().add_attribute(
                    "change_connection",
                    format!("{source_chain}-{destination_chain}"),
                );
            }
            FullOperation::Remove => {
                let chain_to_chain_map = CHAIN_TO_CHAIN_CHANNEL_MAP
                    .load(deps.storage, (&source_chain, &destination_chain))
                    .map_err(|_| RegistryError::ChainChannelLinkDoesNotExist {
                        source_chain: source_chain.clone(),
                        destination_chain: destination_chain.clone(),
                    })?;
                CHAIN_TO_CHAIN_CHANNEL_MAP
                    .remove(deps.storage, (&source_chain, &destination_chain));
                CHANNEL_ON_CHAIN_CHAIN_MAP
                    .remove(deps.storage, (&chain_to_chain_map.value, &source_chain));
                response.clone().add_attribute(
                    "remove_connection",
                    format!("{source_chain}-{destination_chain}"),
                );
            }
            FullOperation::Enable => {
                let chain_to_chain_map = CHAIN_TO_CHAIN_CHANNEL_MAP
                    .load(deps.storage, (&source_chain, &destination_chain))
                    .map_err(|_| RegistryError::ChainChannelLinkDoesNotExist {
                        source_chain: source_chain.clone(),
                        destination_chain: destination_chain.clone(),
                    })?;
                let channel_on_chain_map = CHANNEL_ON_CHAIN_CHAIN_MAP
                    .load(deps.storage, (&chain_to_chain_map.value, &source_chain))
                    .map_err(|_| RegistryError::ChannelDoesNotExistOnChain {
                        channel_id: chain_to_chain_map.value.clone(),
                        source_chain: source_chain.clone(),
                    })?;
                CHAIN_TO_CHAIN_CHANNEL_MAP.save(
                    deps.storage,
                    (&source_chain, &destination_chain),
                    &(&chain_to_chain_map.value, true).into(),
                )?;
                CHANNEL_ON_CHAIN_CHAIN_MAP.save(
                    deps.storage,
                    (&chain_to_chain_map.value, &source_chain),
                    &(channel_on_chain_map.value, true).into(),
                )?;
                response.clone().add_attribute(
                    "enable_connection",
                    format!("{source_chain}-{destination_chain}"),
                );
            }
            FullOperation::Disable => {
                let chain_to_chain_map = CHAIN_TO_CHAIN_CHANNEL_MAP
                    .load(deps.storage, (&source_chain, &destination_chain))
                    .map_err(|_| RegistryError::ChainChannelLinkDoesNotExist {
                        source_chain: source_chain.clone(),
                        destination_chain: destination_chain.clone(),
                    })?;
                let channel_on_chain_map = CHANNEL_ON_CHAIN_CHAIN_MAP
                    .load(deps.storage, (&chain_to_chain_map.value, &source_chain))
                    .map_err(|_| RegistryError::ChannelDoesNotExistOnChain {
                        channel_id: chain_to_chain_map.value.clone(),
                        source_chain: source_chain.clone(),
                    })?;
                CHAIN_TO_CHAIN_CHANNEL_MAP.save(
                    deps.storage,
                    (&source_chain, &destination_chain),
                    &(&chain_to_chain_map.value, false).into(),
                )?;
                CHANNEL_ON_CHAIN_CHAIN_MAP.save(
                    deps.storage,
                    (&chain_to_chain_map.value, &source_chain),
                    &(channel_on_chain_map.value, false).into(),
                )?;
                response.clone().add_attribute(
                    "disable_connection",
                    format!("{source_chain}-{destination_chain}"),
                );
            }
        }
    }
    Ok(response)
}

// Struct for input data for a single chain to bech32 prefix operation
#[cw_serde]
pub struct ChainToBech32PrefixInput {
    pub operation: FullOperation,
    pub chain_name: String,
    pub prefix: String,
    pub new_prefix: Option<String>,
}

pub fn chain_to_prefix_operations(
    deps: DepsMut,
    sender: Addr,
    operations: Vec<ChainToBech32PrefixInput>,
) -> Result<Response, ContractError> {
    let response = Response::new();
    for operation in operations {
        let chain_name = operation.chain_name.to_lowercase();
        let provided_action = operation.operation.clone();

        // Only authorized addresses can call connection CRUD operations
        // If sender is the contract governor, then they are authorized to do CRUD operations on any chain
        // Otherwise, they must be authorized to do CRUD operations on the source_chain they are attempting to modify
        let user_permission =
            check_is_authorized(deps.as_ref(), sender.clone(), Some(chain_name.clone()))?;
        check_action_permission(provided_action, user_permission)?;

        match operation.operation {
            FullOperation::Set => {
                if CHAIN_TO_BECH32_PREFIX_MAP.has(deps.storage, &chain_name) {
                    return Err(ContractError::AliasAlreadyExists { alias: chain_name });
                }
                let prefix = operation.prefix.to_lowercase();
                CHAIN_TO_BECH32_PREFIX_MAP.save(
                    deps.storage,
                    &chain_name,
                    &(prefix.clone(), true).into(),
                )?;

                push_to_map_value(
                    deps.storage,
                    &CHAIN_TO_BECH32_PREFIX_REVERSE_MAP,
                    &prefix,
                    chain_name.clone(),
                )?;

                response
                    .clone()
                    .add_attribute("set_chain_to_prefix", chain_name);
            }
            FullOperation::Change => {
                let map_entry = CHAIN_TO_BECH32_PREFIX_MAP
                    .load(deps.storage, &chain_name)
                    .map_err(|_| RegistryError::AliasDoesNotExist {
                        alias: chain_name.clone(),
                    })?;

                let is_enabled = map_entry.enabled;

                let old_prefix = operation.prefix.to_lowercase();
                let new_prefix = operation
                    .new_prefix
                    .unwrap_or_default()
                    .to_string()
                    .to_lowercase();
                CHAIN_TO_BECH32_PREFIX_MAP.save(
                    deps.storage,
                    &chain_name,
                    &(new_prefix.clone(), is_enabled).into(),
                )?;

                // Remove from the reverse map of the old prefix
                remove_from_map_value(
                    deps.storage,
                    &CHAIN_TO_BECH32_PREFIX_REVERSE_MAP,
                    &old_prefix,
                    chain_name.clone(),
                )?;

                // Add to the reverse map of the new prefix
                push_to_map_value(
                    deps.storage,
                    &CHAIN_TO_BECH32_PREFIX_REVERSE_MAP,
                    &new_prefix,
                    chain_name.clone(),
                )?;

                response
                    .clone()
                    .add_attribute("change_chain_to_prefix", chain_name);
            }
            FullOperation::Remove => {
                CONTRACT_ALIAS_MAP
                    .load(deps.storage, &chain_name)
                    .map_err(|_| RegistryError::AliasDoesNotExist {
                        alias: chain_name.clone(),
                    })?;
                CHAIN_TO_BECH32_PREFIX_MAP.remove(deps.storage, &chain_name);

                let old_prefix = operation.prefix.to_lowercase();
                // Remove from the reverse map of the old prefix
                remove_from_map_value(
                    deps.storage,
                    &CHAIN_TO_BECH32_PREFIX_REVERSE_MAP,
                    &old_prefix,
                    chain_name.clone(),
                )?;

                response
                    .clone()
                    .add_attribute("remove_chain_to_prefix", chain_name);
            }
            FullOperation::Enable => {
                let map_entry = CHAIN_TO_BECH32_PREFIX_MAP
                    .load(deps.storage, &chain_name)
                    .map_err(|_| RegistryError::AliasDoesNotExist {
                        alias: chain_name.clone(),
                    })?;
                CHAIN_TO_BECH32_PREFIX_MAP.save(
                    deps.storage,
                    &chain_name,
                    &(map_entry.value.clone(), true).into(),
                )?;
                // Add to the reverse map of the enabled prefix
                push_to_map_value(
                    deps.storage,
                    &CHAIN_TO_BECH32_PREFIX_REVERSE_MAP,
                    &map_entry.value,
                    chain_name.clone(),
                )?;
                response
                    .clone()
                    .add_attribute("enable_chain_to_prefix", chain_name);
            }
            FullOperation::Disable => {
                let map_entry = CHAIN_TO_BECH32_PREFIX_MAP
                    .load(deps.storage, &chain_name)
                    .map_err(|_| RegistryError::AliasDoesNotExist {
                        alias: chain_name.clone(),
                    })?;
                CHAIN_TO_BECH32_PREFIX_MAP.save(
                    deps.storage,
                    &chain_name,
                    &(map_entry.value.clone(), false).into(),
                )?;
                // Remove from the reverse map of the disabled prefix
                remove_from_map_value(
                    deps.storage,
                    &CHAIN_TO_BECH32_PREFIX_REVERSE_MAP,
                    &map_entry.value,
                    chain_name.clone(),
                )?;

                response
                    .clone()
                    .add_attribute("disable_chain_to_prefix", chain_name);
            }
        }
    }
    Ok(response)
}

// Struct for input data for a single chain to authorized address operation
#[cw_serde]
pub struct AuthorizedAddressInput {
    pub operation: Operation,
    pub source_chain: String,
    pub addr: Addr,
    pub permission: Option<Permission>,
    pub new_addr: Option<Addr>,
}

#[cw_serde]
pub enum Permission {
    GlobalAdmin,
    ChainAdmin,
    ChainMaintainer,
}

fn permission_to_map(permission: &Permission) -> &Map<&str, Addr> {
    match permission {
        Permission::GlobalAdmin => &GLOBAL_ADMIN_MAP,
        Permission::ChainAdmin => &CHAIN_ADMIN_MAP,
        Permission::ChainMaintainer => &CHAIN_MAINTAINER_MAP,
    }
}

pub fn authorized_address_operations(
    deps: DepsMut,
    sender: Addr,
    operation: Vec<AuthorizedAddressInput>,
) -> Result<Response, ContractError> {
    let response = Response::new();
    for operation in operation {
        let addr = operation.addr;
        let source_chain = operation.source_chain.to_lowercase();
        let requested_permission = operation.permission.unwrap();

        // Check if the sender is authorized to make changes to the map of addresses authorized for the given permission
        // GlobalAdmins can add addresses to any map
        // ChainAdmins can modify the ChainAdmin and ChainMaintainer for their own chain
        // ChainMaintainers can only modify the ChainMaintainer for their own chain
        let max_permission =
            check_is_authorized(deps.as_ref(), sender.clone(), Some(source_chain.clone()))?;
        check_permission(requested_permission.clone(), max_permission)?;

        // Pull the correct map from the permission
        let address_map = permission_to_map(&requested_permission);

        match operation.operation {
            Operation::Set => {
                if address_map.has(deps.storage, &source_chain) {
                    return Err(ContractError::ChainAuthorizedAddressAlreadyExists {
                        source_chain,
                    });
                }

                address_map.save(deps.storage, &source_chain, &addr)?;
                response
                    .clone()
                    .add_attribute("set_authorized_address", format!("{source_chain}-{addr}"));
            }
            Operation::Change => {
                address_map.load(deps.storage, &source_chain).map_err(|_| {
                    RegistryError::ChainAuthorizedAddressDoesNotExist {
                        source_chain: source_chain.clone(),
                    }
                })?;

                let new_addr = operation.new_addr.unwrap();

                address_map.remove(deps.storage, &source_chain);
                address_map.save(deps.storage, &source_chain, &new_addr)?;
                response.clone().add_attribute(
                    "change_authorized_address",
                    format!("{source_chain}-{addr}"),
                );
            }
            Operation::Remove => {
                address_map.load(deps.storage, &source_chain).map_err(|_| {
                    RegistryError::ChainAuthorizedAddressDoesNotExist {
                        source_chain: source_chain.clone(),
                    }
                })?;

                address_map.remove(deps.storage, &source_chain);
                response.clone().add_attribute(
                    "remove_authorized_address",
                    format!("{source_chain}-{addr}"),
                );
            }
        }
    }
    Ok(response)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::msg::ExecuteMsg;
    use crate::query::{query_alias_for_denom_path, query_denom_path_for_alias};
    use crate::{contract, helpers::test::initialize_contract};
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    static CREATOR_ADDRESS: &str = "creator";
    static CHAIN_ADMIN: &str = "chain_admin";
    static CHAIN_MAINTAINER: &str = "chain_maintainer";
    static UNAUTHORIZED_ADDRESS: &str = "unauthorized_address";
    use crate::contract::CONTRACT_CHAIN;

    #[test]
    fn test_set_contract_alias() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());
        let alias = "swap_router".to_string();
        let address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9".to_string();

        // Set contract alias swap_router to an address
        let msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Set,
                alias: alias.clone(),
                address: Some(address),
                new_alias: None,
            }],
        };

        let info = mock_info(CREATOR_ADDRESS, &[]);
        let res = contract::execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        assert_eq!(0, res.messages.len());
        assert_eq!(
            CONTRACT_ALIAS_MAP
                .load(&deps.storage, "swap_router")
                .unwrap(),
            "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9"
        );

        // Attempt to set contract alias swap_router to a different address
        let msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Set,
                alias: alias.clone(),
                address: Some("osmo1fsdaf7dsfasndjklk3jndskajnfkdjsfjn3jka".to_string()),
                new_alias: None,
            }],
        };
        let res = contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap_err();
        assert_eq!(res, ContractError::AliasAlreadyExists { alias });

        // Verify that the alias was not updated
        assert_eq!(
            CONTRACT_ALIAS_MAP
                .load(&deps.storage, "swap_router")
                .unwrap(),
            "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9"
        );

        // Attempt to set a new contract alias new_contract_alias to an address via an unauthorized address
        let msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Set,
                alias: "new_contract_alias".to_string(),
                address: Some("osmo1nna7k5lywn99cd63elcfqm6p8c5c4qcuqwwflx".to_string()),
                new_alias: None,
            }],
        };
        let unauthorized_info = mock_info(UNAUTHORIZED_ADDRESS, &[]);
        let res = contract::execute(deps.as_mut(), mock_env(), unauthorized_info, msg).unwrap_err();
        assert_eq!(res, ContractError::Unauthorized {});

        // Verify that the new alias was not set
        assert!(!CONTRACT_ALIAS_MAP.has(&deps.storage, "new_contract_alias"));
    }

    #[test]
    fn test_modify_contract_alias() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        let creator_info = mock_info(CREATOR_ADDRESS, &[]);
        let external_unauthorized_info = mock_info(UNAUTHORIZED_ADDRESS, &[]);

        let alias = "swap_router".to_string();
        let address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9".to_string();
        let new_alias = "new_swap_router".to_string();
        let new_alias_unauthorized = "new_new_swap_router".to_string();

        // Set the contract alias swap_router to an address
        let set_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Set,
                alias: alias.clone(),
                address: Some(address.clone()),
                new_alias: None,
            }],
        };
        let set_alias_result = contract::execute(
            deps.as_mut(),
            mock_env(),
            creator_info.clone(),
            set_alias_msg,
        );
        assert!(set_alias_result.is_ok());

        // Change the contract alias swap_router to new_swap_router
        let change_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Change,
                alias: alias.clone(),
                address: None,
                new_alias: Some(new_alias.clone()),
            }],
        };
        let change_alias_result = contract::execute(
            deps.as_mut(),
            mock_env(),
            creator_info.clone(),
            change_alias_msg,
        );
        assert!(change_alias_result.is_ok());

        // Verify that the contract alias has changed from "swap_router" to "new_swap_router"
        assert_eq!(
            CONTRACT_ALIAS_MAP.load(&deps.storage, &new_alias).unwrap(),
            address
        );

        // Attempt to change an alias that does not exist
        let invalid_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Change,
                alias: alias.clone(),
                address: None,
                new_alias: Some(new_alias.clone()),
            }],
        };
        let invalid_alias_result =
            contract::execute(deps.as_mut(), mock_env(), creator_info, invalid_alias_msg);
        let expected_error = ContractError::from(RegistryError::AliasDoesNotExist { alias });
        assert_eq!(invalid_alias_result.unwrap_err(), expected_error);

        // Attempt to change an existing alias to a new alias but with an unauthorized address
        let unauthorized_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Change,
                alias: new_alias,
                address: None,
                new_alias: Some(new_alias_unauthorized.clone()),
            }],
        };
        let unauthorized_alias_result = contract::execute(
            deps.as_mut(),
            mock_env(),
            external_unauthorized_info,
            unauthorized_alias_msg,
        );
        let expected_error = ContractError::Unauthorized {};
        assert_eq!(unauthorized_alias_result.unwrap_err(), expected_error);
        assert!(!CONTRACT_ALIAS_MAP.has(&deps.storage, &new_alias_unauthorized));
    }

    #[test]
    fn test_remove_contract_alias() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        let alias = "swap_router".to_string();
        let address = "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjatel8rck9".to_string();

        // Set contract alias "swap_router" to an address
        let set_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Set,
                alias: alias.clone(),
                address: Some(address.clone()),
                new_alias: None,
            }],
        };
        let creator_info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(
            deps.as_mut(),
            mock_env(),
            creator_info.clone(),
            set_alias_msg,
        )
        .unwrap();

        // Remove the alias
        let remove_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Remove,
                alias: alias.clone(),
                address: Some(address.clone()),
                new_alias: None,
            }],
        };
        contract::execute(
            deps.as_mut(),
            mock_env(),
            creator_info.clone(),
            remove_alias_msg,
        )
        .unwrap();

        // Verify that the alias no longer exists
        let alias_exists = CONTRACT_ALIAS_MAP
            .may_load(&deps.storage, "swap_router")
            .unwrap()
            .is_some();
        assert!(!alias_exists, "alias should not exist");

        // Attempt to remove an alias that does not exist
        let non_existing_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Remove,
                alias: "non_existing_alias".to_string(),
                address: Some(address.clone()),
                new_alias: None,
            }],
        };
        let result = contract::execute(
            deps.as_mut(),
            mock_env(),
            creator_info.clone(),
            non_existing_alias_msg,
        );

        let expected_error = ContractError::from(RegistryError::AliasDoesNotExist {
            alias: "non_existing_alias".to_string(),
        });
        assert_eq!(result.unwrap_err(), expected_error);

        // Reset the contract alias "swap_router" to an address
        let reset_alias_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Set,
                alias: alias.clone(),
                address: Some(address.clone()),
                new_alias: None,
            }],
        };
        contract::execute(deps.as_mut(), mock_env(), creator_info, reset_alias_msg).unwrap();

        // Attempt to remove an alias with an unauthorized address
        let unauthorized_remove_msg = ExecuteMsg::ModifyContractAlias {
            operations: vec![ContractAliasInput {
                operation: Operation::Remove,
                alias,
                address: Some(address),
                new_alias: None,
            }],
        };
        let unauthorized_info = mock_info(UNAUTHORIZED_ADDRESS, &[]);
        let result = contract::execute(
            deps.as_mut(),
            mock_env(),
            unauthorized_info,
            unauthorized_remove_msg,
        );

        let expected_error = ContractError::Unauthorized {};
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_set_chain_channel_link() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Set,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "COSMOS".to_string(),
                channel_id: Some("CHANNEL-0".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(&deps.storage, (CONTRACT_CHAIN, "cosmos"))
                .unwrap(),
            ("channel-0", true).into()
        );

        // Verify that channel-0 on osmosis is linked to cosmos
        assert_eq!(
            CHANNEL_ON_CHAIN_CHAIN_MAP
                .load(&deps.storage, ("channel-0", CONTRACT_CHAIN))
                .unwrap(),
            ("cosmos", true).into()
        );

        // Attempt to set the canonical channel link between osmosis and cosmos to channel-150
        // This should fail because the link already exists
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Set,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "cosmos".to_string(),
                channel_id: Some("channel-150".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let info_creator = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info_creator, msg);
        assert!(result.is_err());

        let expected_error = ContractError::ChainToChainChannelLinkAlreadyExists {
            source_chain: CONTRACT_CHAIN.to_string(),
            destination_chain: "cosmos".to_string(),
        };
        assert_eq!(result.unwrap_err(), expected_error);
        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(&deps.storage, (CONTRACT_CHAIN, "cosmos"))
                .unwrap(),
            ("channel-0", true).into()
        );
        assert_eq!(
            CHANNEL_ON_CHAIN_CHAIN_MAP
                .load(&deps.storage, ("channel-0", CONTRACT_CHAIN))
                .unwrap(),
            ("cosmos", true).into()
        );

        // Attempt to set the canonical channel link between mars and osmosis to channel-1 with an unauthorized address
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Set,
                source_chain: "mars".to_string(),
                destination_chain: CONTRACT_CHAIN.to_string(),
                channel_id: Some("channel-1".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let info_unauthorized = mock_info(UNAUTHORIZED_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info_unauthorized, msg.clone());
        assert!(result.is_err());

        let expected_error = ContractError::Unauthorized {};
        assert_eq!(result.unwrap_err(), expected_error);
        assert!(!CHAIN_TO_CHAIN_CHANNEL_MAP.has(&deps.storage, ("mars", CONTRACT_CHAIN)));

        // Set the canonical channel link between mars and osmosis to channel-1 with a mars chain admin address
        let chain_admin_info = mock_info(CHAIN_ADMIN, &[]);
        contract::execute(deps.as_mut(), mock_env(), chain_admin_info.clone(), msg).unwrap();
        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(&deps.storage, ("mars", CONTRACT_CHAIN))
                .unwrap(),
            ("channel-1", true).into()
        );
        assert_eq!(
            CHANNEL_ON_CHAIN_CHAIN_MAP
                .load(&deps.storage, ("channel-1", "mars"))
                .unwrap(),
            (CONTRACT_CHAIN, true).into()
        );

        // Set the canonical channel link between juno and mars to channel-2 with a juno chain maintainer address
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Set,
                source_chain: "juno".to_string(),
                destination_chain: "mars".to_string(),
                channel_id: Some("channel-2".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        // Note: the chain admin address for mars and osmo is the chain maintainer for juno
        // This is used to test privilege escalation next
        let chain_admin_and_maintainer_info = mock_info(CHAIN_ADMIN, &[]);
        contract::execute(
            deps.as_mut(),
            mock_env(),
            chain_admin_and_maintainer_info.clone(),
            msg,
        )
        .unwrap();
        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(&deps.storage, ("juno", "mars"))
                .unwrap(),
            ("channel-2", true).into()
        );
        assert_eq!(
            CHANNEL_ON_CHAIN_CHAIN_MAP
                .load(&deps.storage, ("channel-2", "juno"))
                .unwrap(),
            ("mars", true).into()
        );

        // Separate test to ensure that the chain maintainer for juno but a chain admin elsewhere
        // cannot perform a chain admin action (ensure no accidental privilege escalation)
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Remove,
                source_chain: "juno".to_string(),
                destination_chain: "mars".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let result = contract::execute(
            deps.as_mut(),
            mock_env(),
            chain_admin_and_maintainer_info,
            msg,
        );
        assert!(result.is_err());

        let expected_error = ContractError::Unauthorized {};
        assert_eq!(result.unwrap_err(), expected_error);

        // Attempt to set the canonical channel link between regen and mars to channel-3 with a mars chain admin address
        // This should fail because mars should not be able to set a link for regen
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Set,
                source_chain: "regen".to_string(),
                destination_chain: "mars".to_string(),
                channel_id: Some("channel-3".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let result = contract::execute(deps.as_mut(), mock_env(), chain_admin_info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::Unauthorized {};
        assert_eq!(result.unwrap_err(), expected_error);
        assert!(!CHAIN_TO_CHAIN_CHANNEL_MAP.has(&deps.storage, ("regen", "mars")));
    }

    #[test]
    fn test_change_chain_channel_link() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Set,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "COSMOS".to_string(),
                channel_id: Some("CHANNEL-0".to_string()),
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let info_creator = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info_creator.clone(), msg);
        assert!(result.is_ok());

        // Change the canonical channel link between osmosis and cosmos to channel-150 with the global admin address
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Change,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "cosmos".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: Some("channel-150".to_string()),
            }],
        };
        let result = contract::execute(deps.as_mut(), mock_env(), info_creator.clone(), msg);
        assert!(result.is_ok());

        // Verify that the channel between osmosis and cosmos has changed from channel-0 to channel-150
        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(&deps.storage, (CONTRACT_CHAIN, "cosmos"))
                .unwrap(),
            ("channel-150", true).into()
        );

        // Attempt to change a channel link that does not exist
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Change,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "regen".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: Some("channel-1".to_string()),
            }],
        };
        let result = contract::execute(deps.as_mut(), mock_env(), info_creator.clone(), msg);
        assert!(result.is_err());

        let expected_error = ContractError::from(RegistryError::ChainChannelLinkDoesNotExist {
            source_chain: CONTRACT_CHAIN.to_string(),
            destination_chain: "regen".to_string(),
        });
        assert_eq!(result.unwrap_err(), expected_error);

        // Change channel-0 link of osmosis from cosmos to regen with the global admin address
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Change,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "cosmos".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: Some("regen".to_string()),
                new_channel_id: None,
            }],
        };
        let result = contract::execute(deps.as_mut(), mock_env(), info_creator, msg);
        assert!(result.is_ok());

        // Verify that channel-150 on osmosis is linked to regen
        assert_eq!(
            CHANNEL_ON_CHAIN_CHAIN_MAP
                .load(&deps.storage, ("channel-150", CONTRACT_CHAIN))
                .unwrap(),
            ("regen", true).into()
        );

        // Attempt to change the canonical channel link between osmosis and regen to channel-2 with an unauthorized address
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Change,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "regen".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: Some("channel-2".to_string()),
            }],
        };
        let info_unauthorized = mock_info(UNAUTHORIZED_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info_unauthorized, msg.clone());
        assert!(result.is_err());

        let expected_error = ContractError::Unauthorized {};
        assert_eq!(result.unwrap_err(), expected_error);

        // Set the canonical channel link between mars and osmosis to channel-1 with a chain admin address
        let info_chain_admin = mock_info(CHAIN_ADMIN, &[]);
        contract::execute(deps.as_mut(), mock_env(), info_chain_admin, msg).unwrap();
        assert_eq!(
            CHAIN_TO_CHAIN_CHANNEL_MAP
                .load(&deps.storage, (CONTRACT_CHAIN, "regen"))
                .unwrap(),
            ("channel-2", true).into()
        );

        // Attempt to change a link that does not exist
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Change,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "cosmos".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: Some("channel-0".to_string()),
            }],
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::from(RegistryError::ChainChannelLinkDoesNotExist {
            source_chain: CONTRACT_CHAIN.to_string(),
            destination_chain: "cosmos".to_string(),
        });
        assert_eq!(result.unwrap_err(), expected_error);

        // Attempt to update a osmosis channel link with a osmosis chain maintainer address
        // Should fail because chain maintainer is not authorized to update any channel links
        let chain_maintainer_info = mock_info(CHAIN_MAINTAINER, &[]);
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Change,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "regen".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: Some("channel-4".to_string()),
            }],
        };
        let result = contract::execute(deps.as_mut(), mock_env(), chain_maintainer_info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::Unauthorized {};
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_remove_chain_channel_link() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        // Set up channels
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![
                ConnectionInput {
                    operation: FullOperation::Set,
                    source_chain: CONTRACT_CHAIN.to_string(),
                    destination_chain: "COSMOS".to_string(),
                    channel_id: Some("CHANNEL-0".to_string()),
                    new_source_chain: None,
                    new_destination_chain: None,
                    new_channel_id: None,
                },
                ConnectionInput {
                    operation: FullOperation::Set,
                    source_chain: CONTRACT_CHAIN.to_string(),
                    destination_chain: "REGEN".to_string(),
                    channel_id: Some("CHANNEL-1".to_string()),
                    new_source_chain: None,
                    new_destination_chain: None,
                    new_channel_id: None,
                },
            ],
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // Remove the osmosis cosmos link with a global admin address
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Remove,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "cosmos".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info, msg.clone()).unwrap();

        // Verify that the link no longer exists
        assert!(!CHAIN_TO_CHAIN_CHANNEL_MAP.has(&deps.storage, (CONTRACT_CHAIN, "cosmos")));

        let info = mock_info(CREATOR_ADDRESS, &[]);
        let result = contract::execute(deps.as_mut(), mock_env(), info, msg);

        let expected_error = ContractError::from(RegistryError::ChainChannelLinkDoesNotExist {
            source_chain: CONTRACT_CHAIN.to_string(),
            destination_chain: "cosmos".to_string(),
        });
        assert_eq!(result.unwrap_err(), expected_error);

        // Attempt to remove the osmosis regen link with a osmosis chain maintainer address
        // Should fail because chain maintainer is not authorized to remove any channel links
        let chain_maintainer_info = mock_info(CHAIN_MAINTAINER, &[]);
        let msg = ExecuteMsg::ModifyChainChannelLinks {
            operations: vec![ConnectionInput {
                operation: FullOperation::Remove,
                source_chain: CONTRACT_CHAIN.to_string(),
                destination_chain: "regen".to_string(),
                channel_id: None,
                new_source_chain: None,
                new_destination_chain: None,
                new_channel_id: None,
            }],
        };
        let result = contract::execute(deps.as_mut(), mock_env(), chain_maintainer_info, msg);
        assert!(result.is_err());

        let expected_error = ContractError::Unauthorized {};
        assert_eq!(result.unwrap_err(), expected_error);
    }

    #[test]
    fn test_set_bech32_prefix_to_chain() {
        let mut deps = mock_dependencies();
        initialize_contract(deps.as_mut());

        // Set the canonical channel link between osmosis and cosmos to channel-0
        let msg = ExecuteMsg::ModifyBech32Prefixes {
            operations: vec![ChainToBech32PrefixInput {
                operation: FullOperation::Set,
                chain_name: CONTRACT_CHAIN.to_string(),
                prefix: "OSMO".to_string(),
                new_prefix: None,
            }],
        };
        let info = mock_info(CREATOR_ADDRESS, &[]);
        contract::execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_MAP
                .load(&deps.storage, CONTRACT_CHAIN)
                .unwrap(),
            ("osmo", true).into()
        );
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_REVERSE_MAP
                .load(&deps.storage, "osmo")
                .unwrap(),
            vec![CONTRACT_CHAIN]
        );

        // Set another chain with the same prefix
        let msg = ExecuteMsg::ModifyBech32Prefixes {
            operations: vec![ChainToBech32PrefixInput {
                operation: FullOperation::Set,
                chain_name: "ISMISIS".to_string(),
                prefix: "OSMO".to_string(),
                new_prefix: None,
            }],
        };
        contract::execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_MAP
                .load(&deps.storage, "ismisis")
                .unwrap(),
            ("osmo", true).into()
        );
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_REVERSE_MAP
                .load(&deps.storage, "osmo")
                .unwrap(),
            vec![CONTRACT_CHAIN, "ismisis"]
        );

        // Set another chain with the same prefix
        let msg = ExecuteMsg::ModifyBech32Prefixes {
            operations: vec![ChainToBech32PrefixInput {
                operation: FullOperation::Disable,
                chain_name: CONTRACT_CHAIN.to_string(),
                prefix: "OSMO".to_string(),
                new_prefix: None,
            }],
        };
        contract::execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_MAP
                .load(&deps.storage, CONTRACT_CHAIN)
                .unwrap(),
            ("osmo", false).into()
        );
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_REVERSE_MAP
                .load(&deps.storage, "osmo")
                .unwrap(),
            vec!["ismisis"]
        );

        // Set another chain with the same prefix
        let msg = ExecuteMsg::ModifyBech32Prefixes {
            operations: vec![ChainToBech32PrefixInput {
                operation: FullOperation::Enable,
                chain_name: CONTRACT_CHAIN.to_string(),
                prefix: "OSMO".to_string(),
                new_prefix: None,
            }],
        };
        contract::execute(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_MAP
                .load(&deps.storage, CONTRACT_CHAIN)
                .unwrap(),
            ("osmo", true).into()
        );
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_REVERSE_MAP
                .load(&deps.storage, "osmo")
                .unwrap(),
            vec!["ismisis", CONTRACT_CHAIN]
        );

        // Set another chain with the same prefix
        let msg = ExecuteMsg::ModifyBech32Prefixes {
            operations: vec![ChainToBech32PrefixInput {
                operation: FullOperation::Remove,
                chain_name: CONTRACT_CHAIN.to_string(),
                prefix: "OSMO".to_string(),
                new_prefix: None,
            }],
        };
        contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_MAP
                .load(&deps.storage, "ismisis")
                .unwrap(),
            ("osmo", true).into()
        );
        assert_eq!(
            CHAIN_TO_BECH32_PREFIX_REVERSE_MAP
                .load(&deps.storage, "osmo")
                .unwrap(),
            vec!["ismisis"]
        );

        CHAIN_TO_BECH32_PREFIX_MAP
            .load(&deps.storage, CONTRACT_CHAIN)
            .unwrap_err();
    }

    #[test]
    fn test_denom_alias_operations() {
        let mut deps = mock_dependencies();

        initialize_contract(deps.as_mut());

        let path1 = "transfer/channel-0/1denom";

        let msg = ExecuteMsg::ModifyDenomAlias {
            operations: vec![DenomAliasInput {
                operation: FullOperation::Set,
                full_denom_path: path1.to_string(),
                alias: "alias1".to_string(),
            }],
        };

        // Test case: Set an alias
        let info = mock_info(CREATOR_ADDRESS, &[]);
        let res = contract::execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        assert_eq!(
            DENOM_ALIAS_MAP
                .may_load(deps.as_ref().storage, path1)
                .unwrap(),
            Some(("alias1".to_string(), true).into())
        );
        assert_eq!(
            DENOM_ALIAS_REVERSE_MAP
                .may_load(deps.as_ref().storage, "alias1")
                .unwrap(),
            Some((path1.to_string(), true).into())
        );
        assert_eq!(
            res.attributes,
            vec![("set_denom_alias".to_string(), format!("alias1 <=> {path1}"))]
        );

        // Check queries
        assert_eq!(
            query_denom_path_for_alias(deps.as_ref(), "alias1").unwrap(),
            path1
        );
        assert_eq!(
            query_alias_for_denom_path(deps.as_ref(), path1).unwrap(),
            "alias1"
        );

        // Test case: Change an alias
        let change_msg = ExecuteMsg::ModifyDenomAlias {
            operations: vec![DenomAliasInput {
                operation: FullOperation::Change,
                full_denom_path: path1.to_string(),
                alias: "newalias1".to_string(),
            }],
        };

        let change_info = mock_info(CREATOR_ADDRESS, &[]);

        let change_res =
            contract::execute(deps.as_mut(), mock_env(), change_info, change_msg).unwrap();

        assert_eq!(
            DENOM_ALIAS_MAP
                .may_load(deps.as_ref().storage, path1)
                .unwrap(),
            Some(("newalias1".to_string(), true).into())
        );
        assert_eq!(
            DENOM_ALIAS_REVERSE_MAP
                .may_load(deps.as_ref().storage, "alias1")
                .unwrap(),
            None
        );
        assert_eq!(
            DENOM_ALIAS_REVERSE_MAP
                .may_load(deps.as_ref().storage, "newalias1")
                .unwrap(),
            Some((path1.to_string(), true).into())
        );

        assert_eq!(
            change_res.attributes,
            vec![(
                "change_denom_alias".to_string(),
                format!("newalias1 <=> {path1}")
            )]
        );

        // Check queries
        query_denom_path_for_alias(deps.as_ref(), "alias1").unwrap_err();
        assert_eq!(
            query_denom_path_for_alias(deps.as_ref(), "newalias1").unwrap(),
            path1
        );
        assert_eq!(
            query_alias_for_denom_path(deps.as_ref(), path1).unwrap(),
            "newalias1"
        );

        // Test case: Disable an alias
        let disable_msg = ExecuteMsg::ModifyDenomAlias {
            operations: vec![DenomAliasInput {
                operation: FullOperation::Disable,
                full_denom_path: path1.to_string(),
                alias: "newalias1".to_string(),
            }],
        };

        let disable_info = mock_info(CREATOR_ADDRESS, &[]);
        let disable_res =
            contract::execute(deps.as_mut(), mock_env(), disable_info, disable_msg).unwrap();

        assert_eq!(
            DENOM_ALIAS_MAP
                .may_load(deps.as_ref().storage, path1)
                .unwrap(),
            Some(("newalias1".to_string(), false).into())
        );
        assert_eq!(
            DENOM_ALIAS_REVERSE_MAP
                .may_load(deps.as_ref().storage, "newalias1")
                .unwrap(),
            Some((path1.to_string(), false).into())
        );

        assert_eq!(
            disable_res.attributes,
            vec![(
                "disable_denom_alias".to_string(),
                format!("newalias1 <=> {path1}")
            )]
        );

        // Check queries
        query_denom_path_for_alias(deps.as_ref(), "newalias1").unwrap_err();
        query_alias_for_denom_path(deps.as_ref(), path1).unwrap_err();

        // Re-enable the alias
        let enable_msg = ExecuteMsg::ModifyDenomAlias {
            operations: vec![DenomAliasInput {
                operation: FullOperation::Enable,
                full_denom_path: path1.to_string(),
                alias: "doesntmatter".to_string(),
            }],
        };

        let enable_info = mock_info(CREATOR_ADDRESS, &[]);
        let enable_res =
            contract::execute(deps.as_mut(), mock_env(), enable_info, enable_msg).unwrap();

        assert_eq!(
            DENOM_ALIAS_MAP
                .may_load(deps.as_ref().storage, path1)
                .unwrap(),
            Some(("newalias1", true).into())
        );
        assert_eq!(
            DENOM_ALIAS_REVERSE_MAP
                .may_load(deps.as_ref().storage, "newalias1")
                .unwrap(),
            Some((path1.to_string(), true).into())
        );

        assert_eq!(
            enable_res.attributes,
            vec![(
                "enable_denom_alias".to_string(),
                format!("newalias1 <=> {path1}")
            )]
        );

        assert_eq!(
            query_denom_path_for_alias(deps.as_ref(), "newalias1").unwrap(),
            path1
        );
        assert_eq!(
            query_alias_for_denom_path(deps.as_ref(), path1).unwrap(),
            "newalias1"
        );

        // Test case: Remove an alias
        let remove_msg = ExecuteMsg::ModifyDenomAlias {
            operations: vec![DenomAliasInput {
                operation: FullOperation::Remove,
                full_denom_path: path1.to_string(),
                alias: "unusedaliascanbeanything".to_string(),
            }],
        };

        let remove_info = mock_info(CREATOR_ADDRESS, &[]);
        let remove_res =
            contract::execute(deps.as_mut(), mock_env(), remove_info, remove_msg).unwrap();

        assert_eq!(
            DENOM_ALIAS_MAP
                .may_load(deps.as_ref().storage, path1)
                .unwrap(),
            None
        );
        assert_eq!(
            DENOM_ALIAS_REVERSE_MAP
                .may_load(deps.as_ref().storage, "new_alias1")
                .unwrap(),
            None
        );

        assert_eq!(
            remove_res.attributes,
            vec![("remove_denom_alias".to_string(), "newalias1".to_string())]
        );

        query_denom_path_for_alias(deps.as_ref(), "newalias1").unwrap_err();
        query_alias_for_denom_path(deps.as_ref(), path1).unwrap_err();
    }

    #[test]
    fn transfer_ownership() {
        let mut deps = mock_dependencies();

        let owner = initialize_contract(deps.as_mut());
        let owner_info = mock_info(owner.as_str(), &vec![] as &Vec<cosmwasm_std::Coin>);

        let new_owner = "new_owner".to_string();
        // The owner can transfer ownership
        let msg = ExecuteMsg::TransferOwnership {
            new_owner: new_owner.clone(),
        };
        contract::execute(deps.as_mut(), mock_env(), owner_info, msg).unwrap();

        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(new_owner, config.owner);
    }

    #[test]
    fn transfer_ownership_unauthorized() {
        let mut deps = mock_dependencies();

        let owner = initialize_contract(deps.as_mut());

        let other_info = mock_info("other_sender", &vec![] as &Vec<cosmwasm_std::Coin>);

        // An unauthorized user cannot transfer ownership
        let msg = ExecuteMsg::TransferOwnership {
            new_owner: "new_owner".to_string(),
        };
        contract::execute(deps.as_mut(), mock_env(), other_info, msg).unwrap_err();

        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(owner, config.owner);
    }
}
