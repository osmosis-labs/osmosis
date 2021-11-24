# Auth

::: warning NOTE
Osmosis's Auth module inherits from Cosmos SDK's [`auth`](https://docs.cosmos.network/master/modules/auth/) module. This document is a stub, and covers mainly important Osmosis-specific notes about how it is used.
:::

Osmosis's Auth module extends the functionality from Cosmos SDK's `auth` module with a modified ante handler which applies the stability layer fee alongside the all basic transaction validity checks (signatures, nonces, auxiliary fields). In addition, a special vesting account type is defined, which handles the logic for token vesting from the Osmo presale.

## Fees

The Auth module reads the current effective `TaxRate` and `TaxCap` parameters from the [`Treasury`](./spec-treasury.md) module to enforce a stability layer fee.

### Gas Fee

As with any other transaction, [`MsgSend`](./spec-bank.md#msgsend) and [`MsgMultiSend`](./spec-bank.md#msgmultisend) pay a gas fee the size of which depends on validator's preferences (each validator sets his own min-gas-fees) and the complexity of the transaction. [Notes on gas and fees](/Reference/osmosisd/#fees) has a more detailed explanation of how gas is computed. Important detail to note here is that gas fees are specified by the sender when the transaction is outbound.

### Stability Fee

In addition to the gas fee, the ante handler charges a stability fee that is a percentage of the transaction's value only for the **Stable Coins** except **Osmo**. It reads the Tax Rate and Tax Cap parameters from the [`Treasury`](./spec-treasury.md) module to compute the amount of stability tax that needs to be charged.

The **Tax Rate** is a parameter agreed upon by the network that specifies the percentage of payment transactions that will be collected as Tax Proceeds in the block reward, which will be distributed among the validators. The distribution model is a bit complicated and explained in detail [here](../validator/faq.md#how-are-block-provisions-distributed). The taxes collected per transaction cannot exceed the specific **Tax Cap** defined for that transaction's denomination. Every epoch, the Tax Rate and Tax Caps are recalibrated automatically by the network; see [here](spec-treasury.md#monetary-policy-levers) for more details.

For an example `MsgSend` transaction of µSDR tokens,

```text
stability fee = min(1000 * tax_rate, tax_cap(usdr))
```

For a `MsgMultiSend` transaction, a stability fee is charged from every outbound transaction.

## Parameters

The subspace for the Auth module is `auth`.

```go
type Params struct {
	MaxMemoCharacters      uint64 `json:"max_memo_characters" yaml:"max_memo_characters"`
	TxSigLimit             uint64 `json:"tx_sig_limit" yaml:"tx_sig_limit"`
	TxSizeCostPerByte      uint64 `json:"tx_size_cost_per_byte" yaml:"tx_size_cost_per_byte"`
	SigVerifyCostED25519   uint64 `json:"sig_verify_cost_ed25519" yaml:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1 uint64 `json:"sig_verify_cost_secp256k1" yaml:"sig_verify_cost_secp256k1"`
}
```

### MaxMemoCharacters

Maximum permitted number of characters in the memo of a transaction.

- type: `uint64`
- default: `256`

### TxSigLimit

The maximum number of signers in a transaction. A single transaction can have multiple messages and multiple signers. The sig verification cost is much higher than other operations, so we limit this to 100.

- type: `uint64`
- default: `100`

### TxSizeCostPerByte

Used to compute gas consumption of the transaction, `TxSizeCostPerByte * txsize`.

- type: `uint64`
- default: `10`

### SigVerifyCostED25519

The gas cost for verifying ED25519 signatures.

- type: `uint64`
- default: `590`

### SigVerifyCostSecp256k1

The gas cost for verifying Secp256k1 signatures.

- type: `uint64`
- default: `1000`
