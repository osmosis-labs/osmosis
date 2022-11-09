use crate::state::FlowType;
use cosmwasm_std::{Addr, Deps, StdError, Timestamp, Uint256};
use osmosis_std_derive::CosmwasmExt;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Height {
    /// Previously known as "epoch"
    revision_number: Option<u64>,

    /// The height of a block
    revision_height: Option<u64>,
}

// IBC transfer data
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct FungibleTokenData {
    pub denom: String,
    amount: Uint256,
    sender: Addr,
    receiver: Addr,
}

// An IBC packet
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Packet {
    pub sequence: u64,
    pub source_port: String,
    pub source_channel: String,
    pub destination_port: String,
    pub destination_channel: String,
    pub data: FungibleTokenData,
    pub timeout_height: Height,
    pub timeout_timestamp: Option<Timestamp>,
}

// SupplyOf query message definition.
// osmosis-std doesn't currently support the SupplyOf query, so I'm defining it localy so it can be used to obtain the channel value
#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(type_url = "/cosmos.bank.v1beta1.QuerySupplyOfRequest")]
#[proto_query(
    path = "/cosmos.bank.v1beta1.Query/SupplyOf",
    response_type = QuerySupplyOfResponse
)]
pub struct QuerySupplyOfRequest {
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
}

#[derive(
    Clone,
    PartialEq,
    Eq,
    ::prost::Message,
    serde::Serialize,
    serde::Deserialize,
    schemars::JsonSchema,
    CosmwasmExt,
)]
#[proto_message(type_url = "/cosmos.bank.v1beta1.QuerySupplyOf")]
pub struct QuerySupplyOfResponse {
    #[prost(message, optional, tag = "1")]
    pub amount: ::core::option::Option<osmosis_std::types::cosmos::base::v1beta1::Coin>,
}
// End of SupplyOf query message definition

use std::str::FromStr; // Needed to parse the coin's String as Uint256

fn hash_denom(denom: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(denom.as_bytes());
    let result = hasher.finalize();
    let hash = hex::encode(result);
    format!("ibc/{}", hash.to_uppercase())
}

impl Packet {
    pub fn mock(channel_id: String, denom: String, funds: Uint256) -> Packet {
        Packet {
            sequence: 0,
            source_port: "transfer".to_string(),
            source_channel: channel_id.clone(),
            destination_port: "transfer".to_string(),
            destination_channel: channel_id,
            data: crate::packet::FungibleTokenData {
                denom,
                amount: funds,
                sender: Addr::unchecked("sender"),
                receiver: Addr::unchecked("receiver"),
            },
            timeout_height: crate::packet::Height {
                revision_number: None,
                revision_height: None,
            },
            timeout_timestamp: None,
        }
    }

    pub fn channel_value(&self, deps: Deps, direction: &FlowType) -> Result<Uint256, StdError> {
        let res = QuerySupplyOfRequest {
            denom: self.local_denom(direction),
        }
        .query(&deps.querier)?;
        Uint256::from_str(&res.amount.unwrap_or_default().amount)
    }

    pub fn get_funds(&self) -> Uint256 {
        self.data.amount
    }

    fn local_channel(&self, direction: &FlowType) -> String {
        // Pick the appropriate channel depending on whether this is a send or a recv
        match direction {
            FlowType::In => self.destination_channel.clone(),
            FlowType::Out => self.source_channel.clone(),
        }
    }

    fn receiver_chain_is_source(&self) -> bool {
        self.data
            .denom
            .starts_with(&format!("transfer/{}", self.source_channel))
    }

    fn handle_denom_for_sends(&self) -> String {
        if !self.data.denom.starts_with("transfer/") {
            // For native tokens we just use what's on the packet
            return self.data.denom.clone();
        }
        // For non-native tokens, we need to generate the IBCDenom
        hash_denom(&self.data.denom)
    }

    fn handle_denom_for_recvs(&self) -> String {
        if self.receiver_chain_is_source() {
            // These are tokens that have been sent to the counterparty and are returning
            let unprefixed = self
                .data
                .denom
                .strip_prefix(&format!("transfer/{}/", self.source_channel))
                .unwrap_or_default();
            let split: Vec<&str> = unprefixed.split('/').collect();
            if split[0] == unprefixed {
                // This is a native token. Return the unprefixed token
                unprefixed.to_string()
            } else {
                // This is a non-native that was sent to the counterparty.
                // We need to hash it.
                // The ibc-go implementation checks that the denom has been built correctly. We
                // don't need to do that here because if it hasn't, the transfer module will catch it.
                hash_denom(unprefixed)
            }
        } else {
            // Tokens that come directly from the counterparty.
            // Since the sender didn't prefix them, we need to do it here.
            let prefixed = format!("transfer/{}/", self.destination_channel) + &self.data.denom;
            hash_denom(&prefixed)
        }
    }

    fn local_denom(&self, direction: &FlowType) -> String {
        match direction {
            FlowType::In => self.handle_denom_for_recvs(),
            FlowType::Out => self.handle_denom_for_sends(),
        }
    }

    pub fn path_data(&self, direction: &FlowType) -> (String, String) {
        (self.local_channel(direction), self.local_denom(direction))
    }
}

// Helpers

// Create a new packet for testing
#[cfg(test)]
#[macro_export]
macro_rules! test_msg_send {
    (channel_id: $channel_id:expr, denom: $denom:expr, channel_value: $channel_value:expr, funds: $funds:expr) => {
        $crate::msg::SudoMsg::SendPacket {
            packet: $crate::packet::Packet::mock($channel_id, $denom, $funds),
            local_denom: Some($denom),
            channel_value_hint: Some($channel_value),
        }
    };
}

#[cfg(test)]
#[macro_export]
macro_rules! test_msg_recv {
    (channel_id: $channel_id:expr, denom: $denom:expr, channel_value: $channel_value:expr, funds: $funds:expr) => {
        $crate::msg::SudoMsg::RecvPacket {
            packet: $crate::packet::Packet::mock($channel_id, $denom, $funds),
            local_denom: Some($denom),
            channel_value_hint: Some($channel_value),
        }
    };
}
