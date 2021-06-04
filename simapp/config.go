package simapp

// DefaultConfig returns a default configuration suitable for nearly all
// testing requirements.
// func DefaultConfig() network.Config {
// 	encCfg := app.MakeEncodingConfig()

// 	return network.Config{
// 		Codec:             encCfg.Marshaler,
// 		TxConfig:          encCfg.TxConfig,
// 		LegacyAmino:       encCfg.Amino,
// 		InterfaceRegistry: encCfg.InterfaceRegistry,
// 		AccountRetriever:  authtypes.AccountRetriever{},
// 		AppConstructor: func(val network.Validator) servertypes.Application {
// 			return app.NewOsmosisApp(
// 				val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
// 				encCfg,
// 				simapp.EmptyAppOptions{},
// 				baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
// 			)
// 		},
// 		GenesisState:    app.ModuleBasics.DefaultGenesis(encCfg.Marshaler),
// 		TimeoutCommit:   2 * time.Second,
// 		ChainID:         "osmosis-1",
// 		NumValidators:   1,
// 		BondDenom:       sdk.DefaultBondDenom,
// 		MinGasPrices:    fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
// 		AccountTokens:   sdk.TokensFromConsensusPower(1000),
// 		StakingTokens:   sdk.TokensFromConsensusPower(500),
// 		BondedTokens:    sdk.TokensFromConsensusPower(100),
// 		PruningStrategy: storetypes.PruningOptionNothing,
// 		CleanupDir:      true,
// 		SigningAlgo:     string(hd.Secp256k1Type),
// 		KeyringOptions:  []keyring.Option{},
// 	}
// }
