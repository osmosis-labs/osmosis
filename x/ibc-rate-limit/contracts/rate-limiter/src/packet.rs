use cosmwasm_std::{Addr, Deps, Timestamp};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
pub struct Height {
    /// Previously known as "epoch"
    revision_number: Option<u64>,

    /// The height of a block
    revision_height: Option<u64>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
pub struct FungibleTokenData {
    denom: String,
    amount: u128,
    sender: Addr,
    receiver: Addr,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
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
    pub fn channel_value(&self, _deps: Deps) -> u128 {
        // let balance = deps.querier.query_all_balances("address", self.data.denom);
        // deps.querier.sup
        return 125000000000011250 * 2;
    }

    pub fn get_funds(&self) -> u128 {
        return self.data.amount;
    }

    fn local_channel(&self) -> String {
        // Pick the appropriate channel depending on whether this is a send or a recv
        return self.destination_channel.clone();
    }

    fn local_demom(&self) -> String {
        // This should actually convert the denom from the packet to the osmosis denom, but for now, just returning this
        return self.data.denom.clone();
    }

    pub fn path_data(&self) -> (String, String) {
        let denom = self.local_demom();
        let channel = if denom.starts_with("ibc/") {
            self.local_channel()
        } else {
            "any".to_string() // native tokens are rate limited globally
        };

        return (channel, denom);
    }
}
