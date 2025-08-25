## Affiliate Swap (CosmWasm on Osmosis)

This contract routes swaps via Osmosis poolmanager and splits the output by an affiliate fee in basis points. The affiliate portion is sent to a configured Osmosis address and the remainder to the swap caller.

### Instantiate

Fields:
- `owner`: admin address
- `affiliate_addr`: osmosis address receiving fees
- `affiliate_bps`: fee in basis points (0-10000)

### Execute

- `SwapWithFee { input_coin, output_denom, min_output_amount, route }`
  - Sends `input_coin` to contract, performs swap along `route`, enforces `min_output_amount`, then splits output `(affiliate_bps/10000)` to affiliate and the rest back to caller.
- `UpdateAffiliate { affiliate_addr, affiliate_bps }` (owner only)
- `TransferOwnership { new_owner }` (owner only)

### Query

- `Config {}` -> owner, affiliate addr, affiliate bps

### Build and Test

From repository root:

```bash
cargo test -p affiliate-swap
```

Optimize wasm:

```bash
cargo run --bin build-schema -p affiliate-swap
make -C cosmwasm/contracts/affiliate-swap optimize
```

### Deployment (Osmosis testnet)

Use `osmosisd tx wasm store` to upload the optimized wasm, then instantiate with desired params.

### Notes

- Uses stargate message `MsgSwapExactAmountIn`. Contract must hold the input funds in `info.funds` and forwards the output via bank sends after swap success.
