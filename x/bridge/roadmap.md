# Roadmap

## Goals

| Item | Goals                                                                        | link                                                                                                                | Status             |
|------|------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|--------------------|
| 1    | Designed the bridging architecture, including module and observer parts      | [link](https://github.com/chadury2021/Osmosis-Bifrost/blob/master/bridge-docs/bridge/README.md)                     | :heavy_check_mark: |
| 2    | Implemented x/bridge module exposing inbound/outbound transfers API          | [link](https://github.com/chadury2021/Osmosis-Bifrost/blob/master/bridge-docs/bridge/README.md#inbound-transfers)   | :heavy_check_mark: |
| 3    | Employed x/tokenfactory functionality to perform transfers                   | [link](https://github.com/chadury2021/Osmosis-Bifrost/blob/master/bridge-docs/bridge/README.md#minting-and-burning) | :heavy_check_mark: |
| 4    | Implemented observer, allowing fast addition of new external chains          | [link](https://github.com/chadury2021/Osmosis-Bifrost/blob/master/bridge-docs/bridge/README.md#observer)            | :heavy_check_mark: |
| 5    | Wired the observer as a command-line query without modifying the Osmosis app | [link](https://github.com/osmosis-labs/osmosis/pull/7896)                                                           | :heavy_check_mark: |
| 6    | Performed a BTC-to-OSMO transfer using the Bitcoin testnet                   | [link](inbound-transfer.md)                                                                                         | :heavy_check_mark: |
| 7    | Featured easy enabling or disabling bridging for a selected asset            | [link](https://github.com/osmosis-labs/osmosis/blob/main/x/bridge/keeper/assets.go)                                 | :heavy_check_mark: |

## Wins over Thorchain

1. **Fungibility:** the bridged asset is fungible, which means it can be moved across chains, it does not live only as pool liquidity which is the Thorchain case.
2. Observer is off-chain and fully separated from the Osmosis app, so it could be modified without chain upgrades
3. Adding a new asset is a runtime process (not compile time), so it doesn't require either an upgrade or a validator
   restart
4. All transfers are simple and lightweight on-chain operations (mint/burn), so performance should be much better (
   benchmarks are to be done)


## Roadmap

| Item | Agenda | Date | Delivery |
|-|-|-|-|
| 1 | TSS signing: finally decide on a library and an approach to using it | 09-04-2024 |  |
| 2 | Finish OSMO-to-BTC transfers | 16-04-2024 |  |
| 3 | Security audit (scheduled) | 23-04-2024 |  |
| 4 | Refunds | 30-04-2024 |  |
| 5 | Validator fees | 07-05-2024 |  |
| 6 | Validator rotation | 14-05-2024 |  |
| 7 | Minimum transfer amount | 17-05-2024 |  |
| 8 | Implement quarantining | 24-05-2024 |  |
| 9 | Saving observer params into the dedicated config | 27-05-2024 |  |
