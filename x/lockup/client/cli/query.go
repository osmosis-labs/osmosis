package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
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
		GetCmdModuleLockedAmount(),
		GetCmdAccountUnlockableCoins(),
		GetCmdAccountUnlockingCoins(),
		GetCmdAccountLockedCoins(),
		GetCmdAccountLockedPastTime(),
		GetCmdAccountLockedPastTimeNotUnlockingOnly(),
		GetCmdAccountUnlockedBeforeTime(),
		GetCmdAccountLockedPastTimeDenom(),
		GetCmdLockedByID(),
		GetCmdAccountLockedLongerDuration(),
		GetCmdAccountLockedLongerDurationNotUnlockingOnly(),
		GetCmdAccountLockedLongerDurationDenom(),
	)

	return cmd
}

// GetCmdModuleBalance returns full balance of the module
func GetCmdModuleBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-balance",
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

// GetCmdModuleLockedAmount returns locked balance of the module
func GetCmdModuleLockedAmount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-locked-amount",
		Short: "Query module locked amount",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query module locked amount.

Example:
$ %s query lockup module-locked-amount
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

			res, err := queryClient.ModuleLockedAmount(cmd.Context(), &types.ModuleLockedAmountRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountUnlockableCoins returns unlockable coins which are not withdrawn yet
func GetCmdAccountUnlockableCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-unlockable-coins <address>",
		Short: "Query account's unlockable coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account's unlockable coins.

Example:
$ %s query lockup account-unlockable-coins <address>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountUnlockableCoins(cmd.Context(), &types.AccountUnlockableCoinsRequest{Owner: addr})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountUnlockingCoins returns unlocking coins
func GetCmdAccountUnlockingCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-unlocking-coins <address>",
		Short: "Query account's unlocking coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account's unlocking coins.

Example:
$ %s query lockup account-unlocking-coins <address>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountUnlockingCoins(cmd.Context(), &types.AccountUnlockingCoinsRequest{Owner: addr})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountLockedCoins returns locked coins that can't be withdrawn
func GetCmdAccountLockedCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-coins <address>",
		Short: "Query account's locked coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account's locked coins.

Example:
$ %s query lockup account-locked-coins <address>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountLockedCoins(cmd.Context(), &types.AccountLockedCoinsRequest{Owner: addr})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountLockedPastTime returns locked records of an account with unlock time beyond timestamp
func GetCmdAccountLockedPastTime() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-pasttime <address> <timestamp>",
		Short: "Query locked records of an account with unlock time beyond timestamp",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query locked records of an account with unlock time beyond timestamp.

Example:
$ %s query lockup account-locked-pasttime <address> <timestamp>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			i, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				panic(err)
			}
			timestamp := time.Unix(i, 0)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountLockedPastTime(cmd.Context(), &types.AccountLockedPastTimeRequest{Owner: addr, Timestamp: timestamp})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountLockedPastTimeNotUnlockingOnly returns locked records of an account with unlock time beyond timestamp within not unlocking queue
func GetCmdAccountLockedPastTimeNotUnlockingOnly() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-pasttime-not-unlocking <address> <timestamp>",
		Short: "Query locked records of an account with unlock time beyond timestamp within not unlocking queue",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query locked records of an account with unlock time beyond timestamp within not unlocking queue.

Example:
$ %s query lockup account-locked-pasttime-not-unlocking <address> <timestamp>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			i, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				panic(err)
			}
			timestamp := time.Unix(i, 0)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountLockedPastTimeNotUnlockingOnly(cmd.Context(), &types.AccountLockedPastTimeNotUnlockingOnlyRequest{Owner: addr, Timestamp: timestamp})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountUnlockedBeforeTime returns unlocked records with unlock time before timestamp
func GetCmdAccountUnlockedBeforeTime() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-beforetime <address> <timestamp>",
		Short: "Query account's unlocked records before specific time",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account's the total unlocked records with unlock time before timestamp.

Example:
$ %s query lockup account-locked-pasttime <address> <timestamp>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			i, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				panic(err)
			}
			timestamp := time.Unix(i, 0)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountUnlockedBeforeTime(cmd.Context(), &types.AccountUnlockedBeforeTimeRequest{Owner: addr, Timestamp: timestamp})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountLockedPastTimeDenom returns lock records by address, timestamp, denom
func GetCmdAccountLockedPastTimeDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-pasttime-denom <address> <timestamp> <denom>",
		Short: "Query account's lock records by address, timestamp, denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account's lock records by address, timestamp, denom.

Example:
$ %s query lockup account-locked-pasttime-denom <address> <timestamp> <denom>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			i, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				panic(err)
			}
			timestamp := time.Unix(i, 0)

			denom := args[2]

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountLockedPastTimeDenom(cmd.Context(), &types.AccountLockedPastTimeDenomRequest{Owner: addr, Timestamp: timestamp, Denom: denom})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdLockedByID returns lock record by id
func GetCmdLockedByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-by-id <id>",
		Short: "Query account's lock record by id",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account's lock record by id.

Example:
$ %s query lockup lock-by-id <id>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				panic(err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LockedByID(cmd.Context(), &types.LockedRequest{LockId: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountLockedLongerDuration returns account locked records with longer duration
func GetCmdAccountLockedLongerDuration() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-longer-duration <address> <duration>",
		Short: "Query account locked records with longer duration",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account locked records with longer duration.

Example:
$ %s query lockup account-locked-longer-duration <address> <duration>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountLockedLongerDuration(cmd.Context(), &types.AccountLockedLongerDurationRequest{Owner: addr, Duration: duration})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountLockedLongerDurationNotUnlockingOnly returns account locked records with longer duration from unlocking only queue
func GetCmdAccountLockedLongerDurationNotUnlockingOnly() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-longer-duration-not-unlocking <address> <duration>",
		Short: "Query account locked records with longer duration from unlocking only queue",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account locked records with longer duration from unlocking only queue.

Example:
$ %s query lockup account-locked-longer-duration-not-unlocking <address> <duration>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountLockedLongerDurationNotUnlockingOnly(cmd.Context(), &types.AccountLockedLongerDurationNotUnlockingOnlyRequest{Owner: addr, Duration: duration})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountLockedLongerDurationDenom returns account's locked records for a denom with longer duration
func GetCmdAccountLockedLongerDurationDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-locked-longer-duration-denom <address> <duration> <denom>",
		Short: "Query locked records for a denom with longer duration",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account's locked records for a denom with longer duration.

Example:
$ %s query lockup account-locked-pasttime <address> <duration> <denom>
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			denom := args[2]

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountLockedLongerDurationDenom(cmd.Context(), &types.AccountLockedLongerDurationDenomRequest{Owner: addr, Duration: duration, Denom: denom})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
