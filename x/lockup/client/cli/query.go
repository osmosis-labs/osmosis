package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/lockup/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group lockup queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdModuleBalance(),
	)

	return cmd
}

// // Return locked balance of the module
// rpc ModuleLockedAmount(ModuleLockedAmountRequest) returns (ModuleLockedAmountResponse);

// // Returns whole unlockable coins which are not withdrawn yet
// rpc AccountUnlockableCoins(AccountUnlockableCoinsRequest) returns (AccountUnlockableCoinsResponse);
// // Return a locked coins that can't be withdrawn
// rpc AccountLockedCoins(AccountLockedCoinsRequest) returns (AccountLockedCoinsResponse);

// // Returns the total locks of an account whose unlock time is beyond timestamp
// rpc AccountLockedPastTime(AccountLockedPastTimeRequest) returns (AccountLockedPastTimeResponse);
// // Returns the total unlocks of an account whose unlock time is before timestamp
// rpc AccountUnlockedBeforeTime(AccountUnlockedBeforeTimeRequest) returns (AccountUnlockedBeforeTimeResponse);

// // Same as GetAccountLockedPastTime but denom specific
// rpc AccountLockedPastTimeDenom(AccountLockedPastTimeDenomRequest) returns (AccountLockedPastTimeDenomResponse);
// // Returns the length of the initial lock time when the lock was created
// rpc LockedByID(LockedRequest) returns (LockedResponse);

// // Returns account locked with duration longer than specified
// rpc AccountLockedLongerThanDuration(AccountLockedLongerDurationRequest) returns (AccountLockedLongerDurationResponse);
// // Returns account locked with duration longer than specified with specific denom
// rpc AccountLockedLongerThanDurationDenom(AccountLockedLongerDurationDenomRequest) returns (AccountLockedLongerDurationDenomResponse);

// GetCmdModuleBalance return full balance of the module
func GetCmdModuleBalance() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "module-balance ",
		Short: "Query module balance",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query module balance.

Example:
$ %s query lockup module-balance
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ModuleBalance(cmd.Context(), &types.ModuleBalanceRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
