use crate::types::Pubkey;
use cosmwasm_schema::cw_serde;
use cosmwasm_std::{entry_point, Binary, DepsMut, Env, Response, StdError, StdResult};
use cw_storage_plus::{Item, Map};
use osmosis_authenticators as oa;
use sylvia::contract;
use sylvia::types::{InstantiateCtx, QueryCtx};

pub struct CosignerAuthenticator {
    pub(crate) pubkeys: Item<'static, Vec<Pubkey>>,
    pub(crate) named_pubkeys: Map<'static, String, Binary>,
}

#[contract]
#[sv::override_entry_point(sudo=sudo(SudoMsg))]
impl CosignerAuthenticator {
    pub const fn new() -> Self {
        CosignerAuthenticator {
            pubkeys: Item::new("pubkeys"),
            named_pubkeys: Map::new("named_pubkeys"),
        }
    }

    #[msg(instantiate)]
    pub fn instantiate(&self, ctx: InstantiateCtx, pubkeys: Vec<Pubkey>) -> StdResult<Response> {
        self.pubkeys.save(ctx.deps.storage, &pubkeys)?;
        Ok(Response::default())
    }

    #[msg(query)]
    pub fn pubkeys(&self, ctx: QueryCtx) -> StdResult<crate::types::PubkeysResponse> {
        let pubkeys = self.pubkeys.load(ctx.deps.storage)?;
        Ok(crate::types::PubkeysResponse { pubkeys })
    }

    pub fn sudo_authenticate(
        &self,
        deps: DepsMut,
        auth_request: oa::AuthenticationRequest,
    ) -> Result<Response, StdError> {
        let signatures: Vec<Binary> = cosmwasm_std::from_json(&auth_request.signature)?;
        if signatures.len() != self.pubkeys.load(deps.storage)?.len() {
            return Ok(Response::new().set_data(oa::AuthenticationResult::NotAuthenticated {}));
        }

        // The message hash is what gets signed
        let hash = oa::sha256(&auth_request.sign_mode_tx_data.sign_mode_direct);
        for (i, pubkey) in self.pubkeys.load(deps.storage)?.iter().enumerate() {
            // Fetch the pubkey binary from the name map if necessary
            let raw = match pubkey {
                Pubkey::ByName(name) => self
                    .named_pubkeys
                    .load(deps.storage, name.to_string())
                    .or_else(|_| {
                        Err(StdError::generic_err(format!("Pubkey {} not found", name)))
                    })?,
                Pubkey::Raw(raw) => raw.clone(),
            };

            // Verify signature i
            let valid = deps
                .api
                .secp256k1_verify(&hash, &signatures[i], &raw)
                .or_else(|e| {
                    deps.api.debug(&format!("error {:?}", e));
                    Err(StdError::generic_err("Failed to verify signature"))
                })?;

            if !valid {
                return Ok(Response::new().set_data(oa::AuthenticationResult::NotAuthenticated {}));
            }
        }

        Ok(Response::new().set_data(oa::AuthenticationResult::Authenticated {}))
    }
}

#[cw_serde]
pub enum SudoMsg {
    Authenticate(oa::AuthenticationRequest),
}

#[entry_point]
pub fn sudo(
    deps: DepsMut,
    _env: Env,
    SudoMsg::Authenticate(auth_request): SudoMsg,
) -> Result<Response, StdError> {
    CosignerAuthenticator::new().sudo_authenticate(deps, auth_request)
}
