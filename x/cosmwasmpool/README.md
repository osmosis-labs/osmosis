# CosmWasm Pool

The CosmWasm Pool Module is an extension for the Osmosis pools, aiming to create a custom module that allows users to create and manage liquidity pools backed by CosmWasm smart contracts. The feature enables developers to build and deploy custom smart contracts that can be integrated with the rest of the pool types on the Osmosis chain.

The module is built on top of the CosmWasm smart contracting platform, which provides a secure and efficient way to develop and execute WebAssembly (Wasm) smart contracts on the Cosmos SDK.

Having pools in CosmWasm provides several benefits, one of which is avoiding the need for chain upgrades when introducing new functionalities or modifying existing ones related to liquidity pools. This advantage is particularly important in the context of speed of development and iteration.

An example of a CosmWasm pool type: https://github.com/osmosis-labs/transmuter



## Creating new CosmWasm Pool

To create new CosmWasm Pool, you need to create a new CosmWasm contract that implement [CosmWasm Pool interface](#cosmwasm-pool-interface) and store it on chain first. Then new pool can be created by sending `MsgCreateCosmWasmPool`.

`MsgCreateCosmWasmPool` contains `InstantiateMsg`, which is a message that will be passed to the CosmWasm contract when it is instantiated. The structure of the message is defined by the contract developer, and can contain any information that the contract needs to be instantiated. JSON format is used for `InstantiateMsg`.

```mermaid
sequenceDiagram
    participant Client
    participant x/cosmwasmpool
    participant x/wasm
    participant x/poolmanager

    Client ->> x/cosmwasmpool: MsgCreateCosmWasmPool {CodeId, InstantiateMsg, Sender}

    Note over x/wasm: Given there is a pool contract with CodeId


    x/cosmwasmpool ->> x/wasm: Call InstantiateContract(CodeId, InstantiateMsg)
    x/wasm -->> x/cosmwasmpool: ContractAddress
    x/poolmanager ->> x/cosmwasmpool: Call GetNextPoolId()
    x/cosmwasmpool ->> x/poolmanager: Call SetNextPoolId(poolId)

    Note over x/cosmwasmpool: Store CodeId, ContractAddress, and PoolId

    x/cosmwasmpool -->>  Client: MsgCreateCosmWasmPoolResponse {PoolId}
```




## CosmWasm Pool Interface

<!-- 

  
- Flow
  - Initialization
  - swapping: sudo
  - activation/deactivation: per code id
  - fee collection/distribution: TBD

- Contract State
  - this is might varies form contract to contract but we can define some common state

- Pool Contract Interface
  - serialization protocol
  - entrypoints
 -->