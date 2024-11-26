package testutils

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
)

var (
	_ authenticator.Authenticator = &TestingAuthenticator{}
)

type ApproveOn int

const (
	Always ApproveOn = iota
	Never
)

type (
	TestingAuthenticatorData struct{}
	TestingAuthenticator     struct {
		Approve        ApproveOn
		GasConsumption int
		BlockAddition  bool
		BlockRemoval   bool
		Confirm        ApproveOn
	}
)

func (t TestingAuthenticator) Type() string {
	var when string
	if t.Approve == Always {
		when = "Always"
	} else {
		when = "Never"
	}

	var confirm string
	if t.Confirm == Always {
		confirm = "Confirm"
	} else {
		confirm = "Block"
	}

	return "TestingAuthenticator" + when + confirm + fmt.Sprintf("GasConsumption%d", t.GasConsumption) + fmt.Sprintf("BlockAddition%t", t.BlockAddition) + fmt.Sprintf("BlockRemoval%t", t.BlockRemoval)
}

func (t TestingAuthenticator) StaticGas() uint64 {
	return uint64(t.GasConsumption)
}

func (t TestingAuthenticator) Initialize(config []byte) (authenticator.Authenticator, error) {
	return t, nil
}

func (t TestingAuthenticator) Authenticate(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	if t.Approve == Always {
		return nil
	} else {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "TestingAuthenticator authentication error")
	}
}

func (t TestingAuthenticator) Track(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	return nil
}

func (t TestingAuthenticator) ConfirmExecution(ctx sdk.Context, request authenticator.AuthenticationRequest) error {
	if t.Confirm == Always {
		return nil
	} else {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "TestingAuthenticator block")
	}
}

func (t TestingAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	if t.BlockAddition {
		return fmt.Errorf("authenticator could not be added")
	}
	return nil
}

func (t TestingAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	if t.BlockRemoval {
		return fmt.Errorf("authenticator could not be removed")
	}
	return nil
}
