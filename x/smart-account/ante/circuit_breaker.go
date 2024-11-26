package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	smartaccountkeeper "github.com/osmosis-labs/osmosis/v27/x/smart-account/keeper"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

// CircuitBreakerDecorator routes transactions through appropriate ante handlers based on
// the IsCircuitBreakActive function.
type CircuitBreakerDecorator struct {
	smartAccountKeeper           *smartaccountkeeper.Keeper
	authenticatorAnteHandlerFlow sdk.AnteHandler
	originalAnteHandlerFlow      sdk.AnteHandler
}

// NewCircuitBreakerDecorator creates a new instance of CircuitBreakerDecorator with the provided parameters.
func NewCircuitBreakerDecorator(
	smartAccountKeeper *smartaccountkeeper.Keeper,
	auth sdk.AnteHandler,
	classic sdk.AnteHandler,
) CircuitBreakerDecorator {
	return CircuitBreakerDecorator{
		smartAccountKeeper:           smartAccountKeeper,
		authenticatorAnteHandlerFlow: auth,
		originalAnteHandlerFlow:      classic,
	}
}

// AnteHandle checks if a tx is a smart account tx and routes it through the correct series of ante handlers.
func (ad CircuitBreakerDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Check that the authenticator flow is active
	if active, _ := IsCircuitBreakActive(ctx, tx, ad.smartAccountKeeper); active {
		// Return and call the AnteHandle function on all the original decorators.
		return ad.originalAnteHandlerFlow(ctx, tx, simulate)
	}

	// Return and call the AnteHandle function on all the authenticator decorators.
	return ad.authenticatorAnteHandlerFlow(ctx, tx, simulate)
}

// IsCircuitBreakActive checks if smart account are active and if there is a
// selected authenticator, the function will return true is the circuit breaker is active.
func IsCircuitBreakActive(
	ctx sdk.Context,
	tx sdk.Tx,
	smartAccountKeeper *smartaccountkeeper.Keeper,
) (bool, smartaccounttypes.AuthenticatorTxOptions) {
	isSmartAccountActive := smartAccountKeeper.GetIsSmartAccountActive(ctx)
	// If the smart accounts are not active, the circuit breaker activates (i.e. return true).
	if !isSmartAccountActive {
		return true, nil
	}

	// Get the selected authenticator options from the transaction.
	return IsSelectedAuthenticatorTxExtensionMissing(tx, smartAccountKeeper)
}

// IsSelectedAuthenticatorTxExtensionMissing checks to see if the transaction has the correct
// extension, it returns false if we continue to the authenticator flow.
func IsSelectedAuthenticatorTxExtensionMissing(
	tx sdk.Tx,
	smartAccountKeeper *smartaccountkeeper.Keeper,
) (bool, smartaccounttypes.AuthenticatorTxOptions) {
	extTx, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		return true, nil
	}

	// Get the selected authenticator options from the transaction.
	txOptions := smartAccountKeeper.GetAuthenticatorExtension(extTx.GetNonCriticalExtensionOptions())

	// Check if authenticator transaction options are present and there is at least 1 selected.
	if txOptions == nil || len(txOptions.GetSelectedAuthenticators()) < 1 {
		return true, nil
	}

	// Return false with the txOptions if there are authenticator transaction options.
	return false, txOptions
}
