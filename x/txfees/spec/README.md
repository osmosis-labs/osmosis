# Txfees

The txfees modules allows nodes to easily support many tokens for usage as txfees, while letting node operators only specify their tx fee parameters for a single "base" asset.
This is done by having this module maintain an allow-list of token denoms which can be used as tx fees, each with some associated metadata.
Then this metadata is used in tandem with a "Spot Price Calculator" provided to the module, to convert the provided tx fees into their equivalent value in the base denomination.
Currently the only supported metadata & spot price calculator is using a GAMM pool ID & the GAMM keeper.

## State Changes

* Adds a whitelist of tokens that can be used as fees on the chain.
  * Any token not on this list cannot be provided as a tx fee.
* Adds a new SDK message for creating governance proposals for adding new TxFee denoms.

## Local Mempool Filters Added

* If you specify a min-tx-fee in the $BASEDENOM then
  * Your node will allow any tx w/ tx fee in the whitelist of fees, and a sufficient osmo-equivalent price to enter your mempool
  * The osmo-equivalent price for determining sufficiency is rechecked after every block. (During the mempools RecheckTx)
    * TODO: further consider if we want to take this tradeoff. Allows someone who manipulates price for one block to flush txs using that asset as fee from most of the networks' mempools.
    * The simple alternative is only check fee equivalency at a txs entry into the mempool, which allows someone to manipulate price down to have many txs enter the chain at low cost.
    * Another alternative is to use TWAP instead of Spot Price once it is available on-chain
    * The former concern isn't very worrisome as long as some nodes have 0 min tx fees.
* A separate min-gas-fee can be set on every node for arbitrage txs. Methods of detecting an arb tx atm
  * does start token of a swap = final token of swap (definitionally correct)
  * does it have multiple swap messages, with different tx ins. If so, we assume its an arb.
    * This has false positives, but is intended to avoid the obvious solution of splitting an arb into multiple messages.
  * We record all denoms seen across all swaps, and see if any duplicates. (TODO)
  * Contains both JoinPool and ExitPool messages in one tx.
    * Has some false positives.
  * These false positives seem like they primarily will get hit during batching of many distinct operations, not really in one atomic action.
* A max wanted gas per any tx can be set to filter out attack txes.
* If tx wanted gas > than predefined threshold of 1M, then separate 'min-gas-price-for-high-gas-tx' option used to calculate min gas price.

## Queries

base-denom    
Query the base fee denom

denom-pool-id 
Query the pool id associated with a specific whitelisted fee token

fee-tokens    
Query the list of non-basedenom fee tokens and their associated pool ids

## Future directions

* Want to add in a system to add in general "tx fee credits" for different on-chain usages
  * e.g. making 0 fee txs under certain usecases
* If other chains would like to use this, we should brainstorm mechanisms for extending the metadata proto fields
