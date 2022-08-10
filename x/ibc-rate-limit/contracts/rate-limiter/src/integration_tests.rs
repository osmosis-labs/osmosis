#[cfg(test)]
mod tests {
    use crate::helpers::RateLimitingContract;
    use crate::msg::InstantiateMsg;
    use cosmwasm_std::{Addr, Coin, Empty, Uint128};
    use cw_multi_test::{App, AppBuilder, Contract, ContractWrapper, Executor};

    pub fn contract_template() -> Box<dyn Contract<Empty>> {
        let contract = ContractWrapper::new(
            crate::contract::execute,
            crate::contract::instantiate,
            crate::contract::query,
        );
        Box::new(contract)
    }

    const USER: &str = "USER";
    const IBC_ADDR: &str = "IBC_MODULE";
    const GOV_ADDR: &str = "GOV_MODULE";
    const NATIVE_DENOM: &str = "nosmo";

    fn mock_app() -> App {
        AppBuilder::new().build(|router, _, storage| {
            router
                .bank
                .init_balance(
                    storage,
                    &Addr::unchecked(USER),
                    vec![Coin {
                        denom: NATIVE_DENOM.to_string(),
                        amount: Uint128::new(1_000),
                    }],
                )
                .unwrap();
        })
    }

    fn proper_instantiate(channel_quotas: Vec<(String, u32)>) -> (App, RateLimitingContract) {
        let mut app = mock_app();
        let cw_template_id = app.store_code(contract_template());

        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            channel_quotas,
        };

        let cw_template_contract_addr = app
            .instantiate_contract(
                cw_template_id,
                Addr::unchecked(GOV_ADDR),
                &msg,
                &[],
                "test",
                None,
            )
            .unwrap();

        let cw_template_contract = RateLimitingContract(cw_template_contract_addr);

        (app, cw_template_contract)
    }

    mod expiration {
        use cosmwasm_std::Attribute;

        use super::*;
        use crate::{msg::ExecuteMsg, state::RESET_TIME};

        #[test]
        fn expiration() {
            let (mut app, cw_template_contract) =
                proper_instantiate(vec![("channel".to_string(), 10)]);

            // Using all the allowance
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 3_000,
                funds: 300,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let res = app.execute(Addr::unchecked(IBC_ADDR), cosmos_msg).unwrap();

            let Attribute { key, value } = &res.custom_attrs(1)[2];
            assert_eq!(key, "used");
            assert_eq!(value, "300");
            let Attribute { key, value } = &res.custom_attrs(1)[3];
            assert_eq!(key, "max");
            assert_eq!(value, "300");

            // Another packet is rate limited
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 3_000,
                funds: 300,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let err = app
                .execute(Addr::unchecked(IBC_ADDR), cosmos_msg)
                .unwrap_err();

            // TODO: how do we check the error type here?
            println!("{err:?}");

            // ... Time passes
            app.update_block(|b| {
                b.height += 1000;
                b.time = b.time.plus_seconds(RESET_TIME + 1)
            });

            // Sending the packet should work now
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 3_000,
                funds: 300,
            };

            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let res = app.execute(Addr::unchecked(IBC_ADDR), cosmos_msg).unwrap();

            let Attribute { key, value } = &res.custom_attrs(1)[2];
            assert_eq!(key, "used");
            assert_eq!(value, "300");
            let Attribute { key, value } = &res.custom_attrs(1)[3];
            assert_eq!(key, "max");
            assert_eq!(value, "300");
        }
    }
}
