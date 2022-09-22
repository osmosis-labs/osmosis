# Streamswap

## Abstract

Streamswap is a new way and innovative way of selling token sale.
The mechanism allows anyone to create a new Sale event and sell any
amount of tokens in a more democratic way than the traditional solutions.

## Context

Since the first ICO boom, token sale mechanism was one of the driving
force for web3 onboarding.
Promise of a cheap tokens which can quickly accrue value is very attractive
for casual any sort of investors. Easy way of fundrising (the funding team)
opened doors for cohorts of new teams to focus on building on web3.

Traditional mechanisms of token sale included:

- Automated ICO, where team decides about the issuance price and the sale
  happens through a swap controlled by a smart contract.
- Regulated, centralized ICO - token sale controlled by a dedicated company,
  which will preform all operations using centralized services meeting
  regulatory requirements (KYC...). Example: Coinlist sales.
- Balancer style ICO: a novel solution to utilize Dutch Auction mechanism to
  find a fair strike price.

The first two mechanisms are not well suited for early stage startups, where
the token sale price is usually defined by a founding team and can't be
impacted by the ecosystem wisdom. False marketing actions are usually setup
to support their initial price.

The latter mechanism is not democratic - big entities can control the
price movements or place big orders leaving smaller investors with nothing.

## Design

### Sale Creation

[Sale](https://github.com/osmosis-labs/osmosis/blob/main/proto/osmosis/streamswap/v1/state.proto#L11) object represent a particular token sale event and describes the main
required parameters guarding the sale process:

- `treasury`: address where the sale earnings will go. When the sale is over,
  anyone can trigger a [`MsgFinalizeSale`](https://github.com/osmosis-labs/osmosis/blob/main/proto/osmosis/streamswap/v1/tx.proto#L42)
  to clean up the sale state and move the earning to the treasury.
- `id`: unique identifier of the sale.
- `token_out`: denom to sale (distributed to the investors).
  Also known as a base currency.
- `token_in`: payment denom - used to buy `token_out`.
  Also known as quote currency.
- `token_out_supply`: total initial supply of `token_in` to sale.
- `start_time`: Unix timestamp when the sale starts.
- `end_time`: Unix timestamp when the sale ends.
- `name`: Name of the sale.
- `url`: an external resource describing a sale. Can be IPFS link or a
  commonwealth post.

The `Sale` object contains also other internal attributes to describe the current
status of the sale.

Anyone can create a `Sale` by sending [`MsgCreateSale`](https://github.com/osmosis-labs/osmosis/blob/robert%2Fstreamswap-spec/proto/osmosis/streamswap/v1/tx.proto#L21) transaction.
When doing so, `token_out` amount of `Sale.token_out` tokens will be debited from
his account and escrowed in the module account to distribute to Sale investors.
Moreover the creator will be charged `sale_creation_fee` (module param) and the
fee will be transferred to `sale_creation_fee_recipient` (module param) account.
This fee is not recoverable.

See other [module parameters](https://github.com/osmosis-labs/osmosis/main/proto/osmosis/streamswap/v1/params.proto) which control the sale creation.

### Investing and distribution mechanism

Anyone can join a sale by sending a [MsgSubscribe](https://github.com/osmosis-labs/osmosis/blob/main/proto/osmosis/streamswap/v1/tx.proto#L13) transaction.
When doing so, the transaction author has to specify the `amount` he wants to spend in the sale.
That `amount` will be credited from tx author and pledged to the sale.

`MsgSubscribe` can be submitted at any time after the sale is created and before it's end time.

From that moment, the investor will join the **token sale distribution stream**:

- The distribution happens discretely in rounds. Each round is 1 second long.
  We define: `total_rounds := (sale.end_time - sale.start_time) / round_duration`.
- During each round we stream `round_supply := sale.token_out_supply / total_rounds`.
- During each round each investor receives `round_supply * current_investor_pledge / total_remaining_sale_pledge` share of the `token_out`.

At any time an investor can increase his participation for the sale by sending again `MsgSubscribe`
(his pledge will increase accordingly) or cancel it by sending
[`MsgWithdraw`](https://github.com/osmosis-labs/osmosis/blob/main/proto/osmosis/streamswap/v1/tx.proto#32).
When cancelling, the module will send back unspent pledged tokens to the investor
and keep the purchased tokens until the sale end_time.

### Withdrawing purchased tokens

When participating in a sale, investors receives a stream of sale tokens.
These tokens are locked until sale end to avoid second market creating during
the sale. Once sale is finished (block time is after `sale.end_time`), every
investor can send [`MsgExitSale`](https://github.com/osmosis-labs/osmosis/blob/main/proto/osmosis/streamswap/v1/tx.proto#L37)
to close his position and withdraw purchased tokens to his account.

### Withdrawing sale proceedings

To withdraw earned token to the `sale.treasury` account anyone can send a
transaction with [`MsgFinalizeSale`](https://github.com/osmosis-labs/osmosis/blob/main/proto/osmosis/streamswap/v1/tx.proto#L42) after the `sale.end_time`.
This transaction will send `sale.income` tokens from the module escrow account
to the `sale.treasury` and set `sale.finalized = true`.

### Events

The module uses typed events. Please see [`event.proto`](https://github.com/osmosis-labs/osmosis/blob/main/proto/osmosis/streamswap/v1/event.proto)
to inspect list events.

## Consequences

- The new sale mechanism provides a truly democratic way for token distribution and sale.
- It can be easily integrated with AMM pools: proceedings from the sale can
  automatically be pledged to AMM.
- After the sale ends, there are few algorithms to provide token indicative market price:
  - average price: `sale.income / sale.token_out_supply`
  - last streamed price: `round_supply(last_round) / (sale.token_out_supply / total_rounds)`

## Future directions

- providing incentive for sales with `OSMO` or `ATOM` used as a base currency.
