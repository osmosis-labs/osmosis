#[macro_export]
macro_rules! create_amm_entry_points {
    ($contract:ident) => {
        mod wasm {
            use super::$contract;
            use cosmwasm_std::{do_execute, do_instantiate, do_migrate, do_query};

            #[no_mangle]
            extern "C" fn instantiate(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
                do_instantiate(&$contract::instantiate, env_ptr, info_ptr, msg_ptr)
            }

            #[no_mangle]
            fn execute_internal(
                deps: DepsMut,
                _env: Env,
                _info: MessageInfo,
                msg: ExecuteMsg
            ) -> StdResult<Response> {
                ExecuteMsg::Swap{
                    token_in, token_in_max, token_out, token_out_max, max_spot_price,
                } => {
                    let record_in = record(deps.storage, token_in.denom);
                    let record_out = record(deps.storage, token_out.denom);
                    // TODO: expand in_max into generic constraints
                    $contract::swap(
                        record_in, record_in_max, record_out, record_out_max, max_spot_price,
                    )
                }
            }

            #[no_mangle]
            extern "C" fn execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
                do_execute(&execute_internal, env_ptr, info_ptr, msg_ptr)
            }

            #[no_mangle]
            fn query_internal(
                deps: DepsMut,
                _env: Env,
                msg: QueryMsg
            ) -> StdResult<Response> {
                QueryMsg::SpotPrice{} => $contract::spot_price();
                QueryMsg::InGivenOut{} => $contract::in_given_out();
                QueryMsg::OutGivenIn{} => $contract::out_given_in();
            }

            #[no_mangle]
            extern "C" fn query(env_ptr: u32, msg_ptr: u32) -> u32 {
                do_query(&query_internal, env_ptr, msg_ptr)
            }
        }
    }
}
