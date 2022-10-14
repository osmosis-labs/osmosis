package keeper_test

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v12/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// TestMintDenomMsg tests TypeMsgMint message is emitted on a successful mint
func (suite *KeeperTestSuite) TestMintDenomMsg() {
	// Create a denom
	suite.CreateDefaultDenom()

	for _, tc := range []struct {
		desc                  string
		amount                int64
		mintDenom             string
		admin                 string
		expectedErr           error
		expectedMessageEvents int
	}{
		{
			desc:        "denom does not exist",
			amount:      10,
			mintDenom:   "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:       suite.TestAccs[0].String(),
			expectedErr: types.ErrDenomDoesNotExist.Wrapf("denom: %s", "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos"),
		},
		{
			desc:                  "success case",
			amount:                10,
			mintDenom:             suite.defaultDenom,
			admin:                 suite.TestAccs[0].String(),
			expectedErr:           nil,
			expectedMessageEvents: 1,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test mint message
			_, err := suite.msgServer.Mint(sdk.WrapSDKContext(ctx), types.NewMsgMint(tc.admin, sdk.NewInt64Coin(tc.mintDenom, 10)))
			if tc.expectedErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().EqualError(err, tc.expectedErr.Error())
			}
			// Ensure current number and type of event is emitted
			suite.AssertEventEmitted(ctx, types.TypeMsgMint, tc.expectedMessageEvents)
		})
	}
}

// TestBurnDenomMsg tests TypeMsgBurn message is emitted on a successful burn
func (suite *KeeperTestSuite) TestBurnDenomMsg() {
	// Create a denom.
	suite.CreateDefaultDenom()
	// mint 10 default token for testAcc[0]
	suite.msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(suite.defaultDenom, 10)))

	for _, tc := range []struct {
		desc                  string
		amount                int64
		burnDenom             string
		admin                 string
		expectedErr           error
		expectedMessageEvents int
	}{
		{
			desc:        "denom does not exist",
			burnDenom:   "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:       suite.TestAccs[0].String(),
			expectedErr: types.ErrDenomDoesNotExist.Wrapf("denom: %s", "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos"),
		},
		{
			desc:                  "success case",
			burnDenom:             suite.defaultDenom,
			admin:                 suite.TestAccs[0].String(),
			expectedErr:           nil,
			expectedMessageEvents: 1,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test burn message
			_, err := suite.msgServer.Burn(sdk.WrapSDKContext(ctx), types.NewMsgBurn(tc.admin, sdk.NewInt64Coin(tc.burnDenom, 10)))
			if tc.expectedErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().EqualError(err, tc.expectedErr.Error())
			}
			// Ensure current number and type of event is emitted
			suite.AssertEventEmitted(ctx, types.TypeMsgBurn, tc.expectedMessageEvents)
		})
	}
}

// TestCreateDenomMsg tests TypeMsgCreateDenom message is emitted on a successful denom creation
func (suite *KeeperTestSuite) TestCreateDenomMsg() {
	defaultDenomCreationFee := types.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(50000000)))}
	for _, tc := range []struct {
		desc                  string
		denomCreationFee      types.Params
		subdenom              string
		expectedErr           error
		expectedMessageEvents int
	}{
		{
			desc:             "subdenom too long",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			expectedErr:      types.ErrSubdenomTooLong,
		},
		{
			desc:                  "success case: defaultDenomCreationFee",
			denomCreationFee:      defaultDenomCreationFee,
			subdenom:              "evmos",
			expectedErr:           nil,
			expectedMessageEvents: 1,
		},
	} {
		suite.SetupTest()
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tokenFactoryKeeper := suite.App.TokenFactoryKeeper
			ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))
			// Set denom creation fee in params
			tokenFactoryKeeper.SetParams(suite.Ctx, tc.denomCreationFee)
			// Test create denom message
			_, err := suite.msgServer.CreateDenom(sdk.WrapSDKContext(ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), tc.subdenom))

			if tc.expectedErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().EqualError(err, tc.expectedErr.Error())
			}
			// Ensure current number and type of event is emitted
			suite.AssertEventEmitted(ctx, types.TypeMsgCreateDenom, tc.expectedMessageEvents)
		})
	}
}

