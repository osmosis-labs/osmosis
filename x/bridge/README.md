# Bridge

## Abstract

This README outlines an implementation of the bridging mechanism developed to enable cross-chain asset transfers from external chains to Osmosis and vice versa. This leverages the ideas rooted in the Thorchain bridging protocol to connect Osmosis with other blockchains, starting with the BTC network, to facilitate decentralized cross-chain asset transfers.

## Contents

1. **[Concepts](#Concepts)**
2. **[Design](#Design)**
3. **[Events](#Events)**
4. **[Keeper](#Keeper)**
5. **[Queries](#Queries)**
6. **[Messages](#Messages)**

## Concepts

### Super valset

The core concept of cross-chain transfers involves a selected group of validators, known as the **super valset**, which operates the **x/bridge observer**. This functionality enables the bridging of assets between external chains and Osmosis, and vice versa. Initially, the **super valset** is formed by Osmosis validators and is empty at genesis. However, its composition can be altered through governance processes.

The **super valset** is responsible for monitoring activities on external chains and managing all incoming transfers to Osmosis. Conversely, it also tracks events within Osmosis, identifies outbound transfers to external chains, and facilitates these transactions.

### Vault

To facilitate cross-chain transfers, a representative entity, referred to as the vault, must be present on the external chain. Essentially, the vault is an address on the external chain that is persistently being observed by the **super valset** members. If the client wants to execute a cross-chain transfer from the external chain to Osmosis, they must send their assets to the **vault**, specifying the desired Osmosis recipient. After that, the **super valset** will observe this transfer and start handling it.

## Inbound transfers

1. The client sends the transfer in the external chain to the dedicated vault
2. The client should specify the desired recipient address in Osmosis (depending on the chain, e.g., this can be done through the memo)
3. Validators from the **super valset** are running **x/bridge observer** observing the vault
4. As soon as they see a new incoming transfer, they handle it and convert it to the **MsgInboundTransfer** message
5. Validators send this message to the chain proposer
6. Note that this is done by all validators in the **super valset**, so the number of messages in the block would be equal to the cardinality of the **super valset**
7. After the block is formed, validators (all Osmosis validators) start processing it
8. Validators process transfers one by one, accumulating all voters (i.e., senders) in the dedicated list
9. Once the number of voters is greater than or equal to the number of votes needed for the transfer (module param), the transfer is considered to be finalized

![inbound_transfers](images/inbound_transfer.png)

### Observer

The entity created to observe external chains and 

## Concepts

![image.png](images/mint_burn.png)

## Design

## Events

## Keeper

## Queries

## Messages