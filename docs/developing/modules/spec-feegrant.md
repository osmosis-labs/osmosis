# Fee grant

::: warning Note
Osmosis's fee grant module inherits from the Cosmos SDK's [`feegrant`](https://docs.cosmos.network/master/modules/feegrant/) module. This document is a stub and explains mainly important Osmosis-specific notes about how it is used.
:::

This module allows an account, the granter, to permit another account, the grantee, to pay for fees from the granter's account balance. Grantees will not need to maintain their own balance for paying fees.

## Concepts

### Grant

`Grant` is stored in the KVStore to record a grant with full context.

Every `grant` contains the following information:

- `granter`: The account address that gives permission to the grantee.
- `grantee`: The beneficiary account address.
- `allowance`: The [type of fee allowance](#fee-allowance-types) given to the grantee. `Allowance` accepts an interface that implements `FeeAllowanceI` encoded as `Any` type, as shown in the following example:

  ```
    // allowance can be any of basic and filtered fee allowance.
    google.protobuf.Any allowance = 3 [(cosmos_proto.accepts_interface) = "FeeAllowanceI"];
  }
  ```

  The following example shows `FeeAllowanceI`:

  ```
  type FeeAllowanceI interface {
          // Accept can use fee payment requested as well as timestamp of the current block
          // to determine whether or not to process this. This is checked in
          // Keeper.UseGrantedFees and the return values should match how it is handled there.
          //
          // If it returns an error, the fee payment is rejected, otherwise it is accepted.
          // The FeeAllowance implementation is expected to update it's internal state
          // and will be saved again after an acceptance.
          //
          // If remove is true (regardless of the error), the FeeAllowance will be deleted from storage
          // (eg. when it is used up). (See call to RevokeFeeAllowance in Keeper.UseGrantedFees)
          Accept(ctx sdk.Context, fee sdk.Coins, msgs []sdk.Msg) (remove bool, err error)

          // ValidateBasic should evaluate this FeeAllowance for internal consistency.
          // Don't allow negative amounts, or negative periods for example.
          ValidateBasic() error
  }
  ```

Only one fee grant is allowed between a granter and a grantee. Self-grants are prohibited.

### Fee allowance types

- [`BasicAllowance`](#basicallowance)
- [`PeriodicAllowance`](#periodicallowance)

`BasicAllowance` permits the grantee to pay fees by using funds from the  granter's account. If the threshold for either `spend_limit` or `expiration` is met, the grant is removed from the state.

```
// BasicAllowance implements Allowance with a one-time grant of tokens
// that optionally expires. The grantee can use up to SpendLimit to cover fees.
message BasicAllowance {
  option (cosmos_proto.implements_interface) = "FeeAllowanceI";

  // spend_limit specifies the maximum amount of tokens that can be spent
  // by this allowance and will be updated as tokens are spent. If it is
  // empty, there is no spend limit and any amount of coins can be spent.
  repeated cosmos.base.v1beta1.Coin spend_limit = 1
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];

  // expiration specifies an optional time when this allowance expires
  google.protobuf.Timestamp expiration = 2 [(gogoproto.stdtime) = true];
}
```

- `spend_limit`: The amount of tokens from the granter's account that the grantee can spend. This value is optional. If it is blank, no spend limit is assigned, and the grantee can spend any amount of tokens from the granter's account before the expiration is met.

- `expiration`: The date and time when the grant expires. This value is optional. If it is blank, the grant does not expire.

To restrict the grantee when values for `spend_limit` and `expiration` are blank, revoke the grant.

`PeriodicAllowance` is a repeating fee allowance for a specified period and for a specified maximum number of tokens that can spent within that period.

::: details `PeriodicAllowance` code

```
// PeriodicAllowance extends Allowance to allow for both a maximum cap,
// as well as a limit per time period.
message PeriodicAllowance {
  option (cosmos_proto.implements_interface) = "FeeAllowanceI";

  // basic specifies a struct of `BasicAllowance`
  BasicAllowance basic = 1 [(gogoproto.nullable) = false];

  // period specifies the time duration in which period_spend_limit coins can
  // be spent before that allowance is reset
  google.protobuf.Duration period = 2 [(gogoproto.stdduration) = true, (gogoproto.nullable) = false];

  // period_spend_limit specifies the maximum number of coins that can be spent
  // in the period
  repeated cosmos.base.v1beta1.Coin period_spend_limit = 3
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];

  // period_can_spend is the number of coins left to be spent before the period_reset time
  repeated cosmos.base.v1beta1.Coin period_can_spend = 4
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];

  // period_reset is the time at which this period resets and a new one begins,
  // it is calculated from the start time of the first transaction after the
  // last period ended
  google.protobuf.Timestamp period_reset = 5 [(gogoproto.stdtime) = true, (gogoproto.nullable) = false];
}

// AllowedMsgAllowance creates allowance only for specified message types.
message AllowedMsgAllowance {
  option (gogoproto.goproto_getters)         = false;
  option (cosmos_proto.implements_interface) = "FeeAllowanceI";

  // allowance can be any of basic and filtered fee allowance.
  google.protobuf.Any allowance = 1 [(cosmos_proto.accepts_interface) = "FeeAllowanceI"];

  // allowed_messages are the messages for which the grantee has the access.
  repeated string allowed_messages = 2;
}

// Grant is stored in the KVStore to record a grant with full context
message Grant {
  // granter is the address of the user granting an allowance of their funds.
  string              granter   = 1;

  // grantee is the address of the user being granted an allowance of another user's funds.
  string              grantee   = 2;
```
:::

- `basic`: The instance of `BasicAllowance`.  It is optional. If empty, the grant will have not have a `spend_limit` or `expiration`.

- `period`: The duration that `PeriodicAllowance` is granted. After each period expires, `period_spend_limit` is reset.

- `period_spend_limit`: The maximum number of tokens that the grantee is allowed to spend during the period.

- `period_can_spend`: The number of tokens remaining to be spent before the period_reset time.

- `period_reset`: The time when the period ends and a new period begins.

### Fee account flag

To run transactions that use fee grant from the CLI, specify the `FeeAccount` flag followed by the granter's account address. When this flag is set, `clientCtx` appends the granter's account address.

::: details `FeeAccount` code

```
if clientCtx.FeeGranter == nil || flagSet.Changed(flags.FlagFeeAccount) {
  granter, _ := flagSet.GetString(flags.FlagFeeAccount)

  if granter != "" {
    granterAcc, err := sdk.AccAddressFromBech32(granter)
    if err != nil {
      return clientCtx, err
    }

    clientCtx = clientCtx.WithFeeGranterAddress(granterAcc)
  }
}
```

```
package tx

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// GenerateOrBroadcastTxCLI will either generate and print and unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxCLI(clientCtx client.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg) error {
	txf := NewFactoryCLI(clientCtx, flagSet)
	return GenerateOrBroadcastTxWithFactory(clientCtx, txf, msgs...)
}

// GenerateOrBroadcastTxWithFactory will either generate and print and unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxWithFactory(clientCtx client.Context, txf Factory, msgs ...sdk.Msg) error {
	if clientCtx.GenerateOnly {
		return GenerateTx(clientCtx, txf, msgs...)
	}

	return BroadcastTx(clientCtx, txf, msgs...)
}

// GenerateTx will generate an unsigned transaction and print it to the writer
// specified by ctx.Output. If simulation was requested, the gas will be
// simulated and also printed to the same writer before the transaction is
// printed.
func GenerateTx(clientCtx client.Context, txf Factory, msgs ...sdk.Msg) error {
	if txf.SimulateAndExecute() {
		if clientCtx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		_, adjusted, err := CalculateGas(clientCtx.QueryWithData, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return err
	}

	json, err := clientCtx.TxConfig.TxJSONEncoder()(tx.GetTx())
	if err != nil {
		return err
	}

	return clientCtx.PrintString(fmt.Sprintf("%s\n", json))
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(clientCtx client.Context, txf Factory, msgs ...sdk.Msg) error {
	txf, err := PrepareFactory(clientCtx, txf)
	if err != nil {
		return err
	}

	if txf.SimulateAndExecute() || clientCtx.Simulate {
		_, adjusted, err := CalculateGas(clientCtx.QueryWithData, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	if clientCtx.Simulate {
		return nil
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		out, err := clientCtx.TxConfig.TxJSONEncoder()(tx.GetTx())
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", out)

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)

		if err != nil || !ok {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", "cancelled transaction")
			return err
		}
	}

	tx.SetFeeGranter(clientCtx.GetFeeGranterAddress())
	err = Sign(txf, clientCtx.GetFromName(), tx, true)
	if err != nil {
		return err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx.GetTx())
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

// WriteGeneratedTxResponse writes a generated unsigned transaction to the
// provided http.ResponseWriter. It will simulate gas costs if requested by the
// BaseReq. Upon any error, the error will be written to the http.ResponseWriter.
// Note that this function returns the legacy StdTx Amino JSON format for compatibility
// with legacy clients.
func WriteGeneratedTxResponse(
	ctx client.Context, w http.ResponseWriter, br rest.BaseReq, msgs ...sdk.Msg,
) {
	gasAdj, ok := rest.ParseFloat64OrReturnBadRequest(w, br.GasAdjustment, flags.DefaultGasAdjustment)
	if !ok {
		return
	}

	gasSetting, err := flags.ParseGasSetting(br.Gas)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	txf := Factory{fees: br.Fees, gasPrices: br.GasPrices}.
		WithAccountNumber(br.AccountNumber).
		WithSequence(br.Sequence).
		WithGas(gasSetting.Gas).
		WithGasAdjustment(gasAdj).
		WithMemo(br.Memo).
		WithChainID(br.ChainID).
		WithSimulateAndExecute(br.Simulate).
		WithTxConfig(ctx.TxConfig).
		WithTimeoutHeight(br.TimeoutHeight)

	if br.Simulate || gasSetting.Simulate {
		if gasAdj < 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.ErrorInvalidGasAdjustment.Error())
			return
		}

		_, adjusted, err := CalculateGas(ctx.QueryWithData, txf, msgs...)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		txf = txf.WithGas(adjusted)

		if br.Simulate {
			rest.WriteSimulationResponse(w, ctx.LegacyAmino, txf.Gas())
			return
		}
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	stdTx, err := ConvertTxToStdTx(ctx.LegacyAmino, tx.GetTx())
	if rest.CheckInternalServerError(w, err) {
		return
	}

	output, err := ctx.LegacyAmino.MarshalJSON(stdTx)
	if rest.CheckInternalServerError(w, err) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

// BuildUnsignedTx builds a transaction to be signed given a set of messages. The
// transaction is initially created via the provided factory's generator. Once
// created, the fee, memo, and messages are set.
func BuildUnsignedTx(txf Factory, msgs ...sdk.Msg) (client.TxBuilder, error) {
	if txf.chainID == "" {
		return nil, fmt.Errorf("chain ID required but not specified")
	}

	fees := txf.fees

	if !txf.gasPrices.IsZero() {
		if !fees.IsZero() {
			return nil, errors.New("cannot provide both fees and gas prices")
		}

		glDec := sdk.NewDec(int64(txf.gas))

		// Derive the fees based on the provided gas prices, where
		// fee = ceil(gasPrice * gasLimit).
		fees = make(sdk.Coins, len(txf.gasPrices))

		for i, gp := range txf.gasPrices {
			fee := gp.Amount.Mul(glDec)
			fees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	tx := txf.txConfig.NewTxBuilder()

	if err := tx.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	tx.SetMemo(txf.memo)
	tx.SetFeeAmount(fees)
	tx.SetGasLimit(txf.gas)
	tx.SetTimeoutHeight(txf.TimeoutHeight())

	return tx, nil
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func BuildSimTx(txf Factory, msgs ...sdk.Msg) ([]byte, error) {
	txb, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: &secp256k1.PubKey{},
		Data: &signing.SingleSignatureData{
			SignMode: txf.signMode,
		},
		Sequence: txf.Sequence(),
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, err
	}

	protoProvider, ok := txb.(authtx.ProtoTxProvider)
	if !ok {
		return nil, fmt.Errorf("cannot simulate amino tx")
	}
	simReq := tx.SimulateRequest{Tx: protoProvider.GetProtoTx()}

	return simReq.Marshal()
}

// CalculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func CalculateGas(
	queryFunc func(string, []byte) ([]byte, int64, error), txf Factory, msgs ...sdk.Msg,
) (tx.SimulateResponse, uint64, error) {
	txBytes, err := BuildSimTx(txf, msgs...)
	if err != nil {
		return tx.SimulateResponse{}, 0, err
	}

	// TODO This should use the generated tx service Client.
	// https://github.com/cosmos/cosmos-sdk/issues/7726
	bz, _, err := queryFunc("/cosmos.tx.v1beta1.Service/Simulate", txBytes)
	if err != nil {
		return tx.SimulateResponse{}, 0, err
	}

	var simRes tx.SimulateResponse

	if err := simRes.Unmarshal(bz); err != nil {
		return tx.SimulateResponse{}, 0, err
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

// PrepareFactory ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory. A new Factory with
// the updated fields will be returned.
func PrepareFactory(clientCtx client.Context, txf Factory) (Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.accountRetriever.EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.accountNumber, txf.sequence
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.accountRetriever.GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
}

// SignWithPrivKey signs a given tx with the given private key, and returns the
// corresponding SignatureV2 if the signing is successful.
func SignWithPrivKey(
	signMode signing.SignMode, signerData authsigning.SignerData,
	txBuilder client.TxBuilder, priv cryptotypes.PrivKey, txConfig client.TxConfig,
	accSeq uint64,
) (signing.SignatureV2, error) {
	var sigV2 signing.SignatureV2

	// Generate the bytes to be signed.
	signBytes, err := txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return sigV2, err
	}

	// Sign those bytes
	signature, err := priv.Sign(signBytes)
	if err != nil {
		return sigV2, err
	}

	// Construct the SignatureV2 struct
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: signature,
	}

	sigV2 = signing.SignatureV2{
		PubKey:   priv.PubKey(),
		Data:     &sigData,
		Sequence: accSeq,
	}

	return sigV2, nil
}

func checkMultipleSigners(mode signing.SignMode, tx authsigning.Tx) error {
	if mode == signing.SignMode_SIGN_MODE_DIRECT &&
		len(tx.GetSigners()) > 1 {
		return sdkerrors.Wrap(sdkerrors.ErrNotSupported, "Signing in DIRECT mode is only supported for transactions with one signer only")
	}
	return nil
}

// Sign signs a given tx with a named key. The bytes signed over are canconical.
// The resulting signature will be added to the transaction builder overwriting the previous
// ones if overwrite=true (otherwise, the signature will be appended).
// Signing a transaction with mutltiple signers in the DIRECT mode is not supprted and will
// return an error.
// An error is returned upon failure.
func Sign(txf Factory, name string, txBuilder client.TxBuilder, overwriteSig bool) error {
	if txf.keybase == nil {
		return errors.New("keybase must be set prior to signing a transaction")
	}

	signMode := txf.signMode
	if signMode == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		// use the SignModeHandler's default mode if unspecified
		signMode = txf.txConfig.SignModeHandler().DefaultMode()
	}
	if err := checkMultipleSigners(signMode, txBuilder.GetTx()); err != nil {
		return err
	}

	key, err := txf.keybase.Key(name)
	if err != nil {
		return err
	}
	pubKey := key.GetPubKey()
	signerData := authsigning.SignerData{
		ChainID:       txf.chainID,
		AccountNumber: txf.accountNumber,
		Sequence:      txf.sequence,
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generated the
	// sign bytes. This is the reason for setting SetSignatures here, with a
	// nil signature.
	//
	// Note: this line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}
	var prevSignatures []signing.SignatureV2
	if !overwriteSig {
		prevSignatures, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			return err
		}
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return err
	}

	// Generate the bytes to be signed.
	bytesToSign, err := txf.txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return err
	}

	// Sign those bytes
	sigBytes, _, err := txf.keybase.Sign(name, bytesToSign)
	if err != nil {
		return err
	}

	// Construct the SignatureV2 struct
	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}

	if overwriteSig {
		return txBuilder.SetSignatures(sig)
	}
	prevSignatures = append(prevSignatures, sig)
	return txBuilder.SetSignatures(prevSignatures...)
}

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}
```

```
func (w *wrapper) SetFeeGranter(feeGranter sdk.AccAddress) {
	if w.tx.AuthInfo.Fee == nil {
		w.tx.AuthInfo.Fee = &tx.Fee{}
	}

	w.tx.AuthInfo.Fee.Granter = feeGranter.String()

	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	w.authInfoBz = nil
}
```

```
// Fee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
message Fee {
  // amount is the amount of coins to be paid as a fee
  repeated cosmos.base.v1beta1.Coin amount = 1
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];

  // gas_limit is the maximum gas that can be used in transaction processing
  // before an out of gas error occurs
  uint64 gas_limit = 2;

  // if unset, the first signer is responsible for paying the fees. If set, the specified account must pay the fees.
  // the payer must be a tx signer (and thus have signed this field in AuthInfo).
  // setting this field does *not* change the ordering of required signers for the transaction.
  string payer = 3;

  // if set, the fee payer (either the first signer or the value of the payer field) requests that a fee grant be used
  // to pay fees instead of the fee payer's own balance. If an appropriate fee grant does not exist or the chain does
  // not support fee grants, this will fail
  string granter = 4;
}
```
:::

The following example shows a CLI command with the `--fee-account` flag:

```
./osmosisd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --from validator-key --fee-account=osmo1fmcjjt6yc9wqup2r06urnrd928jhrde6gcld6n --chain-id=testnet --fees="10uOsmo"
```

### Granted fee deductions

Fees are deducted from grants in the `auth` ante handler.

### Gas

To prevent DoS attacks, using a filtered `feegrant` incurs gas. To ensure that all the grantee's transactions conform to the filter set by the granter, the SDK iterates over the allowed messages in the filter and charges 10 gas per filtered message. Then, the SDK iterates over the messages sent by the grantee to ensure the messages adhere to the filter, which also charges 10 gas per message. If the SDK finds a message that does not conform to the filter, the SDK stops iterating, and the transaction fails.

::: warning Warning
The gas is charged against the granted allowance. Ensure all your existing messages conform to the filter before you send transactions using your allowance.
:::

## State

### FeeAllowance

Fee allowances are identified by combining `Granter` (the account address that grants permission to another account to spend its available tokens on fees) with `Grantee` (the account address that receives permission to spend the granter's tokens on fees).

The following example shows how a fee allowance is stored in the state:

Grant: `0x00 | grantee_addr_len (1 byte) | grantee_addr_bytes | granter_addr_len (1 byte) | granter_addr_bytes -> ProtocolBuffer(Grant)`

```
// Grant is stored in the KVStore to record a grant with full context
type Grant struct {
	// granter is the address of the user granting an allowance of their funds.
	Granter string `protobuf:"bytes,1,opt,name=granter,proto3" json:"granter,omitempty"`
	// grantee is the address of the user being granted an allowance of another user's funds.
	Grantee string `protobuf:"bytes,2,opt,name=grantee,proto3" json:"grantee,omitempty"`
	// allowance can be any of basic and filtered fee allowance.
	Allowance *types1.Any `protobuf:"bytes,3,opt,name=allowance,proto3" json:"allowance,omitempty"`
}
```

## Message Types

### MsgGrantAllowance

A fee allowance grant will be created with the MsgGrantAllowance message.

```proto
// MsgGrantAllowance adds permission for Grantee to spend up to Allowance
// of fees from the account of Granter.
message MsgGrantAllowance {
  // granter is the address of the user granting an allowance of their funds.
  string              granter   = 1;

  // grantee is the address of the user being granted an allowance of another user's funds.
  string              grantee   = 2;

  // allowance can be any of basic and filtered fee allowance.
  google.protobuf.Any allowance = 3 [(cosmos_proto.accepts_interface) = "FeeAllowanceI"];
}
```

### MsgRevokeAllowance

A fee allowance grant will be revokeed with the MsgRevokeAllowance message.

```proto
// MsgRevokeAllowance removes any existing Allowance from Granter to Grantee.
message MsgRevokeAllowance {
  // granter is the address of the user granting an allowance of their funds.
  string granter = 1;

  // grantee is the address of the user being granted an allowance of another user's funds.
  string grantee = 2;
}
```