// TestChangeAdminDenomMsg tests TypeMsgChangeAdmin message is emitted on a successful admin change
func (suite *KeeperTestSuite) TestChangeAdminDenomMsg() {
	for _, tc := range []struct {
		desc                   string
		msgChangeAdmin         func(denom string) *types.MsgChangeAdmin
		expectedChangeAdminErr error
		expectedAdminIndex     int
		msgMint                func(denom string) *types.MsgMint
		expectedMintErr        error
		expectedMessageEvents  int
	}{
		{
			desc: "non-admins can't change the existing admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(suite.TestAccs[1].String(), denom, suite.TestAccs[2].String())
			},
			expectedChangeAdminErr: types.ErrUnauthorized,
			expectedAdminIndex:     0,
		},
		{
			desc: "success change admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(suite.TestAccs[0].String(), denom, suite.TestAccs[1].String())
			},
			expectedAdminIndex:     1,
			expectedChangeAdminErr: nil,
			expectedMessageEvents:  1,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(suite.TestAccs[1].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintErr: nil,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// setup test
			suite.SetupTest()
			ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))
			// Create a denom and mint
			res, err := suite.msgServer.CreateDenom(sdk.WrapSDKContext(ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
			suite.Require().NoError(err)
			testDenom := res.GetNewTokenDenom()
			_, err = suite.msgServer.Mint(sdk.WrapSDKContext(ctx), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(testDenom, 10)))

			// Test change admin message
			_, errChangeAdmin := suite.msgServer.ChangeAdmin(sdk.WrapSDKContext(ctx), tc.msgChangeAdmin(testDenom))
			if tc.expectedChangeAdminErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().EqualError(errChangeAdmin, tc.expectedChangeAdminErr.Error())
			}

			if tc.msgMint != nil {
				_, err := suite.msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), tc.msgMint(testDenom))
				if tc.expectedMintErr == nil {
					suite.Require().NoError(err)
				} else {
					suite.Require().EqualError(err, tc.expectedMintErr.Error())
				}
			}
			// Ensure current number and type of event is emitted
			suite.AssertEventEmitted(ctx, types.TypeMsgChangeAdmin, tc.expectedMessageEvents)
		})
	}
}

// TestSetDenomMetaDataMsg tests TypeMsgSetDenomMetadata message is emitted on a successful denom metadata change
func (suite *KeeperTestSuite) TestSetDenomMetaDataMsg() {
	// setup test
	suite.SetupTest()
	suite.CreateDefaultDenom()

	for _, tc := range []struct {
		desc                  string
		msgSetDenomMetadata   types.MsgSetDenomMetadata
		expectedErr           error
		expectedMessageEvents int
	}{
		{
			desc: "successful set denom metadata",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(suite.TestAccs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    suite.defaultDenom,
						Exponent: 0,
					},
					{
						Denom:    "uosmo",
						Exponent: 6,
					},
				},
				Base:    suite.defaultDenom,
				Display: "uosmo",
				Name:    "OSMO",
				Symbol:  "OSMO",
			}),
			expectedErr:           nil,
			expectedMessageEvents: 1,
		},
		{
			desc: "non existent factory denom name",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(suite.TestAccs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    fmt.Sprintf("factory/%s/litecoin", suite.TestAccs[0].String()),
						Exponent: 0,
					},
					{
						Denom:    "uosmo",
						Exponent: 6,
					},
				},
				Base:    fmt.Sprintf("factory/%s/litecoin", suite.TestAccs[0].String()),
				Display: "uosmo",
				Name:    "OSMO",
				Symbol:  "OSMO",
			}),
			expectedErr: types.ErrDenomDoesNotExist.Wrapf("denom: factory/%s/litecoin", suite.TestAccs[0].String()),
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test set denom metadata message
			_, err := suite.msgServer.SetDenomMetadata(sdk.WrapSDKContext(ctx), &tc.msgSetDenomMetadata)

			if tc.expectedErr != nil {
				suite.Require().EqualError(err, tc.expectedErr.Error())
			}
			// Ensure current number and type of event is emitted
			suite.AssertEventEmitted(ctx, types.TypeMsgSetDenomMetadata, tc.expectedMessageEvents)
		})
	}
}
