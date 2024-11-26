package types_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

func TestMsgSetHotRoutes(t *testing.T) {
	validStepSize := osmomath.NewInt(1_000_000)
	invalidStepSize := osmomath.NewInt(0)
	cases := []struct {
		description string
		admin       string
		hotRoutes   []types.TokenPairArbRoutes
		pass        bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			[]types.TokenPairArbRoutes{},
			false,
		},
		{
			"Valid message (no arb routes)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{},
			true,
		},
		{
			"Invalid message (nil hot routes)",
			createAccount().String(),
			nil,
			false,
		},
		{
			"Valid message (with arb routes)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			true,
		},
		{
			"Invalid message (mismatched arb denoms)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "eth",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			false,
		},
		{
			"Invalid message (with duplicate arb routes)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			false,
		},
		{
			"Invalid message (with missing trade)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
									TokenOut: "Juno",
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			false,
		},
		{
			"Invalid message (with invalid route length)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			false,
		},
		{
			"Valid message (with multiple routes)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
						{
							Trades: []types.Trade{
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     5,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Juno",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			true,
		},
		{
			"Invalid message (with invalid route hops)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: "Atom",
								},
								{
									Pool:     10,
									TokenIn:  "Akash",
									TokenOut: types.OsmosisDenomination,
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  "Atom",
					TokenOut: "Juno",
				},
			},
			false,
		},
		{
			"Invalid message (unset step size)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: "Atom",
								},
								{
									Pool:     10,
									TokenIn:  "Akash",
									TokenOut: types.OsmosisDenomination,
								},
							},
						},
					},
					TokenIn:  "Atom",
					TokenOut: "Juno",
				},
			},
			false,
		},
		{
			"Invalid message (invalid step size)",
			createAccount().String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Atom",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: "Atom",
								},
								{
									Pool:     10,
									TokenIn:  "Akash",
									TokenOut: types.OsmosisDenomination,
								},
							},
							StepSize: invalidStepSize,
						},
					},
					TokenIn:  "Atom",
					TokenOut: "Juno",
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := types.NewMsgSetHotRoutes(tc.admin, tc.hotRoutes)
			err := msg.ValidateBasic()
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgSetDeveloperAccount(t *testing.T) {
	cases := []struct {
		description string
		admin       string
		developer   string
		pass        bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			createAccount().String(),
			false,
		},
		{
			"Invalid message (invalid developer)",
			createAccount().String(),
			"developer",
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			createAccount().String(),
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := types.NewMsgSetDeveloperAccount(tc.admin, tc.developer)
			err := msg.ValidateBasic()
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgSetMaxPoolPointsPerTx(t *testing.T) {
	cases := []struct {
		description        string
		admin              string
		maxPoolPointsPerTx uint64
		pass               bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
		},
		{
			"Invalid message (too few max pool points per tx)",
			createAccount().String(),
			0,
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			1,
			true,
		},
		{
			"Invalid message (too many max pool points per tx)",
			createAccount().String(),
			types.MaxPoolPointsPerTx + 1,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := types.NewMsgSetMaxPoolPointsPerTx(tc.admin, tc.maxPoolPointsPerTx)
			err := msg.ValidateBasic()
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgSetMaxPoolPointsPerBlock(t *testing.T) {
	cases := []struct {
		description           string
		admin                 string
		maxPoolPointsPerBlock uint64
		pass                  bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
		},
		{
			"Invalid message (0 max pool points per block)",
			createAccount().String(),
			0,
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			10,
			true,
		},
		{
			"Invalid message (too many max pool points per block)",
			createAccount().String(),
			types.MaxPoolPointsPerBlock + 1,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := types.NewMsgSetMaxPoolPointsPerBlock(tc.admin, tc.maxPoolPointsPerBlock)
			err := msg.ValidateBasic()
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgSetPoolTypeInfo(t *testing.T) {
	cases := []struct {
		description    string
		admin          string
		infoByPoolType types.InfoByPoolType
		pass           bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			types.InfoByPoolType{
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Stable:       types.StablePoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{},
			},
			false,
		},
		{
			"Invalid message (invalid pool weights for balancer)",
			createAccount().String(),
			types.InfoByPoolType{
				Balancer:     types.BalancerPoolInfo{Weight: 0},
				Stable:       types.StablePoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{},
			},
			false,
		},
		{
			"Invalid message (invalid pool info for cosmwasm)",
			createAccount().String(),
			types.InfoByPoolType{
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Stable:       types.StablePoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm: types.CosmwasmPoolInfo{
					WeightMaps: []types.WeightMap{
						{
							ContractAddress: "contractAddress",
							Weight:          1,
						},
					},
				},
			},
			false,
		},
		{
			"Invalid message (invalid pool info for concentrated)",
			createAccount().String(),
			types.InfoByPoolType{
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Stable:       types.StablePoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{},
			},
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			types.InfoByPoolType{
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Stable:       types.StablePoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm: types.CosmwasmPoolInfo{
					WeightMaps: []types.WeightMap{
						{
							ContractAddress: createAccount().String(),
							Weight:          1,
						},
					},
				},
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := types.NewMsgSetPoolTypeInfo(tc.admin, tc.infoByPoolType)
			err := msg.ValidateBasic()
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgSetBaseDenoms(t *testing.T) {
	cases := []struct {
		description string
		admin       string
		baseDenoms  []types.BaseDenom
		pass        bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			[]types.BaseDenom{},
			false,
		},
		{
			"Invalid message (empty base denoms)",
			createAccount().String(),
			[]types.BaseDenom{},
			false,
		},
		{
			"Invalid message (nil base denoms list)",
			createAccount().String(),
			nil,
			false,
		},
		{
			"Invalid message (base denoms does not start with osmosis)",
			createAccount().String(),
			[]types.BaseDenom{
				{
					Denom:    "Atom",
					StepSize: osmomath.NewInt(10),
				},
			},
			false,
		},
		{
			"Invalid message (invalid step size)",
			createAccount().String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(0),
				},
			},
			false,
		},
		{
			"Invalid message (duplicate base denoms)",
			createAccount().String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(1),
				},
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(1),
				},
			},
			false,
		},
		{
			"Valid message (single denom)",
			createAccount().String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(1),
				},
			},
			true,
		},
		{
			"Valid message (multiple denoms)",
			createAccount().String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(1),
				},
				{
					Denom:    "Atom",
					StepSize: osmomath.NewInt(1),
				},
				{
					Denom:    "testDenom",
					StepSize: osmomath.NewInt(1),
				},
			},
			true,
		},
		{
			"Invalid message (multiple denoms with a single unset denom)",
			createAccount().String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(1),
				},
				{
					Denom:    "Atom",
					StepSize: osmomath.NewInt(1),
				},
				{
					Denom: "testDenom",
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := types.NewMsgSetBaseDenoms(tc.admin, tc.baseDenoms)
			err := msg.ValidateBasic()
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func createAccount() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	return sdk.AccAddress(pk.Address())
}
