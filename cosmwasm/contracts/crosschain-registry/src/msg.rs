use cosmwasm_schema::{cw_serde, QueryResponses};

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    /// Set a alias->address map in the registry
    SetContractAlias {
        /// The alias to be used for the contract
        contract_alias: String,
        /// The address of the contract
        contract_address: String,
    },
    /// Change an existing alias->address map in the registry
    ChangeContractAlias {
        /// The alias currently used by the contract
        current_contract_alias: String,
        /// The new alias to be used bythe contract
        new_contract_alias: String,
    },
    /// Remove an existing alias->address map in the registry
    RemoveContractAlias {
        /// The alias to be removed
        contract_alias: String,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(GetAddressFromAliasResponse)]
    GetAddressFromAlias { contract_alias: String },
}

// Response for GetAddressFromAlias query
#[cw_serde]
pub struct GetAddressFromAliasResponse {
    pub address: String,
}
