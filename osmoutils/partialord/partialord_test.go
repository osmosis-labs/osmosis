package partialord_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/osmoutils/partialord"
)

func TestAPI(t *testing.T) {
	// begin block use case, we have a dozen modules, but only care about a couple orders.
	// In practice this will be gotten from some API, e.g. app.AllModuleNames()
	moduleNames := []string{
		"auth", "authz", "bank", "capabilities",
		"staking", "distribution", "epochs", "mint", "upgrades", "wasm", "ibc",
		"ibctransfers", "bech32ibc",
	}
	beginBlockOrd := partialord.NewPartialOrdering(moduleNames)
	beginBlockOrd.FirstElements("upgrades", "epochs", "capabilities")
	beginBlockOrd.After("ibctransfers", "ibc")
	beginBlockOrd.After("bech32ibc", "ibctransfers")
	beginBlockOrd.Before("mint", "distribution")
	// This is purely just to test last functionality, doesn't make sense in context
	beginBlockOrd.LastElements("auth", "authz", "wasm")

	totalOrd := beginBlockOrd.TotalOrdering()
	expTotalOrd := []string{
		"upgrades", "epochs", "capabilities",
		"bank", "staking", "mint", "ibc", "distribution", "ibctransfers", "bech32ibc",
		"auth", "authz", "wasm",
	}
	require.Equal(t, expTotalOrd, totalOrd)
}

func TestNonStandardAPIOrder(t *testing.T) {
	// This test uses direct ordering before First, and after Last
	names := []string{"A", "B", "C", "D", "E", "F", "G"}
	ord := partialord.NewPartialOrdering(names)
	ord.After("A", "C")
	ord.After("A", "D")
	ord.After("E", "B")
	// overrides the "A" after "C" & "A" after "D" constraints
	ord.FirstElements("A", "B", "C")
	expOrdering := []string{"A", "B", "C", "D", "E", "F", "G"}
	require.Equal(t, expOrdering, ord.TotalOrdering())

	ord.After("E", "D")
	expOrdering = []string{"A", "B", "C", "D", "F", "G", "E"}
	require.Equal(t, expOrdering, ord.TotalOrdering())

	ord.LastElements("G")
	ord.After("F", "E")
	expOrdering = []string{"A", "B", "C", "D", "E", "F", "G"}
	require.Equal(t, expOrdering, ord.TotalOrdering())
}
