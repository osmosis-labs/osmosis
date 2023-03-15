use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Coin, Deps, Timestamp};
use crosschain_swaps::ibc::MsgTransfer;
use itertools::Itertools;

pub use crate::error::RegistryError;
use crate::{
    helpers::{hash_denom_trace, DenomTrace, QueryDenomTraceRequest},
    msg::QueryMsg,
};
use std::convert::AsRef;

// IBC transfer port
const TRANSFER_PORT: &str = "transfer";
// IBC timeout
pub const PACKET_LIFETIME: u64 = 604_800u64; // One week in seconds

#[cw_serde]
pub struct Chain(String);
#[cw_serde]
pub struct ChannelId(String);

impl ChannelId {
    pub fn new(channel_id: &str) -> Result<Self, RegistryError> {
        if !ChannelId::validate(channel_id) {
            return Err(RegistryError::InvalidChannelId(channel_id.to_string()));
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

#[cw_serde]
pub struct ForwardingMemo {
    pub receiver: String,
    pub port: String,
    pub channel: ChannelId,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub next: Option<Box<Memo>>,
}

#[cw_serde]
pub struct Memo {
    forward: ForwardingMemo,
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
    pub fn new(deps: Deps<'a>, registry_contract: String) -> Result<Self, RegistryError> {
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

    /// Get a contract address by its alias
    /// Example: get_contract("registries") -> "osmo1..."
    pub fn get_contract(self, alias: String) -> Result<String, RegistryError> {
        self.deps
            .querier
            .query_wasm_smart(
                &self.registry_contract,
                &QueryMsg::GetAddressFromAlias {
                    contract_alias: alias.clone(),
                },
            )
            .map_err(|_e| RegistryError::AliasDoesNotExist { alias })
    }

    /// Get a the name of the chain connected via channel `via_channel` on chain `on_chain`.
    /// Example: get_connected_chain("osmosis", "channel-42") -> "juno"
    pub fn get_connected_chain(
        &self,
        on_chain: &str,
        via_channel: &str,
    ) -> Result<String, RegistryError> {
        self.deps
            .querier
            .query_wasm_smart(
                &self.registry_contract,
                &QueryMsg::GetDestinationChainFromSourceChainViaChannel {
                    on_chain: on_chain.to_string(),
                    via_channel: via_channel.to_string(),
                },
            )
            .map_err(|_e| RegistryError::ChannelToChainChainLinkDoesNotExist {
                channel_id: via_channel.to_string(),
                source_chain: on_chain.to_string(),
            })
    }

    /// Get the channel id for the channel connecting chain `on_chain` to chain `for_chain`.
    /// Example: get_channel("osmosis", "juno") -> "channel-42"
    /// Example: get_channel("juno", "osmosis") -> "channel-0"
    pub fn get_channel(&self, for_chain: &str, on_chain: &str) -> Result<String, RegistryError> {
        self.deps
            .querier
            .query_wasm_smart(
                &self.registry_contract,
                &QueryMsg::GetChannelFromChainPair {
                    source_chain: on_chain.to_string(),
                    destination_chain: for_chain.to_string(),
                },
            )
            .map_err(|_e| RegistryError::ChainChannelLinkDoesNotExist {
                source_chain: on_chain.to_string(),
                destination_chain: for_chain.to_string(),
            })
    }

    /// Re-encodes the bech32 address for the receiving chain
    /// Example: encode_addr_for_chain("osmo1...", "juno") -> "juno1..."
    pub fn encode_addr_for_chain(&self, addr: &str, chain: &str) -> Result<String, RegistryError> {
        let (_, data, variant) = bech32::decode(addr).map_err(|e| RegistryError::Bech32Error {
            action: "decoding".into(),
            addr: addr.into(),
            source: e,
        })?;

        let response: String = self.deps.querier.query_wasm_smart(
            &self.registry_contract,
            &QueryMsg::GetBech32PrefixFromChainName {
                chain_name: chain.to_string(),
            },
        )?;

        let receiver =
            bech32::encode(&response, data, variant).map_err(|e| RegistryError::Bech32Error {
                action: "encoding".into(),
                addr: addr.into(),
                source: e,
            })?;

        Ok(receiver)
    }

    /// Get the bech32 prefix for the given chain
    /// Example: get_bech32_prefix("osmosis") -> "osmo"
    pub fn get_bech32_prefix(&self, chain: &str) -> Result<String, RegistryError> {
        self.deps
            .api
            .debug(&format!("Getting prefix for chain: {chain}"));
        let prefix: String = self
            .deps
            .querier
            .query_wasm_smart(
                &self.registry_contract,
                &QueryMsg::GetBech32PrefixFromChainName {
                    chain_name: chain.to_string(),
                },
            )
            .map_err(|e| {
                self.deps.api.debug(&format!("Got error: {e}"));
                RegistryError::Bech32PrefixDoesNotExist {
                    chain: chain.into(),
                }
            })?;
        if prefix.is_empty() {
            return Err(RegistryError::Bech32PrefixDoesNotExist {
                chain: chain.into(),
            });
        }
        Ok(prefix)
    }

    /// Returns the IBC path the denom has taken to get to the current chain
    /// Example: unwrap_denom_path("ibc/0A...") -> [{"local_denom":"ibc/0A","on":"osmosis","via":"channel-17"},{"local_denom":"ibc/1B","on":"middle_chain","via":"channel-75"},{"local_denom":"token0","on":"source_chain","via":null}
    pub fn unwrap_denom_path(&self, denom: &str) -> Result<Vec<MultiHopDenom>, RegistryError> {
        self.deps.api.debug(&format!("Unwrapping denom {denom}"));
        // Check that the denom is an IBC denom
        if !denom.starts_with("ibc/") {
            return Err(RegistryError::InvalidIBCDenom {
                denom: denom.into(),
            });
        }

        // Get the denom trace
        let res = QueryDenomTraceRequest {
            hash: denom.to_string(),
        }
        .query(&self.deps.querier)?;

        let DenomTrace { path, base_denom } = match res.denom_trace {
            Some(denom_trace) => Ok(denom_trace),
            None => Err(RegistryError::NoDenomTrace {
                denom: denom.into(),
            }),
        }?;

        self.deps
            .api
            .debug(&format!("procesing denom trace {path}"));
        // Let's iterate over the parts of the denom trace and extract the
        // chain/channels into a more useful structure: MultiHopDenom
        let mut hops: Vec<MultiHopDenom> = vec![];
        let mut current_chain = "osmosis".to_string(); // The initial chain is always osmosis
        let mut rest: &str = &path;
        let parts = path.split('/');

        for chunk in &parts.chunks(2) {
            let Some((port, channel)) = chunk.take(2).collect_tuple() else {
                return Err(RegistryError::InvalidDenomTracePath{ path: path.clone(), denom: denom.into() });
            };

            // Check that the port is "transfer"
            if port != TRANSFER_PORT {
                return Err(RegistryError::InvalidTransferPort { port: port.into() });
            }

            // Check that the channel is valid
            let full_trace = rest.to_owned() + "/" + &base_denom;
            hops.push(MultiHopDenom {
                local_denom: hash_denom_trace(&full_trace),
                on: Chain(current_chain.clone().to_string()),
                via: Some(ChannelId::new(channel)?),
            });

            current_chain = self.get_connected_chain(&current_chain, channel)?;
            rest = rest
                .trim_start_matches(&format!("{port}/{channel}"))
                .trim_start_matches('/'); // hops other than first and last will have this slash
        }

        hops.push(MultiHopDenom {
            local_denom: base_denom,
            on: Chain(current_chain),
            via: None,
        });

        Ok(hops)
    }

    /// Returns an IBC MsgTransfer that with a packet forward middleware memo
    /// that will send the coin back to its original chain and then to the
    /// receiver in `into_chain`.
    ///
    /// If the receiver `into_chain` is not specified, we assume the receiver is
    /// the current chain (where the the registries are hosted and the denom
    /// original denom exists)
    ///
    /// `own_addr` must the the address of the contract that is calling this
    /// function.
    ///
    /// `block_time` is the current block time. This is needed to calculate the
    /// timeout timestamp.
    pub fn unwrap_coin_into(
        &self,
        coin: Coin,
        receiver: String,
        into_chain: Option<&str>,
        own_addr: String,
        block_time: Timestamp,
    ) -> Result<MsgTransfer, RegistryError> {
        let path = self.unwrap_denom_path(&coin.denom)?;
        self.deps
            .api
            .debug(&format!("Generating unwrap transfer message for: {path:?}"));
        if path.len() < 2 {
            return Err(RegistryError::InvalidMultiHopLengthMin {
                length: path.len(),
                min: 2,
            });
        }

        let MultiHopDenom {
            local_denom: _,
            on: first_chain,
            via: first_channel,
        } = path
            .first()
            .ok_or_else(|| RegistryError::InvalidDenomTracePath {
                path: format!("{:?}", path.clone()),
                denom: coin.denom.clone(),
            })?;

        let first_channel = match first_channel {
            Some(channel) => Ok(channel),
            None => Err(RegistryError::InvalidDenomTrace {
                error: "First hop must contain a channel".to_string(),
            }),
        }?;

        // default the receiver chain to the first chain if it isn't provided
        let receiver_chain = match into_chain {
            Some(chain) => chain,
            None => first_chain.as_ref(),
        };
        let receiver_chain: &str = &receiver_chain.to_lowercase();

        // validate the receiver matches the chain
        let receiver_prefix = self.get_bech32_prefix(receiver_chain)?;
        if receiver[..receiver_prefix.len()] != receiver_prefix {
            return Err(RegistryError::InvalidReceiverPrefix {
                receiver,
                chain: receiver_chain.into(),
            });
        }

        let ts = block_time.plus_seconds(PACKET_LIFETIME);
        let path_iter = path.iter().skip(1);

        let mut next: Option<Box<Memo>> = None;
        let mut prev_chain: &str = receiver_chain;

        for hop in path_iter.rev() {
            // If the last hop is the same as the receiver chain, we don't need
            // to forward anymore
            if hop.via.is_none() && hop.on.as_ref() == receiver_chain {
                continue;
            }

            // To unwrap we use the channel through which the token came, but once on the native
            // chain, we need to get the channel that connects that chain to the receiver.
            let channel = match &hop.via {
                Some(channel) => channel.to_owned(),
                None => ChannelId(self.get_channel(prev_chain, hop.on.as_ref())?),
            };

            next = Some(Box::new(Memo {
                forward: ForwardingMemo {
                    receiver: self.encode_addr_for_chain(&receiver, prev_chain)?,
                    port: TRANSFER_PORT.to_string(),
                    channel,
                    next,
                },
            }));
            prev_chain = hop.on.as_ref();
        }

        let memo =
            serde_json_wasm::to_string(&next).map_err(|e| RegistryError::SerialiaztionError {
                error: e.to_string(),
            })?;

        // encode the receiver address for the first chain
        let first_receiver = self.encode_addr_for_chain(&receiver, first_chain.as_ref())?;

        Ok(MsgTransfer {
            source_port: TRANSFER_PORT.to_string(),
            source_channel: first_channel.to_owned().as_ref().to_string(),
            token: Some(coin.into()),
            sender: own_addr,
            receiver: first_receiver,
            timeout_height: None,
            timeout_timestamp: Some(ts.nanos()),
            memo,
        })
    }
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_channel_id() {
        assert!(ChannelId::validate("channel-0"));
        assert!(ChannelId::validate("channel-1"));
        assert!(ChannelId::validate("channel-1234567890"));
        assert!(!ChannelId::validate("channel-"));
        assert!(!ChannelId::validate("channel-abc"));
        assert!(!ChannelId::validate("channel-1234567890a"));
        assert!(!ChannelId::validate("channel-1234567890-"));
        assert!(!ChannelId::validate("channel-1234567890-abc"));
        assert!(!ChannelId::validate("channel-1234567890-1234567890"));
    }

    #[test]
    fn test_forwarding_memo() {
        let memo = Memo {
            forward: ForwardingMemo {
                receiver: "receiver".to_string(),
                port: "port".to_string(),
                channel: ChannelId::new("channel-0").unwrap(),
                next: Some(Box::new(Memo {
                    forward: ForwardingMemo {
                        receiver: "receiver2".to_string(),
                        port: "port2".to_string(),
                        channel: ChannelId::new("channel-1").unwrap(),
                        next: None,
                    },
                })),
            },
        };
        let encoded = serde_json_wasm::to_string(&memo).unwrap();
        let decoded: Memo = serde_json_wasm::from_str(&encoded).unwrap();
        assert_eq!(memo, decoded);
        assert_eq!(
            encoded,
            r#"{"forward":{"receiver":"receiver","port":"port","channel":"channel-0","next":{"forward":{"receiver":"receiver2","port":"port2","channel":"channel-1"}}}}"#
        )
    }
}
