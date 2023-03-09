# Swap Router Registry Contract

This is a CosmWasm smart contract that allows the creation and maintenance of cross-chain connection channels between IBC-enabled blockchains. This contract acts as a central registry where various blockchains can create, update, and delete IBC channels in a coordinated way, without having to deal with the complexity of handling low-level details when creating cross-chain swap messages.

The registry contains the following mappings:

- contract alias to contract address
  - maps a human-readable name to the address of a target contract
- chain channel to source chain/destination chain
  - maps a channel number to the source and destination chains it connects to
- chain name to Bech32 prefix
  - maps a chain name to its corresponding Bech32 prefix, which is used for address encoding

It also exposes a query entry point to retrieve the address from the alias, the destination chain from the source chain via the channel, the channel from the chain pair, the bech32 prefix from the chain name, and the native denom on the source chain from the IBC denom trace.

## Operations

### ModifyContractAlias

The `ModifyContractAlias` operation allows the owner to create, update, or delete aliases that can be used to identify contracts on other blockchains. The operation expects a vector of ContractAliasOperation, where each operation is either a CreateAlias, UpdateAlias, or DeleteAlias operation.

### ModifyChainChannelLinks

The `ModifyChainChannelLinks` operation allows the owner to create, update, or delete IBC channel links between blockchains. The operation expects a vector of ConnectionOperation, where each operation is either a CreateConnection, UpdateConnection, or DeleteConnection operation.

### ModifyBech32Prefixes

The `ModifyBech32Prefixes` operation allows the owner to create, update, or delete Bech32 prefixes for each blockchain. The operation expects a vector of ChainToPrefixOperation, where each operation is either a CreatePrefix, UpdatePrefix, or DeletePrefix operation.

### UnwrapCoin

The `UnwrapCoin` operation allows the contract to receive funds, which are then transferred to a destination address. This operation expects a receiver address where the funds will be transferred.

## Queries

### GetAddressFromAlias

The `GetAddressFromAlias` query allows a caller to retrieve the address of a contract on another blockchain, given the alias of that contract.

### GetDestinationChainFromSourceChainViaChannel

The `GetDestinationChainFromSourceChainViaChannel` query allows a caller to retrieve the destination chain for an IBC channel given the source chain and the channel id.

### GetChannelFromChainPair

The `GetChannelFromChainPair` query allows a caller to retrieve the channel id for an IBC channel given the source and destination chain.

### GetBech32PrefixFromChainName

The `GetBech32PrefixFromChainName` query allows a caller to retrieve the Bech32 prefix for a given blockchain.

### GetDenomTrace

The `GetDenomTrace` query allows a caller to retrieve the denom trace for a given IBC denom.

### UnwrapDenom

The `UnwrapDenom` query allows a caller to retrieve the denom trace for a given IBC denom and then unwrap it into its constituent coins on the originating blockchain.
