use cosmwasm_std::{DepsMut, Response};

use crate::state;
use crate::{
    state::{INFLIGHT_PACKETS, RECOVERY_STATES},
    ContractError,
};

// TODO: Deal with timeouts

/// Called by the chain when the ack for a packet that has configured this contract as its
/// callback has been received.
///
/// The chain needs to verify that the ack is valid ack for the packet with  the matching
/// source_channel and sequence before calling this function.
///
/// If the contract didn't send the IBC packet with (source_channel, sequence), we return a
/// success and no other changes are made.
///
/// If this contract sent the IBC packet, its data will be stored in
/// INFLIGHT_PACKETS. At this point the ack can be a success or a failure.
///
/// If it's a success, we remove the inflight packet and return. The packet will
/// no longer be tracked.
///
/// If it's a failure, the sent funds will have been returned to this contract.
/// We then store the amount and original sender on RECOVERY_STATES so that the
/// sender can recover the funds by calling execute::Recover{}.
pub fn receive_ack(
    deps: DepsMut,
    source_channel: String,
    sequence: u64,
    _ack: String,
    success: bool,
) -> Result<Response, ContractError> {
    // deps.api.debug(&format!(
    //     "received ack for packet {channel:?} {sequence:?}: {ack:?}, {success:?}"
    // ));
    let response = Response::new()
        .add_attribute("contract", "crosschain_swaps")
        .add_attribute("action", "receive_ack");

    // Check if there is an inflight packet for the received (channel, sequence)
    let sent_packet = INFLIGHT_PACKETS.may_load(deps.storage, (&source_channel, sequence))?;
    let Some(mut recovery) = sent_packet else {
        // If there isn't, continue
        return Ok(response.add_attribute("msg", "received unexpected ack"))
    };
    // Remove the in-flight packet
    INFLIGHT_PACKETS.remove(deps.storage, (&source_channel, sequence));

    if success {
        // If the acc is successful, there is nothing else to do and the crosschain swap has been completed
        return Ok(response.add_attribute("msg", "packet successfully delviered"));
    }

    // If the ack is a failure, we create a recovery for the original sender of the packet.
    let recovery_addr = recovery.recovery_addr.clone();
    RECOVERY_STATES.update(deps.storage, &recovery_addr, |recoveries| {
        // Since the recovery state and the in-flight packet store the same
        // data, we can just modify the status and store the object in the
        // RECOVERY_STATES map.
        recovery.status = state::ibc::Status::AckFailure;
        let Some(mut recoveries) = recoveries else {
            return Ok::<_, ContractError>(vec![recovery])
        };
        recoveries.push(recovery);
        Ok(recoveries)
    })?;

    Ok(response
        .add_attribute("msg", "recovery stored")
        .add_attribute("reecovery_addr", recovery_addr))
}
