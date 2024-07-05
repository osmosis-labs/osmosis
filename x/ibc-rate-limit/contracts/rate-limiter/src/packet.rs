use crate::state::flow::FlowType;
use cosmwasm_std::{Addr, Deps, StdError, Uint256};
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
    pub timeout_timestamp: Option<u64>,
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
    pub fn mock(
        source_channel: String,
        dest_channel: String,
        denom: String,
        funds: Uint256,
    ) -> Packet {
        Packet {
            sequence: 0,
            source_port: "transfer".to_string(),
            source_channel,
            destination_port: "transfer".to_string(),
            destination_channel: dest_channel,
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
            packet: $crate::packet::Packet::mock($channel_id, $channel_id, $denom, $funds),
            channel_value_mock: Some($channel_value),
        }
    };
}

#[cfg(test)]
#[macro_export]
macro_rules! test_msg_recv {
    (channel_id: $channel_id:expr, denom: $denom:expr, channel_value: $channel_value:expr, funds: $funds:expr) => {
        $crate::msg::SudoMsg::RecvPacket {
            packet: $crate::packet::Packet::mock(
                $channel_id,
                $channel_id,
                format!("transfer/{}/{}", $channel_id, $denom),
                $funds,
            ),
            channel_value_mock: Some($channel_value),
        }
    };
}

#[cfg(test)]
pub mod tests {
    use crate::msg::SudoMsg;

    use super::*;

