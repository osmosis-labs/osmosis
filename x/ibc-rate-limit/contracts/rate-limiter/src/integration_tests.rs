#[cfg(test)]
mod tests {
    use crate::helpers::RateLimitingContract;
    use crate::msg::{Channel, InstantiateMsg};
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

    fn proper_instantiate(channels: Vec<Channel>) -> (App, RateLimitingContract) {
        let mut app = mock_app();
        let cw_template_id = app.store_code(contract_template());

        let msg = InstantiateMsg {
            gov_module: Addr::unchecked(GOV_ADDR),
            ibc_module: Addr::unchecked(IBC_ADDR),
            channels,
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
        use crate::{
            msg::{Channel, ExecuteMsg, QuotaMsg},
            state::{RESET_TIME_DAILY, RESET_TIME_MONTHLY, RESET_TIME_WEEKLY},
        };

        #[test]
        fn expiration() {
            let quota = QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 10, 10);

            let (mut app, cw_template_contract) = proper_instantiate(vec![Channel {
                name: "channel".to_string(),
                quotas: vec![quota],
            }]);

            // Using all the allowance
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 3_000,
                funds: 300,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let res = app.execute(Addr::unchecked(IBC_ADDR), cosmos_msg).unwrap();

            let Attribute { key, value } = &res.custom_attrs(1)[2];
            assert_eq!(key, "weekly_used");
            assert_eq!(value, "300");
            let Attribute { key, value } = &res.custom_attrs(1)[3];
            assert_eq!(key, "weekly_max");
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
                b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
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
            assert_eq!(key, "weekly_used");
            assert_eq!(value, "300");
            let Attribute { key, value } = &res.custom_attrs(1)[3];
            assert_eq!(key, "weekly_max");
            assert_eq!(value, "300");
        }

        #[test]
        fn multiple_quotas() {
            let quotas = vec![
                QuotaMsg::new("daily", RESET_TIME_DAILY, 1, 1),
                QuotaMsg::new("weekly", RESET_TIME_WEEKLY, 5, 5),
                QuotaMsg::new("monthly", RESET_TIME_MONTHLY, 5, 5),
            ];

            let (mut app, cw_template_contract) = proper_instantiate(vec![Channel {
                name: "channel".to_string(),
                quotas,
            }]);

            // Sending 1% to use the daily allowance
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 100,
                funds: 1,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let res = app.execute(Addr::unchecked(IBC_ADDR), cosmos_msg).unwrap();

            println!("{res:?}");

            // Another packet is rate limited
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 100,
                funds: 1,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let _err = app
                .execute(Addr::unchecked(IBC_ADDR), cosmos_msg)
                .unwrap_err();

            // ... One day passes
            app.update_block(|b| {
                b.height += 10;
                b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
            });

            // Sending the packet should work now
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 100,
                funds: 1,
            };

            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            app.execute(Addr::unchecked(IBC_ADDR), cosmos_msg).unwrap();

            // Do that for 4 more days
            for _ in 1..4 {
                // ... One day passes
                app.update_block(|b| {
                    b.height += 10;
                    b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
                });

                // Sending the packet should work now
                let msg = ExecuteMsg::SendPacket {
                    channel_id: "channel".to_string(),
                    channel_value: 100,
                    funds: 1,
                };
                let cosmos_msg = cw_template_contract.call(msg).unwrap();
                app.execute(Addr::unchecked(IBC_ADDR), cosmos_msg).unwrap();
            }

            // ... One day passes
            app.update_block(|b| {
                b.height += 10;
                b.time = b.time.plus_seconds(RESET_TIME_DAILY + 1)
            });

            // We now have exceeded the weekly limit!  Even if the daily limit allows us, the weekly doesn't
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 100,
                funds: 1,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let _err = app
                .execute(Addr::unchecked(IBC_ADDR), cosmos_msg)
                .unwrap_err();

            // ... One week passes
            app.update_block(|b| {
                b.height += 10;
                b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
            });

            // We can still can't send because the weekly and monthly limits are the same
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 100,
                funds: 1,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let _err = app
                .execute(Addr::unchecked(IBC_ADDR), cosmos_msg)
                .unwrap_err();

            // Waiting a week again, doesn't help!!
            // ... One week passes
            app.update_block(|b| {
                b.height += 10;
                b.time = b.time.plus_seconds(RESET_TIME_WEEKLY + 1)
            });

            // We can still can't send because the  monthly limit hasn't passed
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 100,
                funds: 1,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let _err = app
                .execute(Addr::unchecked(IBC_ADDR), cosmos_msg)
                .unwrap_err();

            // Only after two more weeks we can send again
            app.update_block(|b| {
                b.height += 10;
                b.time = b.time.plus_seconds((RESET_TIME_WEEKLY * 2) + 1) // Two weeks
            });

            println!("{:?}", app.block_info());
            let msg = ExecuteMsg::SendPacket {
                channel_id: "channel".to_string(),
                channel_value: 100,
                funds: 1,
            };
            let cosmos_msg = cw_template_contract.call(msg).unwrap();
            let _err = app.execute(Addr::unchecked(IBC_ADDR), cosmos_msg).unwrap();
        }
    }
}
