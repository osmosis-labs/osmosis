# Router

## Trade-offs To Re-evaluate

- Router skips found route if token OUT is found in the intermediary
path by calling `validateAndFilterRoutes` function
- Router skips found route if token IN is found in the intermediary
path by calling `validateAndFilterRoutes` function
- In the above 2 cases, we could exit early instead of continuing to search for such routes

## Future Problems

```json
{
  "amount_in": {
    "denom": "uosmo",
    "amount": "90000000000"
  },
  "amount_out": "61378463821",
  "route": [
    {
      "Route": {
        "Pools": [
          {
            "PoolI": {
              "underlying_pool": {
                "address": "osmo19e2mf7cywkv7zaug6nk5f87d07fxrdgrladvymh2gwv5crvm3vnsuewhh7",
                "id": 1,
                "pool_params": {
                  "swap_fee": "0.010000000000000000",
                  "exit_fee": "0.000000000000000000"
                },
                "future_pool_governor": "",
                "total_weight": "10737418240.000000000000000000",
                "total_shares": {
                  "denom": "gamm/pool/1",
                  "amount": "100000000000000000000"
                },
                "pool_assets": [
                  {
                    "token": {
                      "denom": "uosmo",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  },
                  {
                    "token": {
                      "denom": "uusdc",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  }
                ]
              },
              "sqs_model": {
                "total_value_locked_usdc": "1",
                "balances": [
                  {
                    "denom": "uosmo",
                    "amount": "100000000000"
                  },
                  {
                    "denom": "uusdc",
                    "amount": "100000000000"
                  }
                ]
              }
            },
            "token_out_denom": "uusdc"
          }
        ]
      },
      "out_amount": "30819785541",
      "in_amount": "45000000000"
    },
    {
      "Route": {
        "Pools": [
          {
            "PoolI": {
              "underlying_pool": {
                "address": "osmo1pjkt93g9lhntcpxk6pn04xwa87gf23wpjghjudql5p7n2exujh7szrdvtc",
                "id": 5,
                "pool_params": {
                  "swap_fee": "0.010000000000000000",
                  "exit_fee": "0.000000000000000000"
                },
                "future_pool_governor": "",
                "total_weight": "10737418240.000000000000000000",
                "total_shares": {
                  "denom": "gamm/pool/5",
                  "amount": "100000000000000000000"
                },
                "pool_assets": [
                  {
                    "token": {
                      "denom": "uion",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  },
                  {
                    "token": {
                      "denom": "uosmo",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  }
                ]
              },
              "sqs_model": {
                "total_value_locked_usdc": "1",
                "balances": [
                  {
                    "denom": "uion",
                    "amount": "100000000000"
                  },
                  {
                    "denom": "uosmo",
                    "amount": "100000000000"
                  }
                ]
              }
            },
            "token_out_denom": "uosmo"
          },
          {
            "PoolI": {
              "underlying_pool": {
                "address": "osmo1ad4r3uh5pdn5pgg5hnl6u5utfeqmpwstlvgvg2h2jdztrcnwkqgs3hs85z",
                "id": 4,
                "pool_params": {
                  "swap_fee": "0.010000000000000000",
                  "exit_fee": "0.000000000000000000"
                },
                "future_pool_governor": "",
                "total_weight": "10737418240.000000000000000000",
                "total_shares": {
                  "denom": "gamm/pool/4",
                  "amount": "100000000000000000000"
                },
                "pool_assets": [
                  {
                    "token": {
                      "denom": "uosmo",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  },
                  {
                    "token": {
                      "denom": "uusdc",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  }
                ]
              },
              "sqs_model": {
                "total_value_locked_usdc": "1",
                "balances": [
                  {
                    "denom": "uosmo",
                    "amount": "100000000000"
                  },
                  {
                    "denom": "uusdc",
                    "amount": "100000000000"
                  }
                ]
              }
            },
            "token_out_denom": "uusdc"
          }
        ]
      },
      "out_amount": "15279339140",
      "in_amount": "22500000000"
    },
    {
      "Route": {
        "Pools": [
          {
            "PoolI": {
              "underlying_pool": {
                "address": "osmo1pjkt93g9lhntcpxk6pn04xwa87gf23wpjghjudql5p7n2exujh7szrdvtc",
                "id": 5,
                "pool_params": {
                  "swap_fee": "0.010000000000000000",
                  "exit_fee": "0.000000000000000000"
                },
                "future_pool_governor": "",
                "total_weight": "10737418240.000000000000000000",
                "total_shares": {
                  "denom": "gamm/pool/5",
                  "amount": "100000000000000000000"
                },
                "pool_assets": [
                  {
                    "token": {
                      "denom": "uion",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  },
                  {
                    "token": {
                      "denom": "uosmo",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  }
                ]
              },
              "sqs_model": {
                "total_value_locked_usdc": "1",
                "balances": [
                  {
                    "denom": "uion",
                    "amount": "100000000000"
                  },
                  {
                    "denom": "uosmo",
                    "amount": "100000000000"
                  }
                ]
              }
            },
            "token_out_denom": "uosmo"
          },
          {
            "PoolI": {
              "underlying_pool": {
                "address": "osmo1rumch2tw3vq7dqatv2e4q389vkrwp3t6pkaphx5hdtpc9yyjah9swxheaq",
                "id": 7,
                "pool_params": {
                  "swap_fee": "0.010000000000000000",
                  "exit_fee": "0.000000000000000000"
                },
                "future_pool_governor": "",
                "total_weight": "10737418240.000000000000000000",
                "total_shares": {
                  "denom": "gamm/pool/7",
                  "amount": "100000000000000000000"
                },
                "pool_assets": [
                  {
                    "token": {
                      "denom": "uosmo",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  },
                  {
                    "token": {
                      "denom": "uusdc",
                      "amount": "100000000000"
                    },
                    "weight": "5368709120"
                  }
                ]
              },
              "sqs_model": {
                "total_value_locked_usdc": "1",
                "balances": [
                  {
                    "denom": "uosmo",
                    "amount": "100000000000"
                  },
                  {
                    "denom": "uusdc",
                    "amount": "100000000000"
                  }
                ]
              }
            },
            "token_out_denom": "uusdc"
          }
        ]
      },
      "out_amount": "15279339140",
      "in_amount": "22500000000"
    }
  ]
}
```
- Note that there are 2 routes that start with pool ID 5. In the client side, we end up filtering them out by uniquer pool IDs.
If we don't, the client does not have the pool model updated. As a result, the quotes are invalid.
However, it is still optimal to split. We need to refactor the general algorithm so that instead of separate routes
we have a prefix tree data structure. Then, as we iterate the prefix tree, we can determine the optimal split and not
have to worry about duplicate pool IDs.