    #[test]
    fn send_native() {
        let packet = Packet::mock(
            format!("channel-17-local"),
            format!("channel-42-counterparty"),
            format!("uosmo"),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::Out), "uosmo");
    }

    #[test]
    fn send_non_native() {
        // The transfer module "unhashes" the denom from
        // ibc/09E4864A262249507925831FBAD69DAD08F66FAAA0640714E765912A0751289A
        // to port/channel/denom before passing it along to the contrace
        let packet = Packet::mock(
            format!("channel-17-local"),
            format!("channel-42-counterparty"),
            format!("transfer/channel-17-local/ujuno"),
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::Out),
            "ibc/09E4864A262249507925831FBAD69DAD08F66FAAA0640714E765912A0751289A"
        );
    }

    #[test]
    fn receive_non_native() {
        // The counterparty chain sends their own native token to us
        let packet = Packet::mock(
            format!("channel-42-counterparty"), // The counterparty's channel is the source here
            format!("channel-17-local"),        // Our channel is the dest channel
            format!("ujuno"),                   // This is unwrapped. It is our job to wrap it
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::In),
            "ibc/09E4864A262249507925831FBAD69DAD08F66FAAA0640714E765912A0751289A"
        );
    }

    #[test]
    fn receive_native() {
        // The counterparty chain sends us back our native token that they had wrapped
        let packet = Packet::mock(
            format!("channel-42-counterparty"), // The counterparty's channel is the source here
            format!("channel-17-local"),        // Our channel is the dest channel
            format!("transfer/channel-42-counterparty/uosmo"),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uosmo");
    }

    // Let's assume we have two chains A and B (local and counterparty) connected in the following way:
    //
    // Chain A <---> channel-17-local <---> channel-42-counterparty <---> Chain B
    //
    // The following tests should pass
    //

    const WRAPPED_OSMO_ON_HUB_TRACE: &str = "transfer/channel-141/uosmo";
    const WRAPPED_ATOM_ON_OSMOSIS_TRACE: &str = "transfer/channel-0/uatom";
    const WRAPPED_ATOM_ON_OSMOSIS_HASH: &str =
        "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2";
    const WRAPPED_OSMO_ON_HUB_HASH: &str =
        "ibc/14F9BC3E44B8A9C1BE1FB08980FAB87034C9905EF17CF2F5008FC085218811CC";

    #[test]
    fn sanity_check() {
        // Examples using the official channels as of Nov 2022.

        // uatom sent to osmosis
        let packet = Packet::mock(
            format!("channel-141"), // from: hub
            format!("channel-0"),   // to: osmosis
            format!("uatom"),
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::In),
            WRAPPED_ATOM_ON_OSMOSIS_HASH.clone()
        );

        // uatom on osmosis sent back to the hub
        let packet = Packet::mock(
            format!("channel-0"),                      // from: osmosis
            format!("channel-141"),                    // to: hub
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uatom");

        // osmo sent to the hub
        let packet = Packet::mock(
            format!("channel-0"),   // from: osmosis
            format!("channel-141"), // to: hub
            format!("uosmo"),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::Out), "uosmo");

        // osmo on the hub sent back to osmosis
        // send
        let packet = Packet::mock(
            format!("channel-141"),                // from: hub
            format!("channel-0"),                  // to: osmosis
            WRAPPED_OSMO_ON_HUB_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::Out), WRAPPED_OSMO_ON_HUB_HASH);

        // receive
        let packet = Packet::mock(
            format!("channel-141"),                // from: hub
            format!("channel-0"),                  // to: osmosis
            WRAPPED_OSMO_ON_HUB_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uosmo");

        // Now let's pretend we're the hub.
        // The following tests are from perspective of the the hub (i.e.: if this contract were deployed there)
        //
        // osmo sent to the hub
        let packet = Packet::mock(
            format!("channel-0"),   // from: osmosis
            format!("channel-141"), // to: hub
            format!("uosmo"),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), WRAPPED_OSMO_ON_HUB_HASH);

        // uosmo on the hub sent back to the osmosis
        let packet = Packet::mock(
            format!("channel-141"),                // from: hub
            format!("channel-0"),                  // to: osmosis
            WRAPPED_OSMO_ON_HUB_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uosmo");

        // uatom sent to osmosis
        let packet = Packet::mock(
            format!("channel-141"), // from: hub
            format!("channel-0"),   // to: osmosis
            format!("uatom"),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::Out), "uatom");

        // utaom on the osmosis sent back to the hub
        // send
        let packet = Packet::mock(
            format!("channel-0"),                      // from: osmosis
            format!("channel-141"),                    // to: hub
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::Out),
            WRAPPED_ATOM_ON_OSMOSIS_HASH
        );

        // receive
        let packet = Packet::mock(
            format!("channel-0"),                      // from: osmosis
            format!("channel-141"),                    // to: hub
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uatom");
    }

    #[test]
    fn sanity_double() {
        // Now let's deal with double wrapping

        let juno_wrapped_osmosis_wrapped_atom_hash =
            "ibc/6CDD4663F2F09CD62285E2D45891FC149A3568E316CE3EBBE201A71A78A69388";

        // Send uatom on stored on osmosis to juno
        // send
        let packet = Packet::mock(
            format!("channel-42"),                     // from: osmosis
            format!("channel-0"),                      // to: juno
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::Out),
            WRAPPED_ATOM_ON_OSMOSIS_HASH
        );

        // receive
        let packet = Packet::mock(
            format!("channel-42"), // from: osmosis
            format!("channel-0"),  // to: juno
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(),
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::In),
            juno_wrapped_osmosis_wrapped_atom_hash
        );

        // Send back that multi-wrapped token to osmosis
        // send
        let packet = Packet::mock(
            format!("channel-0"),  // from: juno
            format!("channel-42"), // to: osmosis
            format!("{}{}", "transfer/channel-0/", WRAPPED_ATOM_ON_OSMOSIS_TRACE), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::Out),
            juno_wrapped_osmosis_wrapped_atom_hash
        );

        // receive
        let packet = Packet::mock(
            format!("channel-0"),  // from: juno
            format!("channel-42"), // to: osmosis
            format!("{}{}", "transfer/channel-0/", WRAPPED_ATOM_ON_OSMOSIS_TRACE), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::In),
            WRAPPED_ATOM_ON_OSMOSIS_HASH
        );
    }

    #[test]
    fn tokenfactory_packet() {
        let json = r#"{"send_packet":{"packet":{"sequence":4,"source_port":"transfer","source_channel":"channel-0","destination_port":"transfer","destination_channel":"channel-1491","data":{"denom":"transfer/channel-0/factory/osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj/czar","amount":"100000000000000000","sender":"osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks","receiver":"osmo1c584m4lq25h83yp6ag8hh4htjr92d954vklzja"},"timeout_height":{},"timeout_timestamp":1668024476848430980}}}"#;
        let parsed: SudoMsg = serde_json_wasm::from_str(json).unwrap();
        //println!("{parsed:?}");

        match parsed {
            SudoMsg::SendPacket { packet, .. } => {
                assert_eq!(
                    packet.local_denom(&FlowType::Out),
                    "ibc/07A1508F49D0753EDF95FA18CA38C0D6974867D793EB36F13A2AF1A5BB148B22"
                );
            }
            _ => panic!("parsed into wrong variant"),
        }
    }

    #[test]
    fn packet_with_memo() {
        // extra fields (like memo) get ignored.
        let json = r#"{"recv_packet":{"packet":{"sequence":1,"source_port":"transfer","source_channel":"channel-0","destination_port":"transfer","destination_channel":"channel-0","data":{"denom":"stake","amount":"1","sender":"osmo177uaalkhra6wth6hc9hu79f72eq903kwcusx4r","receiver":"osmo1fj6yt4pwfea4865z763fvhwktlpe020ef93dlq","memo":"some info"},"timeout_height":{"revision_height":100}}}}"#;
        let _parsed: SudoMsg = serde_json_wasm::from_str(json).unwrap();
        //println!("{parsed:?}");
    }
}
