pub mod ibc;
pub mod state;
use cosmwasm_schema::cw_serde;
#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, DepsMut, Env, MessageInfo, Response, StdError};
use ibc::{ContractAck, IBCAckResponse, IBCAsync, Packet};
use osmosis_std_derive::CosmwasmExt;
use state::INFLIGHT_PACKETS;

// Messages
#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    Echo {
        msg: String,
    },
    Async {
        use_async: bool,
    },
    #[serde(rename = "force_emit_ibc_ack")]
    ForceEmitIBCAck {
        packet: Packet,
        channel: String,
    },
}

/// Message type for `sudo` entry_point
#[cw_serde]
pub enum SudoMsg {
    #[serde(rename = "ibc_async")]
    IBCAsync(IBCAsync),
}

// Instantiate
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, StdError> {
    Ok(Response::new())
}

// Execute
fn simple_response(msg: String) -> Response {
    Response::new()
        .add_attribute("echo", msg)
        .set_data(b"this should echo")
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
#[proto_message(type_url = "/osmosis.ibchooks.MsgEmitIBCAck")]
pub struct MsgEmitIBCAck {
    #[prost(string, tag = "1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    pub packet_sequence: u64,
    #[prost(string, tag = "3")]
    pub channel: ::prost::alloc::string::String,
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, StdError> {
    match msg {
        ExecuteMsg::Echo { msg } => Ok(simple_response(msg)),
        ExecuteMsg::Async { use_async } => {
            if use_async {
                Ok(Response::new().add_attribute("ibc_async_ack", "true"))
            } else {
                Ok(Response::default())
            }
        }
        ExecuteMsg::ForceEmitIBCAck { packet, channel } => {
            INFLIGHT_PACKETS.save(
                deps.storage,
                (&packet.destination_channel, packet.sequence),
                &packet,
            )?;
            let msg = MsgEmitIBCAck {
                sender: env.contract.address.to_string(),
                packet_sequence: packet.sequence,
                channel,
            };
            Ok(Response::new().add_message(msg))
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(deps: DepsMut, _env: Env, msg: SudoMsg) -> Result<Response, StdError> {
    match msg {
        SudoMsg::IBCAsync(IBCAsync::RequestAck {
            channel,
            packet_sequence,
        }) => Ok(Response::new().set_data(to_binary(&IBCAckResponse {
            packet: INFLIGHT_PACKETS.load(deps.storage, (&channel, packet_sequence))?,
            contract_ack: ContractAck {
                contract_result: base64::encode("success"),
                ibc_ack: base64::encode("ack"),
            },
        })?)),
    }
}
