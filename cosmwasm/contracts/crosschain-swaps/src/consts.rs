use enum_repr::EnumRepr;

// Msg Reply IDs
#[EnumRepr(type = "u64")]
pub enum MsgReplyID {
    Swap = 1,
    Forward = 2,
}

// Callback key
pub const CALLBACK_KEY: &str = "ibc_callback";
