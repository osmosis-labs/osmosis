use crate::execute;
use crate::ContractError;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::testing::{mock_dependencies, MockApi, MockQuerier, MockStorage};
use cosmwasm_std::{to_binary, Addr, CosmosMsg, OwnedDeps, StdResult, WasmMsg};

use crate::msg::ExecuteMsg;

/// CwTemplateContract is a wrapper around Addr that provides a lot of helpers
/// for working with this.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct CwTemplateContract(pub Addr);

impl CwTemplateContract {
    pub fn addr(&self) -> Addr {
        self.0.clone()
    }

    pub fn call<T: Into<ExecuteMsg>>(&self, msg: T) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Execute {
            contract_addr: self.addr().into(),
            msg,
            funds: vec![],
        }
        .into())
    }
}

pub fn setup() -> Result<OwnedDeps<MockStorage, MockApi, MockQuerier>, ContractError> {
    let mut deps = mock_dependencies();
    execute::set_contract_alias(
        deps.as_mut(),
        "contract_one".to_string(),
        "osmo1dfaselkjh32hnkljw3nlklk2lknmes".to_string(),
    )?;
    execute::set_contract_alias(
        deps.as_mut(),
        "contract_two".to_string(),
        "osmo1dfg4k3jhlknlfkjdslkjkl43klnfdl".to_string(),
    )?;
    execute::set_contract_alias(
        deps.as_mut(),
        "contract_three".to_string(),
        "osmo1dfgjlk4lkfklkld32fsdajknjrrgfg".to_string(),
    )?;
    Ok(deps)
}
