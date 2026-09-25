use crate::state::flow::FlowType;
use cosmwasm_std::{Addr, Deps, StdError, Uint256};
use osmosis_std_derive::CosmwasmExt;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct Height {
    /// Previously known as "epoch"
    pub revision_number: Option<u64>,

    /// The height of a block
    pub revision_height: Option<u64>,
}

// IBC transfer data
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct FungibleTokenData {
    pub denom: String,
    pub amount: Uint256,
    pub sender: Addr,
    pub receiver: Addr,
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

pub(crate) fn hash_denom(denom: &str) -> String {
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

    /// Identifies this packet by the content the chain passes unchanged
    /// between the send authorisation and a later acknowledgement or timeout:
    /// port, channel, sender, receiver, denom, amount and timeouts. The
    /// sequence is left out because it is not assigned yet when the send is
    /// authorised, and the destination because the chain does not fill it in
    /// at that point either.
    pub fn send_key(&self) -> String {
        let mut hasher = Sha256::new();
        for part in [
            self.source_port.as_str(),
            self.source_channel.as_str(),
            self.data.sender.as_str(),
            self.data.receiver.as_str(),
            self.data.denom.as_str(),
        ] {
            hasher.update(part.as_bytes());
            hasher.update([0u8]);
        }
        hasher.update(self.data.amount.to_string().as_bytes());
        hasher.update([0u8]);
        let timeouts = format!(
            "{:?}/{:?}/{:?}",
            self.timeout_height.revision_number,
            self.timeout_height.revision_height,
            self.timeout_timestamp
        );
        hasher.update(timeouts.as_bytes());
        format!("{:x}", hasher.finalize())
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

    fn handle_denom_for_sends(&self) -> String {
        if !self.data.denom.starts_with("transfer/") {
            // For native tokens we just use what's on the packet
            return self.data.denom.clone();
        }
        // For non-native tokens, we need to generate the IBCDenom
        hash_denom(&self.data.denom)
    }

    fn handle_denom_for_recvs(&self) -> String {
        // A token is returning to this chain only when its trace starts with the
        // counterparty's port/channel followed by a slash. The slash is load-bearing:
        // ibc-go's ReceiverChainIsSource compares against "{port}/{channel}/", and
        // without it a source of channel-8 would also claim traces from channel-80
        // through channel-89 and channel-800 onwards, which belong to other chains.
        // Doing the strip once and branching on its result means the classification
        // and the unprefixing can never disagree.
        let voucher_prefix = format!("transfer/{}/", self.source_channel);
        match self.data.denom.strip_prefix(&voucher_prefix) {
            Some(unprefixed) => {
                // These are tokens that have been sent to the counterparty and are returning.
                // ibc-go's rule: if what remains still carries a port/channel prefix it is a
                // voucher this chain had already wrapped, so it lives here as ibc/HASH.
                // Anything else is a native denom (uosmo, factory/..., gamm/pool/N, ...)
                // and is used as-is.
                // The ibc-go implementation checks that the denom has been built correctly. We
                // don't need to do that here because if it hasn't, the transfer module will catch it.
                if unprefixed.starts_with("transfer/") {
                    hash_denom(unprefixed)
                } else {
                    unprefixed.to_string()
                }
            }
            None => {
                // Tokens that come directly from the counterparty (or from further away).
                // Since the sender didn't prefix them, we need to do it here.
                let channel = &self.destination_channel;
                let prefixed = format!("transfer/{}/{}", channel, self.data.denom);
                hash_denom(&prefixed)
            }
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
            "channel-17-local".to_string(),
            "channel-42-counterparty".to_string(),
            "uosmo".to_string(),
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
            "channel-17-local".to_string(),
            "channel-42-counterparty".to_string(),
            "transfer/channel-17-local/ujuno".to_string(),
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
            "channel-42-counterparty".to_string(), // The counterparty's channel is the source here
            "channel-17-local".to_string(),        // Our channel is the dest channel
            "ujuno".to_string(),                   // This is unwrapped. It is our job to wrap it
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
            "channel-42-counterparty".to_string(), // The counterparty's channel is the source here
            "channel-17-local".to_string(),        // Our channel is the dest channel
            "transfer/channel-42-counterparty/uosmo".to_string(),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uosmo");
    }

    #[test]
    fn receive_tokenfactory_token() {
        // The counterparty chain sends us back our native token that they had wrapped
        let packet = Packet::mock(
            "channel-42-counterparty".to_string(), // The counterparty's channel is the source here
            "channel-17-local".to_string(),        // Our channel is the dest channel
            "transfer/channel-42-counterparty/factory/osmo1em6xs47hd82806f5cxgyufguxrrc7l0aqx7nzzptjuqgswczk8csavdxek/alloyed/allUSDT".to_string(),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "factory/osmo1em6xs47hd82806f5cxgyufguxrrc7l0aqx7nzzptjuqgswczk8csavdxek/alloyed/allUSDT");
    }

    #[test]
    fn receive_native_lp_share() {
        // A native denom with slashes that is not a tokenfactory denom must still be
        // recognised as native when it returns (the old split-on-slash heuristic hashed it)
        let packet = Packet::mock(
            "channel-42-counterparty".to_string(),
            "channel-17-local".to_string(),
            "transfer/channel-42-counterparty/gamm/pool/1".to_string(),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "gamm/pool/1");
    }

    // Regression tests for the prefix collision reported in osmosis-labs/osmosis#9742.
    //
    // Injective's channel to Osmosis is channel-8 and Osmosis' side is channel-122.
    // Stride's stATOM lives on Injective as transfer/channel-89/stuatom. "channel-8"
    // is a string prefix of "channel-89", so a check without the trailing slash
    // misclassified the packet as an Osmosis-native token returning home.
    const INJECTIVE_TO_OSMOSIS: &str = "channel-8";
    const OSMOSIS_FROM_INJECTIVE: &str = "channel-122";
    const STATOM_ON_INJECTIVE_TRACE: &str = "transfer/channel-89/stuatom";

    #[test]
    fn receive_colliding_channel_prefix_is_foreign() {
        let packet = Packet::mock(
            INJECTIVE_TO_OSMOSIS.to_string(),
            OSMOSIS_FROM_INJECTIVE.to_string(),
            STATOM_ON_INJECTIVE_TRACE.to_string(),
            1_000_000_u128.into(),
        );
        // Osmosis must prefix its own channel and hash the full two-hop trace
        let expected = hash_denom(&format!(
            "transfer/{}/{}",
            OSMOSIS_FROM_INJECTIVE, STATOM_ON_INJECTIVE_TRACE
        ));
        assert_eq!(packet.local_denom(&FlowType::In), expected);
        assert!(!packet.local_denom(&FlowType::In).is_empty());
    }

    #[test]
    fn receive_exact_channel_prefix_is_returning_native() {
        // The same source channel with a genuinely returning native token still unwraps
        let packet = Packet::mock(
            INJECTIVE_TO_OSMOSIS.to_string(),
            OSMOSIS_FROM_INJECTIVE.to_string(),
            format!("transfer/{}/uosmo", INJECTIVE_TO_OSMOSIS),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uosmo");

        // and a returning tokenfactory denom is untouched
        let alloyed =
            "factory/osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3/alloyed/allBTC";
        let packet = Packet::mock(
            INJECTIVE_TO_OSMOSIS.to_string(),
            OSMOSIS_FROM_INJECTIVE.to_string(),
            format!("transfer/{}/{}", INJECTIVE_TO_OSMOSIS, alloyed),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), alloyed);
    }

    #[test]
    fn receive_exact_channel_prefix_returning_voucher_is_hashed() {
        // A voucher Osmosis had wrapped (ATOM from the Hub) returning from Injective
        let packet = Packet::mock(
            INJECTIVE_TO_OSMOSIS.to_string(),
            OSMOSIS_FROM_INJECTIVE.to_string(),
            format!(
                "transfer/{}/{}",
                INJECTIVE_TO_OSMOSIS, WRAPPED_ATOM_ON_OSMOSIS_TRACE
            ),
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::In),
            WRAPPED_ATOM_ON_OSMOSIS_HASH
        );
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
            "channel-141".to_string(), // from: hub
            "channel-0".to_string(),   // to: osmosis
            "uatom".to_string(),
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::In),
            WRAPPED_ATOM_ON_OSMOSIS_HASH
        );

        // uatom on osmosis sent back to the hub
        let packet = Packet::mock(
            "channel-0".to_string(),                   // from: osmosis
            "channel-141".to_string(),                 // to: hub
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uatom");

        // osmo sent to the hub
        let packet = Packet::mock(
            "channel-0".to_string(),   // from: osmosis
            "channel-141".to_string(), // to: hub
            "uosmo".to_string(),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::Out), "uosmo");

        // osmo on the hub sent back to osmosis
        // send
        let packet = Packet::mock(
            "channel-141".to_string(),             // from: hub
            "channel-0".to_string(),               // to: osmosis
            WRAPPED_OSMO_ON_HUB_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::Out), WRAPPED_OSMO_ON_HUB_HASH);

        // receive
        let packet = Packet::mock(
            "channel-141".to_string(),             // from: hub
            "channel-0".to_string(),               // to: osmosis
            WRAPPED_OSMO_ON_HUB_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uosmo");

        // Now let's pretend we're the hub.
        // The following tests are from perspective of the the hub (i.e.: if this contract were deployed there)
        //
        // osmo sent to the hub
        let packet = Packet::mock(
            "channel-0".to_string(),   // from: osmosis
            "channel-141".to_string(), // to: hub
            "uosmo".to_string(),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), WRAPPED_OSMO_ON_HUB_HASH);

        // uosmo on the hub sent back to the osmosis
        let packet = Packet::mock(
            "channel-141".to_string(),             // from: hub
            "channel-0".to_string(),               // to: osmosis
            WRAPPED_OSMO_ON_HUB_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::In), "uosmo");

        // uatom sent to osmosis
        let packet = Packet::mock(
            "channel-141".to_string(), // from: hub
            "channel-0".to_string(),   // to: osmosis
            "uatom".to_string(),
            0_u128.into(),
        );
        assert_eq!(packet.local_denom(&FlowType::Out), "uatom");

        // utaom on the osmosis sent back to the hub
        // send
        let packet = Packet::mock(
            "channel-0".to_string(),                   // from: osmosis
            "channel-141".to_string(),                 // to: hub
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::Out),
            WRAPPED_ATOM_ON_OSMOSIS_HASH
        );

        // receive
        let packet = Packet::mock(
            "channel-0".to_string(),                   // from: osmosis
            "channel-141".to_string(),                 // to: hub
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
            "channel-42".to_string(),                  // from: osmosis
            "channel-0".to_string(),                   // to: juno
            WRAPPED_ATOM_ON_OSMOSIS_TRACE.to_string(), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::Out),
            WRAPPED_ATOM_ON_OSMOSIS_HASH
        );

        // receive
        let packet = Packet::mock(
            "channel-42".to_string(), // from: osmosis
            "channel-0".to_string(),  // to: juno
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
            "channel-0".to_string(),  // from: juno
            "channel-42".to_string(), // to: osmosis
            format!("{}{}", "transfer/channel-0/", WRAPPED_ATOM_ON_OSMOSIS_TRACE), // unwrapped before reaching the contract
            0_u128.into(),
        );
        assert_eq!(
            packet.local_denom(&FlowType::Out),
            juno_wrapped_osmosis_wrapped_atom_hash
        );

        // receive
        let packet = Packet::mock(
            "channel-0".to_string(),  // from: juno
            "channel-42".to_string(), // to: osmosis
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
