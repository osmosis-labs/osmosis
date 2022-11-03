use cosmwasm_std::{Addr, Deps, Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::state::FlowType;

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Height {
    /// Previously known as "epoch"
    revision_number: Option<u64>,

    /// The height of a block
    revision_height: Option<u64>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct FungibleTokenData {
    denom: String,
    amount: Uint256,
    sender: Addr,
    receiver: Addr,
}

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

    pub fn channel_value(&self, _deps: Deps) -> Uint256 {
        // let balance = deps.querier.query_all_balances("address", self.data.denom);
        // deps.querier.sup
        return Uint256::from(125000000000011250_u128 * 2);
    }

    pub fn get_funds(&self) -> Uint256 {
        return self.data.amount;
    }

    fn local_channel(&self, direction: &FlowType) -> String {
        // Pick the appropriate channel depending on whether this is a send or a recv
        match direction {
            FlowType::In => self.destination_channel.clone(),
            FlowType::Out => self.source_channel.clone(),
        }
    }

    fn local_demom(&self) -> String {
        // This should actually convert the denom from the packet to the osmosis denom, but for now, just returning this
        return self.data.denom.clone();
    }

    pub fn path_data(&self, direction: &FlowType) -> (String, String) {
        let denom = self.local_demom();
        let channel = if denom.starts_with("transfer/") {
            // We should probably use the hash here, but need to figure out how to do that in cosmwasm
            self.local_channel(direction)
        } else {
            "any".to_string() // native tokens are rate limited globally
        };

        return (channel, denom);
    }
}

// Helpers

// Create a new packet for testing
#[macro_export]
macro_rules! test_msg_send {
    (channel_id: $channel_id:expr, denom: $denom:expr, channel_value: $channel_value:expr, funds: $funds:expr) => {
        crate::msg::SudoMsg::SendPacket {
            packet: crate::packet::Packet::mock($channel_id, $denom, $funds),
            local_denom: Some($denom),
            channel_value_hint: Some($channel_value),
        }
    };
}

#[macro_export]
macro_rules! test_msg_recv {
    (channel_id: $channel_id:expr, denom: $denom:expr, channel_value: $channel_value:expr, funds: $funds:expr) => {
        crate::msg::SudoMsg::RecvPacket {
            packet: crate::packet::Packet::mock($channel_id, $denom, $funds),
            local_denom: Some($denom),
            channel_value_hint: Some($channel_value),
        }
    };
}
