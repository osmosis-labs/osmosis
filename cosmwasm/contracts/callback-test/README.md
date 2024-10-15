# callback-test

This contract is a sample contract which is used to test the x/callback module functionalty.

The implementation is based on the sample counter contract and has been modified to have callback msg in the Sudo entrypoint.

The following changes have been made

```rust
// msg.rs

#[cw_serde]
pub enum SudoMsg {
    Callback { 
        job_id: u64
    },
    Error {
        module_name: String,
        error_code: u32,
        contract_address: String,
        input_payload: String,
        error_message: String,
    },
}
```

```rust
//contract.rs

pub mod sudo {
    use super::*;
    use std::u64;

    pub fn handle_callback(deps: DepsMut, job_id: u64) -> Result<Response, ContractError> {
        STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
            if job_id == 0 {
                state.count -= 1; // Decrement the count
            };
            if job_id == 1 {
                state.count += 1; // Increment the count
            };
            if job_id == 2 {
                return Err(ContractError::SomeError {}); // Throw an error
            }
            // else do nothing
            Ok(state)
        })?;

        Ok(Response::new().add_attribute("action", "handle_callback"))
    }

    pub fn handle_error(deps: DepsMut, module_name: String, error_code: u32, _contract_address: String, _input_payload: String, _error_message: String) -> Result<Response, ContractError> {
        STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
            if module_name == "callback" && error_code == 2 {
                state.count = 0; // reset the counter
            }
            Ok(state)
        })?;

        Ok(Response::new().add_attribute("action", "handle_error"))
    }
}
```

Relevant test has been added as well in contract.rs and the default counter init/execute tests removed
