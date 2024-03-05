package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	authenticatorkeeper "github.com/osmosis-labs/osmosis/v23/x/authenticator/keeper"
	authenticatortypes "github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
)

// CircuitBreakerDecorator routes transactions through appropriate ante handlers based on
// the IsCircuitBreakActive function.
type CircuitBreakerDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	auth                sdk.AnteHandler
	classic             sdk.AnteHandler
}

// NewCircuitBreakerDecorator creates a new instance of CircuitBreakerDecorator with the provided parameters.
func NewCircuitBreakerDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	auth sdk.AnteHandler,
	classic sdk.AnteHandler,
) CircuitBreakerDecorator {
	return CircuitBreakerDecorator{
		authenticatorKeeper: authenticatorKeeper,
		auth:                auth,
		classic:             classic,
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
	if active, _ := IsCircuitBreakActive(ctx, tx, ad.authenticatorKeeper); active {
		// Return and call the AnteHandle function on all the original decorators.
		return ad.classic(ctx, tx, simulate)
	}

	// Return and call the AnteHandle function on all the authenticator decorators.
	return ad.auth(ctx, tx, simulate)
}

// IsCircuitBreakActive checks if smart account are active and if there is a
// selected authenticator, the function will return true is the circuit breaker is active.
func IsCircuitBreakActive(
	ctx sdk.Context,
	tx sdk.Tx,
	authenticatorKeeper *authenticatorkeeper.Keeper,
) (bool, authenticatortypes.AuthenticatorTxOptions) {
	authenticatorParams := authenticatorKeeper.GetParams(ctx)
	if !authenticatorParams.AreSmartAccountsActive {
		return true, nil
	}

	extTx, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		return true, nil
	}

	// Get the selected authenticator options from the transaction.
	txOptions := authenticatorKeeper.GetAuthenticatorExtension(extTx.GetNonCriticalExtensionOptions())
	if txOptions == nil {
		return true, nil
	}

	return false, txOptions
}
