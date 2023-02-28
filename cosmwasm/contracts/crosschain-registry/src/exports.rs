use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Deps, StdError};
use itertools::Itertools;

use crate::{
    helpers::{hash_denom_trace, DenomTrace, QueryDenomTraceRequest},
    msg::QueryMsg,
};
use std::convert::AsRef;

#[cw_serde]
pub struct Chain(String);
#[cw_serde]
pub struct ChannelId(String);

impl ChannelId {
    pub fn new(channel_id: &str) -> Result<Self, StdError> {
        if !ChannelId::validate(channel_id) {
            return Err(StdError::generic_err("Invalid channel id"));
        }
        Ok(Self(channel_id.to_string()))
    }

    pub fn validate(channel_id: &str) -> bool {
        if !channel_id.starts_with("channel-") {
            return false;
        }
        // Check that what comes after "channel-" is a valid int
        let channel_num = &channel_id[8..];
        if channel_num.parse::<u64>().is_err() {
            return false;
        }
        true
    }
}

impl AsRef<str> for ChannelId {
    fn as_ref(&self) -> &str {
        &self.0
    }
}

impl AsRef<str> for Chain {
    fn as_ref(&self) -> &str {
        &self.0
    }
}

// We will assume here that chains use the standard ibc-go formats. This is ok
// because we will be checking the channels in the registry and failing if they
// are not valid. We also need to enforce that all ports are explicitly "transfer"
#[cw_serde]
pub struct MultiHopDenom {
    pub local_denom: String,
    pub on: Chain,
    pub via: Option<ChannelId>, // This is optional because native tokens have no channel
}

pub struct Registries<'a> {
    pub deps: Deps<'a>,
    pub registry_contract: String,
}

impl<'a> Registries<'a> {
    pub fn new(deps: Deps<'a>, registry_contract: String) -> Result<Self, StdError> {
        deps.api.addr_validate(&registry_contract)?;
        Ok(Self {
            deps,
            registry_contract,
        })
    }

    #[allow(dead_code)]
    fn default(deps: Deps<'a>) -> Self {
        Self {
            deps,
            registry_contract: "todo: hard code the addr here".to_string(),
        }
    }

    pub fn get_contract(self: &Self, alias: String) -> Result<String, StdError> {
        self.deps.querier.query_wasm_smart(
            &self.registry_contract,
            &QueryMsg::GetAddressFromAlias {
                contract_alias: alias,
            },
        )
    }

    pub fn get_channel(self: &Self, from_chain: &str, to_chain: &str) -> Result<String, StdError> {
        self.deps.querier.query_wasm_smart(
            &self.registry_contract,
            &QueryMsg::GetChainToChainChannelLink {
                source_chain: from_chain.to_string(),
                destination_chain: to_chain.to_string(),
            },
        )
    }

    pub fn get_connected_chain(
        &self,
        on_chain: &str,
        via_channel: &str,
    ) -> Result<String, StdError> {
        self.deps.querier.query_wasm_smart(
            &self.registry_contract,
            &QueryMsg::GetConnectedChainViaChannel {
                on_chain: on_chain.to_string(),
                via_channel: via_channel.to_string(),
            },
        )
    }

    pub fn unwrap_denom(self: &Self, denom: &str) -> Result<Vec<MultiHopDenom>, StdError> {
        // Check that the denom is an IBC denom
        if !denom.starts_with("ibc/") {
            return Err(StdError::generic_err(format!(
                "Denom {denom} is not an IBC denom",
            )));
        }

        // Get the denom trace
        let res = QueryDenomTraceRequest {
            hash: denom.to_string(),
        }
        .query(&self.deps.querier)?;

        let DenomTrace { path, base_denom } = match res.denom_trace {
            Some(denom_trace) => Ok(denom_trace),
            None => Err(StdError::generic_err("No denom trace found")),
        }?;

        let mut hops: Vec<MultiHopDenom> = vec![];
        let mut current_chain = "osmosis".to_string();
        let rest = path.clone();
        let parts = path.split('/');

        for (port, channel) in parts.tuple_windows() {
            // Check that the port is "transfer"
            if port != "transfer" {
                return Err(StdError::generic_err(format!(
                    "Port {} is not a valid port",
                    port
                )));
            }

            // Check that the channel is valid
            let full_trace = rest.clone() + &base_denom;
            hops.push(MultiHopDenom {
                local_denom: hash_denom_trace(&full_trace),
                on: Chain(current_chain.clone().to_string()),
                via: Some(ChannelId::new(channel)?),
            });

            current_chain = self
                .get_connected_chain(&current_chain, channel)
                .map_err(|e| {
                    StdError::generic_err(format!(
                        "Error getting connected chain for {}/{}: {}",
                        current_chain, channel, e
                    ))
                })?;
        }

        hops.push(MultiHopDenom {
            local_denom: base_denom,
            on: Chain(current_chain),
            via: None,
        });

        Ok(hops)
    }
}
