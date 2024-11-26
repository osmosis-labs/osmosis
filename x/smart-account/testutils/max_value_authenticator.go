package testutils

import (
	"encoding/json"
	"fmt"

	proto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
)

// This is a very naive implementation of an authenticator that tracks sends and blocks if the total amount sent is greater than 3_000
var _ authenticator.Authenticator = &MaxAmountAuthenticator{}

type MaxAmountAuthenticatorData struct {
	Amount osmomath.Int
}
type MaxAmountAuthenticator struct {
	KvStoreKey storetypes.StoreKey
}

func (m MaxAmountAuthenticator) StaticGas() uint64 {
	return 0
}

func (m MaxAmountAuthenticator) Type() string {
	return "MaxAmountAuthenticator"
}

func (m MaxAmountAuthenticator) Initialize(config []byte) (authenticator.Authenticator, error) {
	return m, nil
}

func (m MaxAmountAuthenticator) Authenticate(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	if request.Msg.TypeURL != "/cosmos.bank.v1beta1.MsgSend" {
		return nil
	}
	// unmarshal the message.value into the bank.MsgSend struct
	var send banktypes.MsgSend
	err := proto.Unmarshal(request.Msg.Value, &send)
	if err != nil {
		return err
	}
	if m.GetAmount(ctx).Add(send.Amount[0].Amount).GTE(osmomath.NewInt(3_000)) {
		return fmt.Errorf("total amount sent is greater than 3_000")
	}
	return nil
}

func (m MaxAmountAuthenticator) Track(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	return nil
}

func (m MaxAmountAuthenticator) ConfirmExecution(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	if request.Msg.TypeURL != "/cosmos.bank.v1beta1.MsgSend" {
		return nil
	}
	// unmarshal the message.value into the bank.MsgSend struct
	var send banktypes.MsgSend
	err := proto.Unmarshal(request.Msg.Value, &send)
	if err != nil {
		return nil
	}
	m.SetAmount(ctx, m.GetAmount(ctx).Add(send.Amount[0].Amount))
	return nil
}

func (m MaxAmountAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return nil
}

func (m MaxAmountAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return nil
}

// The following methods for MaxAmountAuthenticator are similar to the set and get value methods for StatefulAuthenticator but set and get an int
func (m MaxAmountAuthenticator) SetAmount(ctx sdk.Context, amount osmomath.Int) {
	kvStore := prefix.NewStore(ctx.KVStore(m.KvStoreKey), []byte(m.Type()))
	maxAmountData := MaxAmountAuthenticatorData{Amount: amount}
	newBz, _ := json.Marshal(maxAmountData)
	kvStore.Set([]byte("amount"), newBz)
}

func (m MaxAmountAuthenticator) GetAmount(ctx sdk.Context) osmomath.Int {
	kvStore := prefix.NewStore(ctx.KVStore(m.KvStoreKey), []byte(m.Type()))
	bz := kvStore.Get([]byte("amount")) // global value. On the real thing we may want the account
	var amountData MaxAmountAuthenticatorData
	err := json.Unmarshal(bz, &amountData)
	// if we can't unmarshal, we just assume it's 0
	if err != nil {
		return osmomath.NewInt(0)
	}
	return amountData.Amount
}
