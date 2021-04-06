macro_rules! use_std {
    () => {
        use super::$contract;
        use cosmwasm_std::{do_execute, do_instantiate, do_query, DepsMut, Env, MessageInfo, QueryResponse, Response, StdResult};
    }
}

#[macro_export]
macro_rules! create_amm_entry_points {
    (@swappable; $contract:ident, true) => {
        #[no_mangle]
        pub fn execute_internal(
            deps: DepsMut,
            _env: Env,
            _info: MessageInfo,
            msg: ExecuteMsg
        ) -> StdResult<Response> {
            ExecuteMsg::Swap{
                pool_id, token_in, token_in_max, token_out, token_out_max, max_spot_price,
            } => {
                let pool = pool(deps.storage, pool_id);
                $contract::swap(
                    pool,
                    token_in, token_in_max, token_out, token_out_max, max_spot_price,
                );
            }
        }

        #[no_mangle]
        pub extern "C" fn execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
            do_execute(&execute_internal, env_ptr, info_ptr, msg_ptr)
        }
    }

    (@swappable; $contract:ident, false) => {
        #[no_mangle]
        pub extern "C" fn execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
            do_execute(&$contract::execute, env_ptr, info_ptr, msg_ptr)
        }
    }

    (@queryable; $contract:ident, true) => {
        #[no_mangle]
        pub fn query_internal(
            deps: DepsMut,
            _env: Env,
            msg: ExecuteMsg
        ) -> StdResult<Response> {
            match msg {
                QueryMsg::SpotPrice{} => $contract::spot_price(msg);
                QueryMsg::InGivenOut{} => $contract::in_given_out(msg);
                QueryMsg::OutGivenIn{} => $contract::out_given_in(msg);
            }
        }

        #[no_mangle]
        pub extern "C" fn query(env_ptr: u32, msg_ptr: u32) -> u32 {
            do_query(&query_internal, env_ptr, msg_ptr)
        }
    }

    (@queryable; $contract:ident, false) => {
        #[no_mangle]
        pub extern "C" fn query(env_ptr: u32, msg_ptr: u32) -> u32 {
            do_query(&$contract::query, env_ptr, msg_ptr)
        }
    }

    // use auto filling by default
    ($contract:ident) => {
        mod wasm {
            use_std!();
            $crate::create_amm_entry_points!(@swappable; $contract, true);
            $crate::create_amm_entry_points!(@queryable; $contract, true);
        }
    }
}

#[macro_export]
macro_rules! create_amm_manual_entry_points {
    ($contract:ident) => {
        mod wasm {
            use_std!();
            $crate::create_amm_entry_points!(@swappable; $contract, false);
            $crate::create_amm_entry_points!(@queryable; $contract, false);
        }
    }
}

/*
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
*/
