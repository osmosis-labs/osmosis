use ibc::core::ics04_channel::packet::Packet as IBCPacket;
use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
pub struct Packet(pub IBCPacket);

impl Packet {
    pub fn channel_value(&self) -> u128 {
        todo!()
    }

    pub fn get_funds(&self) -> u128 {
        todo!()
    }

    pub fn path_data(&self) -> (String, String) {
        todo!()
    }
}
