use cosmwasm_std::{Addr, Timestamp};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cw_storage_plus::{Item, Map};

pub enum FlowType {
    In,
    Out,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema, Copy)]
pub struct Flow {
    pub inflow: u128,
    pub outflow: u128,
    pub period_end: Timestamp,
}

impl Flow {
    pub fn new(inflow: impl Into<u128>, outflow: impl Into<u128>) -> Self {
        Self {
            inflow: inflow.into(),
            outflow: outflow.into(),
            period_end: Timestamp::from_nanos(1),
        }
    }

    pub fn add_flow(&mut self, direction: FlowType, value: u128) {
        match direction {
            FlowType::In => self.inflow = self.inflow.saturating_add(value),
            FlowType::Out => self.outflow = self.outflow.saturating_add(value),
        }
    }

    pub fn volume(&self) -> u128 {
        self.inflow.abs_diff(self.outflow)
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Quota {
    max_percentage: u32,
}

impl Quota {
    pub fn apply(&self, total_value: &u128) -> u128 {
        total_value * (self.max_percentage as u128) / 100_u128
    }
}

impl From<u32> for Quota {
    fn from(max_percentage: u32) -> Self {
        if max_percentage > 100 {
            Quota {
                max_percentage: 100,
            }
        } else {
            Quota { max_percentage }
        }
    }
}

pub const IBCMODULE: Item<Addr> = Item::new("ibc_module");
// For simplicity, the map keys (ibc channel) refers to the "host" channel on the
// osmosis side. This means that on PacketSend it will refer to the source
// channel while on PacketRecv it refers to the destination channel.
//
// It is the responsibility of the go module to pass the appropriate channel
// when sending the messages
pub const QUOTA: Map<String, Quota> = Map::new("quota");
pub const FLOW: Map<String, Flow> = Map::new("flow");
