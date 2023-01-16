use osmosis_std::types::osmosis::tokenfactory::v1beta1::{
    MsgBurn, MsgBurnResponse, MsgChangeAdmin, MsgChangeAdminResponse, MsgCreateDenom,
    MsgCreateDenomResponse, MsgMint, MsgMintResponse, MsgSetDenomMetadata,
    MsgSetDenomMetadataResponse, QueryDenomAuthorityMetadataRequest,
    QueryDenomAuthorityMetadataResponse, QueryDenomsFromCreatorRequest,
    QueryDenomsFromCreatorResponse, QueryParamsRequest, QueryParamsResponse,
};

use crate::module::Module;
use crate::runner::Runner;
use crate::{fn_execute, fn_query};

pub struct TokenFactory<'a, R: Runner<'a>> {
    runner: &'a R,
}

impl<'a, R: Runner<'a>> Module<'a, R> for TokenFactory<'a, R> {
    fn new(runner: &'a R) -> Self {
        Self { runner }
    }
}

impl<'a, R> TokenFactory<'a, R>
where
    R: Runner<'a>,
{
    fn_execute! {
        pub create_denom: MsgCreateDenom => MsgCreateDenomResponse
    }

    fn_execute! {
        pub mint: MsgMint => MsgMintResponse
    }

    fn_execute! {
        pub burn: MsgBurn => MsgBurnResponse
    }

    fn_execute! {
        pub change_admin: MsgChangeAdmin => MsgChangeAdminResponse
    }

    fn_execute! {
        pub set_denom_metadata: MsgSetDenomMetadata => MsgSetDenomMetadataResponse
    }

    fn_query! {
        pub query_params ["/osmosis.tokenfactory.v1beta1.Query/Params"]: QueryParamsRequest => QueryParamsResponse
    }

    fn_query! {
        pub query_denom_authority_metadata ["/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata"]: QueryDenomAuthorityMetadataRequest => QueryDenomAuthorityMetadataResponse
    }

    fn_query! {
        pub query_denoms_from_creator ["/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator"]: QueryDenomsFromCreatorRequest => QueryDenomsFromCreatorResponse
    }
}

#[cfg(test)]
mod tests {
    use cosmrs::proto::cosmos::bank::v1beta1::QueryBalanceRequest;
    use cosmwasm_std::Coin;
    use osmosis_std::types::osmosis::tokenfactory::v1beta1::{
        MsgBurn, MsgChangeAdmin, MsgCreateDenom, MsgMint, QueryDenomAuthorityMetadataRequest,
        QueryDenomsFromCreatorRequest,
    };

    use crate::{Account, Bank, Module, OsmosisTestApp, TokenFactory};

    #[test]
    fn tokenfactory_integration() {
        let app = OsmosisTestApp::new();
        let signer = app.init_account(&[Coin::new(10000000, "uosmo")]).unwrap();
        let tokenfactory = TokenFactory::new(&app);
        let bank = Bank::new(&app);

        // create denom
        let subdenom = "udenom";
        let denom = tokenfactory
            .create_denom(
                MsgCreateDenom {
                    sender: signer.address(),
                    subdenom: subdenom.to_owned(),
                },
                &signer,
            )
            .unwrap()
            .data
            .new_token_denom;

        assert_eq!(format!("factory/{}/{}", signer.address(), subdenom), denom);

        // denom from creator
        let denoms = tokenfactory
            .query_denoms_from_creator(&QueryDenomsFromCreatorRequest {
                creator: signer.address(),
            })
            .unwrap()
            .denoms;

        assert_eq!(denoms, [denom.clone()]);

        // mint
        let coin: osmosis_std::types::cosmos::base::v1beta1::Coin =
            Coin::new(1000000000, denom.clone()).into();
        tokenfactory
            .mint(
                MsgMint {
                    sender: signer.address(),
                    amount: Some(coin.clone()),
                },
                &signer,
            )
            .unwrap();

        let balance = bank
            .query_balance(&QueryBalanceRequest {
                address: signer.address(),
                denom: denom.clone(),
            })
            .unwrap()
            .balance
            .unwrap();

        assert_eq!(coin.amount, balance.amount);
        assert_eq!(coin.denom, balance.denom);

        // burn
        tokenfactory
            .burn(
                MsgBurn {
                    sender: signer.address(),
                    amount: Some(coin.clone()),
                },
                &signer,
            )
            .unwrap();

        let balance = bank
            .query_balance(&QueryBalanceRequest {
                address: signer.address(),
                denom: denom.clone(),
            })
            .unwrap()
            .balance
            .unwrap();

        assert_eq!("0", balance.amount);
        assert_eq!(coin.denom, balance.denom);

        // change admin
        let new_admin = app.init_account(&[]).unwrap();
        tokenfactory
            .change_admin(
                MsgChangeAdmin {
                    sender: signer.address(),
                    denom: denom.clone(),
                    new_admin: new_admin.address(),
                },
                &signer,
            )
            .unwrap();

        // query denom auth
        let auth_metadata = tokenfactory
            .query_denom_authority_metadata(&QueryDenomAuthorityMetadataRequest { denom })
            .unwrap()
            .authority_metadata
            .unwrap();

        assert_eq!(new_admin.address(), auth_metadata.admin);
    }
}
