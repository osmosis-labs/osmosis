#[macro_export]
macro_rules! create_entry_points {
    ($contract:ident) => {
        mod wasm {
            use super::$contract{
                swap, calc_spot_price,
                InstantiateMsg, ExecuteMsg, QueryMsg
            };
            use cosmwasm_std::{
                do_execute, do_instantiate, do_migrate, do_query,
                DepsMut, Env, MessageInfo, Response, QueryResponse, StdError,
            };


            #[no_mangle]
            fn _instantiate(deps: DepsMut, env: Env, info: MessageInfo, msg: InstantiateMsg) -> Result<Response, StdError>

            #[no_mangle]
            extern "C" fn instantiate(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
                do_instantiate(&_instantiate, env_ptr, info_ptr, msg_ptr)
            }

            fn _execute(deps: DepsMut, env: Env, info: MessageInfo, msg: ExecuteMsg) -> Result<Response, StdError> {
                // Swap is for internal module interface. Pool ID is already known by calling a
                // specific contract.
                // TODO: define default ExecuteMsg
                ExecuteMsg::Swap{
                    token_in, token_in_max, token_out, token_out_max, max_spot_price
                } => {
                    let record_in = record(deps.storage, token_in.denom);
                    let record_out = record(deps.storage, token_out.denom);
                    // TODO: generalize contraints
                    swap(record_in, token_in_max, record_out, token_out_max, max_spot_price)
                }
            }

            #[no_mangle]
            extern "C" fn execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32 {
                do_execute(&_execute, env_ptr, info_ptr, msg_ptr)
            }

            fn _query(deps: Deps, env: Env, msg: QueryMsg) -> Result<QueryResponse, StdError> {
                QueryMsg::SpotPrice{token_in_denom, token_out_denom} => {
                    // TODO: params
                    let record_in = record_read(deps.storage, token_in_denom);
                    let record_out = record_read(deps.storage, token_out_denom);
                    spot_price(record_in, record_out);
                }
                QueryMsg::OutGivenIn{token_in_denom, token_out_denom, token_in_amount} => {
                    // TODO: params
                    let record_in = record_read(deps.storage, token_in_denom);
                    let record_out = record_read(deps.storage, token_out_denom);
                    out_given_in(record_in, record_out, token_in_amount);
                }
                QueryMsg::InGivenOut{token_in_denom, token_out_denom, token_out_amount} => {
                    let record_in = record_read(deps.storage, token_in_denom);
                    let record_out = record_read(deps.storage, token_out_denom);
                    in_given_out(record_in, record_out, token_out_amount);
                }
            }

            #[no_mangle]
            extern "C" fn query(env_ptr: u32, msg_ptr: u32) -> u32 {
                do_query(&_query, env_ptr, msg_ptr)
            }
        }
    }
}
