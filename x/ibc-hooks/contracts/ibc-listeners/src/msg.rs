use core::fmt;

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Addr;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde] // Important. If you ever change this. Make sure to change UnsubscribeAll as well
pub enum EventType {
    Acknowledgement,
    Timeout,
}

impl fmt::Display for EventType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            EventType::Acknowledgement => write!(f, "ack"),
            EventType::Timeout => write!(f, "timeout"),
        }
    }
}

#[cw_serde]
pub enum ExecuteMsg {
    Subscribe {
        channel: String,
        sequence: u64,
        event: EventType,
    },
}

#[cw_serde]
pub enum SudoMsg {
    UnSubscribeAll { channel: String, sequence: u64 },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Vec<Addr>)]
    Listeners {
        channel: String,
        sequence: u64,
        event: EventType,
    },
}

#[cw_serde]
pub enum MigrateMsg {}
