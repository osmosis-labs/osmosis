use cosmwasm_schema::cw_serde;

#[cw_serde]
pub struct Height {
    /// Previously known as "epoch"
    #[serde(skip_serializing_if = "Option::is_none")]
    revision_number: Option<u64>,

    /// The height of a block
    #[serde(skip_serializing_if = "Option::is_none")]
    revision_height: Option<u64>,
}

// An IBC packet
#[cw_serde]
pub struct Packet {
    pub sequence: u64,
    pub source_port: String,
    pub source_channel: String,
    pub destination_port: String,
    pub destination_channel: String,
    pub data: String, // FungibleTokenData
    pub timeout_height: Height,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timeout_timestamp: Option<u64>,
}

// The following are part of the wasm_hooks interface
#[cw_serde]
pub enum IBCAsyncOptions {
    #[serde(rename = "request_ack")]
    RequestAck {
        /// The source channel (osmosis side) of the IBC packet
        source_channel: String,
        /// The sequence number that the packet was sent with
        packet_sequence: u64,
    },
}

#[cw_serde]
pub struct OnRecvPacketAsyncResponse {
    pub is_async_ack: bool,
}

#[cw_serde]
pub struct ContractAck {
    pub contract_result: String,
    pub ibc_ack: String,
}


#[cw_serde]
#[serde(tag = "type", content = "content")]
pub enum IBCAck {
    AckResponse{
        packet: Packet,
        contract_ack: ContractAck,
    },
    AckError {
        packet: Packet,
        error_description: String,
        error_response: String,
    }
}

