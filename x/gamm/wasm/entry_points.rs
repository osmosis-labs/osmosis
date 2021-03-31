macro_rules! use_std {
    () => {
        use super::$contract;
        use cosmwasm_std::{do_execute, do_instantiate, do_query, DepsMut, Env, MessageInfo, QueryResponse, Response, StdResult};
    }
}

#[macro_export]
macro_rules! create_amm_entry_points {
    (@swappable; $contract:ident) => {
        #[no_mangle]
        pub fn execute_internal(
            deps: DepsMut,
            _env: Env,
            _info: MessageInfo,
            msg: ExecuteMsg
        ) -> StdResult<Response> {
            ExecuteMsg::Swap{
                token_in, token_in_max, token_out, token_out_max, max_spot_price,
            } => {
                let pool_in = pools(deps.storage).
                $contract::swap(
                    token_in, token_in_max, token_out, token_out_max, max_spot_price,
                );
            }
        }

        #[no_mangle]
        pub extern "C" fn execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
            do_execute(&execute_internal, env_ptr, info_ptr, msg_ptr)
        }
    }

    (@queryable; $contract:ident) => {
        #[no_mangle]
        pub fn query_internal(
            deps: DepsMut,
            _env: Env,
            msg: ExecuteMsg
        ) -> StdResult<Response> {
            QueryMsg::SpotPrice{} => $contract::spot_price();
            QueryMsg::InGivenOut{} => $contract::in_given_out();
            QueryMsg::OutGivenIn{} => $contract::out_given_in();
        }

        #[no_mangle]
        pub extern "C" fn query(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
            do_query(query_internal, env_ptr, msg_ptr)
        }
    }

    ($contract:ident) => {
        mod wasm {
            use_std!();
            $crate::create_amm_entry_points!(@swappable; $contract);
            $crate::create_amm_entry_points!(@queryable; $contract);
        }
    }
}
#[macro_export]
macro_rules! create_amm_swap_entry_points {
    ($contract:ident) => {
        mod wasm {
            use_std!();
            $crate::create_amm_entry_points!(@swappable; $contract);
        }
    }
}
#[macro_export]
macro_rules! create_amm_querier_entry_points {
    ($contract:ident) => {
        mod wasm {
            use_std!();
            $crate::create_amm_entry_points!(@queryable; $contract);
        }
    }
}
