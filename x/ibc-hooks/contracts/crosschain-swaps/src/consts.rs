use phf::phf_map;

// Msg Reply IDs
pub const SWAP_REPLY_ID: u64 = 1u64;
pub const FORWARD_REPLY_ID: u64 = 2u64;

// IBC timeout
pub const PACKET_LIFETIME: u64 = 86400u64;

#[derive(Debug)]
pub struct ChainData<'a> {
    pub channel: &'a str,
    pub addr_prefix: &'a str,
}

// Known channels
pub const CHANNEL_MAP: phf::Map<&'static str, ChainData> = phf_map! {
    // This is only used for testing. We should remove it from the final build.
    // Can we add this conditionally using features?
    "osmosis-test" => ChainData{channel:"channel-0", addr_prefix: "osmo"},
    "osmosis-bad-test" => ChainData{channel:"channel-0", addr_prefix: "juno"},

    // These are the actual hard-coded channels
    "cosmoshub" => ChainData{channel:"channel-0", addr_prefix: "cosmos"},
    "juno" => ChainData{channel:"channel-42", addr_prefix: "juno"},
    "axelar" => ChainData{channel:"channel-208", addr_prefix: "axelar"},
};
