use enum_repr::EnumRepr;

// Msg Reply IDs
#[EnumRepr(type = "u64")]
pub enum MsgReplyID {
    Swap = 1,
    Forward = 2,
}

// IBC timeout
pub const PACKET_LIFETIME: u64 = 86400u64;

// Callback key
pub const CALLBACK_KEY: &str = "ibc_callback";
