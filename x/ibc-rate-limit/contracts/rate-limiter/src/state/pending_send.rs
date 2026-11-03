use cosmwasm_std::{Timestamp, Uint256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::packet::Packet;

/// Clock skew allowed between this chain and a counterparty when deciding
/// that a packet's timeout timestamp has passed.
pub const TIMEOUT_SKEW_SECONDS: u64 = 60 * 60;

/// How long a record is kept past its last window end when the packet has no
/// timeout timestamp (height-only timeout), since the contract cannot tell
/// when a counterparty height will be reached. Long enough that holding a
/// packet's acknowledgement back for the whole period, to have it settle a
/// later identical send once the record is gone, means locking the sent
/// amount up for a year.
pub const HEIGHT_TIMEOUT_RETENTION_SECONDS: u64 = 365 * 24 * 60 * 60;

/// Outer bound on any record's retention past its last window end, so that
/// a timeout timestamp far in the future cannot make a record permanent. The
/// same reasoning as for the height-only retention applies: a failure held
/// back this long to settle a later identical send costs its sender ten
/// years of locked capital.
pub const MAX_RETENTION_SECONDS: u64 = 10 * 365 * 24 * 60 * 60;

/// A window a send was counted in: the quota's name, the window's end and
/// the flow's generation at the time. The generation is `None` for records
/// written before it was stored, which then match on name and end alone.
pub type WindowId = (String, Timestamp, Option<u64>);

/// A send that passed the rate limit and has not been settled yet.
///
/// It records, for every quota the send was counted against, the window it
/// was counted in. A send that later fails (error acknowledgement or timeout)
/// is refunded to a quota only while that quota is still in the same window,
/// so a refund can only ever restore what the same window charged.
///
/// Records are keyed by packet content (Packet::send_key), which is not a
/// unique identity: the same content can be sent again as long as the
/// packet's timeout has not passed on the counterparty. One record exists per
/// content key; identical sends counted in the same windows share it and
/// `count` says how many are outstanding, so that every one of them can be
/// refunded if it fails. A record is retained until both its windows have
/// ended and the packet's timeout has passed by the clock-skew margin, and an
/// identical send that would be counted in a different window while the
/// record is retained is rejected, so within the retention bounds a
/// settlement cannot credit a window other than the one that was charged.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct PendingSend {
    pub funds: Uint256,
    /// (quota name, period_end) for each quota on the (channel, denom) path
    pub channel_windows: Vec<(String, Timestamp)>,
    /// (quota name, period_end) for each quota on the ("any", denom) path
    pub any_windows: Vec<(String, Timestamp)>,
    /// When the record may be removed without settlement, see `retention_end`
    pub retain_until: Timestamp,
    /// How many identical sends this record stands for. Each settlement
    /// refunds `funds` once and lowers it by one; the record goes at zero.
    /// Records written by the first 0.2.0 build (on the testnet) lack this
    /// field and the generations below.
    #[serde(default = "one")]
    pub count: u32,
    /// The flow generation of each entry in `channel_windows`, in order.
    /// Empty on records written before generations were stored.
    #[serde(default)]
    pub channel_generations: Vec<u64>,
    /// The flow generation of each entry in `any_windows`, in order.
    #[serde(default)]
    pub any_generations: Vec<u64>,
}

fn one() -> u32 {
    1
}

impl PendingSend {
    /// When a record for `packet` may be dropped unsettled: after the last
    /// window it was counted in has ended, and after the packet's timeout
    /// timestamp has passed, each plus the clock-skew margin. A height-only
    /// timeout gets a fixed long retention instead. A timeout far in the
    /// future keeps the record for as long, up to the outer bound, because a
    /// failure relayed after the record is gone would settle a newer
    /// identical send; an acknowledgement can be relayed at any time.
    pub fn retention_end(latest_period_end: Timestamp, packet: &Packet) -> Timestamp {
        let cap = saturating_plus_seconds(latest_period_end, MAX_RETENTION_SECONDS);
        let until = match packet.timeout_timestamp {
            Some(timeout) if timeout > 0 => {
                let timeout = Timestamp::from_nanos(timeout);
                let latest = if timeout > latest_period_end {
                    timeout
                } else {
                    latest_period_end
                };
                saturating_plus_seconds(latest, TIMEOUT_SKEW_SECONDS)
            }
            _ => saturating_plus_seconds(latest_period_end, HEIGHT_TIMEOUT_RETENTION_SECONDS),
        };
        if until > cap {
            cap
        } else {
            until
        }
    }

    /// Whether the record may be removed without settlement.
    pub fn is_expired(&self, now: Timestamp) -> bool {
        self.retain_until < now
    }

    /// The windows on the (channel, denom) path this record was counted in
    pub fn channel_ids(&self) -> Vec<WindowId> {
        zip_ids(&self.channel_windows, &self.channel_generations)
    }

    /// The windows on the ("any", denom) path this record was counted in
    pub fn any_ids(&self) -> Vec<WindowId> {
        zip_ids(&self.any_windows, &self.any_generations)
    }
}

fn zip_ids(windows: &[(String, Timestamp)], generations: &[u64]) -> Vec<WindowId> {
    windows
        .iter()
        .enumerate()
        .map(|(i, (name, period_end))| (name.clone(), *period_end, generations.get(i).copied()))
        .collect()
}

/// Splits window ids into the two vectors a record stores them as
pub fn split_ids(ids: Vec<(String, Timestamp, u64)>) -> (Vec<(String, Timestamp)>, Vec<u64>) {
    ids.into_iter()
        .map(|(name, period_end, generation)| ((name, period_end), generation))
        .unzip()
}

/// Whether recorded windows and the windows a transfer would be counted in
/// are the same, entry by entry. A recorded entry without a generation
/// matches on name and end alone.
pub fn same_windows(recorded: &[WindowId], upcoming: &[(String, Timestamp, u64)]) -> bool {
    recorded.len() == upcoming.len()
        && recorded.iter().zip(upcoming).all(|(r, u)| {
            r.0 == u.0
                && r.1 == u.1
                && match r.2 {
                    Some(generation) => generation == u.2,
                    None => true,
                }
        })
}

/// `Timestamp::plus_seconds` aborts on overflow; a far-future timeout is
/// caller-supplied data and must not be able to do that.
fn saturating_plus_seconds(timestamp: Timestamp, seconds: u64) -> Timestamp {
    Timestamp::from_nanos(
        timestamp
            .nanos()
            .saturating_add(seconds.saturating_mul(1_000_000_000)),
    )
}
