package v10

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v9/app/keepers"
)

func CreateUpgradeHandler(
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		AgreedUponIrregularStateChange(ctx, keepers)
		return fromVM, nil
	}
}

type TransferFromAddress struct {
	Addr sdk.AccAddress
	// proof string
}

var TransferFromAddresses = []TransferFromAddress{}

func init() {
	// (1) osmo1hq8tlgq0kqz9e56532zghdhz7g8gtjymdltqer
	// firestake address
	// https://www.mintscan.io/cosmos/txs/8841476422795A83014F55C2C8915E21635D02011B3BA0CA92020893F5ED30DA
	addr, err := sdk.AccAddressFromBech32("osmo1hq8tlgq0kqz9e56532zghdhz7g8gtjymdltqer")
	if err != nil {
		panic(err)
	}
	TransferFromAddresses = append(TransferFromAddresses, TransferFromAddress{addr})
	// (2) osmo1v44mqmhvtn8cw373xv0hw6npddccnr70lqsk9s
	// firestake address
	// https://www.mintscan.io/cosmos/txs/D6187E67AAA4A16A414D41D8FADD35B1B082F50A548989C7FD9B083D2CAEB0C0
	addr, err = sdk.AccAddressFromBech32("osmo1v44mqmhvtn8cw373xv0hw6npddccnr70lqsk9s")
	if err != nil {
		panic(err)
	}
	TransferFromAddresses = append(TransferFromAddresses, TransferFromAddress{addr})
	// (3) osmo10t26acjmemggsahq6uvyucm4tj3z0mhz23ljh2
	// https://www.mintscan.io/cosmos/txs/0F536D1FAB700363B8A3EE47431BA7E7D80A40F55976C9B80A9F6984C8D0198A
	addr, err = sdk.AccAddressFromBech32("osmo10t26acjmemggsahq6uvyucm4tj3z0mhz23ljh2")
	if err != nil {
		panic(err)
	}
	TransferFromAddresses = append(TransferFromAddresses, TransferFromAddress{addr})
	// (4) osmo18qx59wy8s3ytax3e0akna934e86mw776vlzjtq
	// https://www.mintscan.io/cosmos/txs/0929B33C9F6368F10652D63218DC0E6B5AF8B0F986D041900AFDFC1B5EAC040D
	addr, err = sdk.AccAddressFromBech32("osmo18qx59wy8s3ytax3e0akna934e86mw776vlzjtq")
	if err != nil {
		panic(err)
	}
	TransferFromAddresses = append(TransferFromAddresses, TransferFromAddress{addr})
}

// Validators by choosing this binary and through explicit off-chain signalling
// have chosen the approach of doing an irregular state change.
// The change being, transferring all liquid funds from consenting addresses to
// to the recovery address.
// This is done with cryptographic approval of this action.
func AgreedUponIrregularStateChange(ctx sdk.Context, keepers *keepers.AppKeepers) {
	for _, addr := range TransferFromAddresses {
		ForceTransferAllTokens(ctx, keepers, addr.Addr, RecoveryAddress)
	}
}

// ForceTransferAllTokens from address `from` to address `to`.
// Assumes neither of `from` or `to` are a module account, from is non-nil.
func ForceTransferAllTokens(ctx sdk.Context, keepers *keepers.AppKeepers, from sdk.AccAddress, to sdk.AccAddress) {
	balances := keepers.BankKeeper.GetAllBalances(ctx, from)
	err := keepers.BankKeeper.SendCoins(ctx, from, to, balances)
	if err != nil {
		panic(err)
	}
}

func VerifyTx() {}
