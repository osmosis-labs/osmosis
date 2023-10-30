package authenticator

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/v20/x/authenticator/iface"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CosmwasmAuthenticator struct {
	contractKeeper *keeper.PermissionedKeeper
	ak             *authkeeper.AccountKeeper
	cdc            codectypes.AnyUnpacker
	sigModeHandler authsigning.SignModeHandler

	contractAddr sdk.AccAddress
}

var (
	_ iface.Authenticator = &CosmwasmAuthenticator{}
)

func NewCosmwasmAuthenticator(contractKeeper *keeper.PermissionedKeeper, accountKeeper *authkeeper.AccountKeeper, sigModeHandler authsigning.SignModeHandler, cdc codectypes.AnyUnpacker) CosmwasmAuthenticator {
	return CosmwasmAuthenticator{
		contractKeeper: contractKeeper,
		sigModeHandler: sigModeHandler,
		ak:             accountKeeper,
		cdc:            cdc,
	}
}

func (cwa CosmwasmAuthenticator) Type() string {
	return "CosmwasmAuthenticator"
}

func (cwa CosmwasmAuthenticator) StaticGas() uint64 {
	return 0
}

type CosmwasmAuthenticatorInitData struct {
	Contract string `json:"contract"`
}

func (cwa CosmwasmAuthenticator) Initialize(data []byte) (iface.Authenticator, error) {
	var initData CosmwasmAuthenticatorInitData
	err := json.Unmarshal(data, &initData)
	if err != nil {
		return nil, err
	}
	if len(initData.Contract) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing contract address")
	}
	contractAddr, err := sdk.AccAddressFromBech32(initData.Contract)
	if err != nil {
		return nil, err
	}
	cwa.contractAddr = contractAddr
	return cwa, nil
}

func (cwa CosmwasmAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int,
	simulate bool,
) (iface.AuthenticatorData, error) {
	signers, signatures, signingTx, err := GetCommonAuthenticationData(ctx, tx, messageIndex, simulate)
	if err != nil {
		return SignatureData{}, err
	}

	// Get the signature for the message at msgIndex
	return SignatureData{
		Signers:    signers,
		Signatures: signatures,
		Tx:         signingTx,
		Simulate:   simulate,
	}, nil
}

type AuthenticateMsg struct {
	TxData        []byte         `json:"tx_data"`
	Account       sdk.AccAddress `json:"account"`
	Msg           codectypes.Any `json:"msg"`
	SignatureData SignatureData  `json:"signature_data"`
}

func (cwa CosmwasmAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.AuthenticationResult {
	signatureData, ok := authenticationData.(SignatureData)
	if !ok {
		return iface.Rejected("invalid signature verification data", sdkerrors.ErrInvalidType)
	}

	encodedMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil
	}

	// Retrieve and build the signer data struct
	genesis := ctx.IsGenesis() || ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	baseAccount := cwa.ak.GetAccount(ctx, account)
	if !genesis {
		accNum = baseAccount.GetAccountNumber()
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      baseAccount.GetSequence(),
	}

	// TODO: We probably want to pass a form of this to the contract that is easily parsable (or provide a helper function to unmarshall it)
	signBytes, err := cwa.sigModeHandler.GetSignBytes(txsigning.SignMode_SIGN_MODE_DIRECT, signerData, signatureData.Tx)
	if err != nil {
		return iface.Rejected("failed to get signBytes", err)
	}

	// TODO: We need to improve the api here

	// Should we use authtypes.StdSignDoc here instead?
	authMsg := AuthenticateMsg{
		TxData:        signBytes,
		Account:       account,
		Msg:           *encodedMsg,
		SignatureData: signatureData,
	}
	bz, err := json.Marshal(authMsg)
	if err != nil {
		panic(err)
		return iface.Rejected("failed to marshall AuthenticateMsg", err)
	}

	result, err := cwa.contractKeeper.Sudo(ctx, cwa.contractAddr, bz)
	if err != nil {
		return iface.Rejected("failed to sudo", err)
	}
	fmt.Println("result", result)
	// TODO: interpret the result
	if len(result) > 0 {
		return iface.Authenticated()
	}

	return iface.NotAuthenticated()
}

func (cwa CosmwasmAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (cwa CosmwasmAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	return iface.Confirm()
}

func (cwa CosmwasmAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	_, err := sdk.AccAddressFromBech32(string(data))
	if err != nil {
		return err
	}
	// TODO: check contract address length. Check contract exists?
	return nil
}

func (cwa CosmwasmAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}
