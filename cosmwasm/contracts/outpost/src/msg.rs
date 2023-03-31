use cosmwasm_schema::{cw_serde, QueryResponses};

/// Message type for `instantiate` entry_point
#[cw_serde]
pub struct InstantiateMsg {
    pub osmosis_channel: String,
    pub crosschain_swaps_contract: String,
}

/// Information about which contract to call
#[cw_serde]
pub struct WasmHookExecute {
    pub wasm: Wasm,
}

#[cw_serde]
pub struct Wasm {
    pub contract: String,
    pub msg: crosschain_swaps::ExecuteMsg,
}

/// Message type for `execute` entry_point
#[cw_serde]
pub enum ExecuteMsg {
    /// Execute a swap and forward it to the receiver address on the specified ibc channel
    OsmosisSwap {
        /// The final denom to be received (as represented on osmosis)
        output_denom: String,
        /// The receiver of the IBC packet to be sent after the swap
        receiver: String,
        /// Slippage for the swap
        slippage: swaprouter::Slippage,
        /// If for any reason the swap were to fail, users can specify a
        /// "recovery address" that can clain the funds on osmosis after a
        /// confirmed failure.
        on_failed_delivery: crosschain_swaps::FailedDeliveryAction,

        // Optional
        /// Execute a contract when the crosschain swaps has finished.
        /// This is only avaibale on chains that support wasm hooks
        #[cfg(feature = "callbacks")]
        callback: Option<crosschain_swaps::msg::Callback>,
    },
}

/// Message type for `migrate` entry_point
#[cw_serde]
pub enum MigrateMsg {}

/// Message type for `query` entry_point
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    // This example query variant indicates that any client can query the contract
    // using `YourQuery` and it will return `YourQueryResponse`
    // This `returns` information will be included in contract's schema
    // which is used for client code generation.
    //
    // #[returns(YourQueryResponse)]
    // YourQuery {},
}

// We define a custom struct for each query response
// #[cw_serde]
// pub struct YourQueryResponse {}
