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

// #[cw_serde]
// pub struct FungibleTokenData {
//     pub denom: String,
//     amount: Uint256,
//     sender: String,
//     receiver: String,
// }

// An IBC packet
#[cw_serde]
pub struct Packet {
    pub sequence: u64,
    pub source_port: String,
    pub source_channel: String,
    pub destination_port: String,
    pub destination_channel: String,
    pub data: String, // FungiibleTokenData
    pub timeout_height: Height,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timeout_timestamp: Option<u64>,
}

#[cw_serde]
pub enum IBCAsync {
    #[serde(rename = "request_ack")]
    RequestAck {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        packet_sequence: u64,
    },
}

#[cw_serde]
pub struct ContractAck {
    pub contract_result: String,
    pub ibc_ack: String,
}

#[cw_serde]
pub struct IBCAckResponse {
    pub packet: Packet,
    pub contract_ack: ContractAck,
}
