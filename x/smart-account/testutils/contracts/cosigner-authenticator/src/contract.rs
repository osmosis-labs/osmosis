use crate::types::Signature;
use cosmwasm_schema::cw_serde;
use cosmwasm_std::{entry_point, Binary, DepsMut, Env, Response, StdError, StdResult};
use cw_storage_plus::Item;
use osmosis_authenticators as oa;
use sylvia::types::{InstantiateCtx, QueryCtx};
use sylvia::{contract, entry_points};

pub struct CosignerAuthenticator {
    pub(crate) pubkeys: Item<'static, Vec<Binary>>,
}

#[entry_points]
#[contract]
#[sv::override_entry_point(sudo=sudo(SudoMsg))]
impl CosignerAuthenticator {
    const fn new() -> Self {
        CosignerAuthenticator {
            pubkeys: Item::new("pubkeys"),
        }
    }

    #[msg(instantiate)]
    fn instantiate(&self, ctx: InstantiateCtx, pubkeys: Vec<Binary>) -> StdResult<Response> {
        self.pubkeys.save(ctx.deps.storage, &pubkeys)?;
        Ok(Response::default())
    }

    #[msg(query)]
    fn pubkeys(&self, ctx: QueryCtx) -> StdResult<crate::types::PubkeysResponse> {
        let pubkeys = self.pubkeys.load(ctx.deps.storage)?;
        Ok(crate::types::PubkeysResponse { pubkeys })
    }

    fn sudo_authenticate(
        &self,
        deps: DepsMut,
        auth_request: oa::AuthenticationRequest,
    ) -> Result<Response, StdError> {
        deps.api.debug(&format!("auth_request {:?}", auth_request));
        let sigs: Vec<Signature> = cosmwasm_std::from_json(&auth_request.signature)?;

        if sigs.len() != self.pubkeys.load(deps.storage)?.len() {
            return Ok(Response::new().set_data(oa::AuthenticationResult::NotAuthenticated {}));
        }

        let mut pubkeys = self.pubkeys.load(deps.storage)?;
        pubkeys.push(auth_request.authenticator_params.clone().unwrap());

        // The message hash is what gets signed
        for (i, pubkey) in pubkeys.iter().enumerate() {
            let hash = oa::sha256(&concat(
                &auth_request.sign_mode_tx_data.sign_mode_direct,
                &sigs[i].salt,
            ));
            // Verify signature i
            let valid = deps
                .api
                .secp256k1_verify(&hash, &sigs[i].signature, &pubkey)
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

fn concat(a: &Binary, b: &Binary) -> Binary {
    let mut combined = a.to_vec();
    combined.extend(b.as_slice());
    Binary(combined)
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
