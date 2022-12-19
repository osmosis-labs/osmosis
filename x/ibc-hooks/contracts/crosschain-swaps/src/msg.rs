use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Addr, Coin};
use swaprouter::msg::Slippage;

/// Message type for `instantiate` entry_point
#[cw_serde]
pub struct InstantiateMsg {
    /// This should be an instance of the Osmosis swaprouter contract
    pub swap_contract: String,
    /// If set to true, the contract will add a callback request on the packet
    /// so that it gets notified when an ack is received or if the packet timed
    /// out. If set to false, any funds sent on a packet that fails after a swap
    /// will be stuck in this contract.
    ///
    /// The information about the packet sender and recovery address is still
    /// stored, so recovery could be possible after a contract upgrade.
    pub track_ibc_sends: Option<bool>,
    /// These are the channels that will be accepted by the contract. This is
    /// needed to avoid sending packets to addresses not supported by the
    /// receiving chain. The channels are specified as (bech32_prefix, channel_id)
    pub channels: Vec<(String, String)>,
}

#[cw_serde]
pub struct Recovery {
    /// An osmosis addres used to recover any tokens that get stuck in the
    /// contract due to IBC failures
    pub recovery_addr: Addr,
}

/// Message type for `execute` entry_point
#[cw_serde]
pub enum ExecuteMsg {
    /// Execute a swap and forward it to the receiver address on the specified ibc channel
    OsmosisSwap {
        input_coin: Coin,
        output_denom: String,
        receiver: Addr,
        slippage: Slippage,
        next_memo: Option<String>,
        failed_delivery: Option<Recovery>,
    },
    /// Executing a recover will transfer any recoverable tokens that the sender
    /// has in this contract to its account.
    ///
    /// This is only usable if the contract is configured with track_ibc_sends.
    ///
    /// The only tokens that are considered recoverable for a "sender" are those
    /// returned by an IBC transfer sent by this contract, that are known to
    /// have failed, and that originated with a message specifying the "sender"
    /// as its recovery address.
    Recover {},
}

/// Message type for `query` entry_point
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// Returns the list of transfers that are recoverable for an Addr
    #[returns(Vec<crate::state::IBCTransfer>)]
    Recoverable { addr: Addr },
}

// tmp structure for crosschain response
#[cw_serde]
pub struct CrosschainSwapResponse {
    pub msg: String, // Do we want to provide more detailed information here?
}

/// Message type for `migrate` entry_point
#[cw_serde]
pub enum MigrateMsg {}

/// Message type for `sudo` entry_point
#[cw_serde]
pub enum SudoMsg {
    ReceivePacket {},
    ReceiveAck {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        sequence: u64,
        /// String encoded version of the ack as seen by OnAcknowledgementPacket(..)
        ack: String,
        /// Wether an ack is a success of failure according to the transfer spec
        success: bool,
    },
    ReceiveTimeout {},
}
