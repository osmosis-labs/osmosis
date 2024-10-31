package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	qcGetter := types.NewQueryClient
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdModuleBalance)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdModuleLockedAmount)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdAccountUnlockingCoins)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdAccountLockedPastTime)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdAccountLockedPastTimeNotUnlockingOnly)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdTotalLockedByDenom)
	cmd.AddCommand(
		GetCmdAccountUnlockableCoins(),
		GetCmdAccountLockedCoins(),
		GetCmdAccountUnlockedBeforeTime(),
		GetCmdAccountLockedPastTimeDenom(),
		GetCmdLockedByID(),
		GetCmdLockRewardReceiver(),
		GetCmdAccountLockedLongerDuration(),
		GetCmdAccountLockedLongerDurationNotUnlockingOnly(),
		GetCmdAccountLockedLongerDurationDenom(),
		GetCmdOutputLocksJson(),
		GetCmdSyntheticLockupsByLockupID(),
		GetCmdSyntheticLockupByLockupID(),
		GetCmdAccountLockedDuration(),
		GetCmdNextLockID(),
		osmocli.GetParams[*types.QueryParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)

	return cmd
}

// GetCmdModuleBalance returns full balance of the lockup module.
// Lockup module is where coins of locks are held.
// This includes locked balance and unlocked balance of the module.
func GetCmdModuleBalance() (*osmocli.QueryDescriptor, *types.ModuleBalanceRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "module-balance",
		Short: "Query module balance",
		Long:  `{{.Short}}`,
	}, &types.ModuleBalanceRequest{}
}

// GetCmdModuleLockedAmount returns locked balance of the module,
// which are all the tokens not unlocking + tokens that are not finished unlocking.
func GetCmdModuleLockedAmount() (*osmocli.QueryDescriptor, *types.ModuleLockedAmountRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "module-locked-amount",
		Short: "Query locked amount",
		Long:  `{{.Short}}`,
	}, &types.ModuleLockedAmountRequest{}
}

// GetCmdAccountUnlockableCoins returns unlockable coins which has finished unlocking.
// TODO: DELETE THIS + Actual query in subsequent PR
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

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountUnlockableCoins(cmd.Context(), &types.AccountUnlockableCoinsRequest{Owner: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAccountUnlockingCoins returns unlocking coins of a specific account.
func GetCmdAccountUnlockingCoins() (*osmocli.QueryDescriptor, *types.AccountUnlockingCoinsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "account-unlocking-coins",
		Short: "Query account's unlocking coins",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} account-unlocking-coins <address>`,
	}, &types.AccountUnlockingCoinsRequest{}
}

// GetCmdAccountLockedCoins returns locked coins that that are still in a locked state from the specified account.
func GetCmdAccountLockedCoins() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AccountLockedCoinsRequest](
		"account-locked-coins",
		"Query account's locked coins",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} account-locked-coins <address>
`, types.ModuleName, types.NewQueryClient)
}

// GetCmdAccountLockedPastTime returns locks of an account with unlock time beyond timestamp.
func GetCmdAccountLockedPastTime() (*osmocli.QueryDescriptor, *types.AccountLockedPastTimeRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "account-locked-pastime",
		Short: "Query locked records of an account with unlock time beyond timestamp",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} account-locked-pastime <address> <timestamp>
`,
	}, &types.AccountLockedPastTimeRequest{}
}

// GetCmdAccountLockedPastTimeNotUnlockingOnly returns locks of an account with unlock time beyond provided timestamp
// amongst the locks that are in the unlocking queue.
func GetCmdAccountLockedPastTimeNotUnlockingOnly() (*osmocli.QueryDescriptor, *types.AccountLockedPastTimeNotUnlockingOnlyRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "account-locked-pastime-not-unlocking",
		Short: "Query locked records of an account with unlock time beyond timestamp within not unlocking queue.",
		Long: `{{.Short}}
Timestamp is UNIX time in seconds.{{.ExampleHeader}}
{{.CommandPrefix}} account-locked-pastime-not-unlocking <address> <timestamp>
`,
	}, &types.AccountLockedPastTimeNotUnlockingOnlyRequest{}
}

// GetCmdAccountUnlockedBeforeTime returns locks with unlock time before the provided timestamp.
func GetCmdAccountUnlockedBeforeTime() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AccountUnlockedBeforeTimeRequest](
		"account-locked-beforetime",
		"Query account's unlocked records before specific time",
		`{{.Short}}
Timestamp is UNIX time in seconds.{{.ExampleHeader}}
{{.CommandPrefix}} account-locked-pastime <address> <timestamp>
`, types.ModuleName, types.NewQueryClient)
}

// GetCmdAccountLockedPastTimeDenom returns locks of an account whose unlock time is
// beyond given timestamp, and locks with the specified denom.
func GetCmdAccountLockedPastTimeDenom() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AccountLockedPastTimeDenomRequest](
		"account-locked-pastime-denom",
		"Query account's lock records by address, timestamp, denom",
		`{{.Short}}
Timestamp is UNIX time in seconds.{{.ExampleHeader}}
{{.CommandPrefix}} account-locked-pastime-denom <address> <timestamp> <denom>
`, types.ModuleName, types.NewQueryClient)
}

// GetCmdLockedByID returns lock by id.
func GetCmdLockedByID() *cobra.Command {
	q := osmocli.QueryDescriptor{
		Use:   "lock-by-id",
		Short: "Query account's lock record by id",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} lock-by-id 1`,
		QueryFnName: "LockedByID",
	}
	q.Long = osmocli.FormatLongDesc(q.Long, osmocli.NewLongMetadata(types.ModuleName).WithShort(q.Short))
	return osmocli.BuildQueryCli[*types.LockedRequest](&q, types.NewQueryClient)
}

// GetCmdLockRewardReceiver returns reward receiver for the given lock id
func GetCmdLockRewardReceiver() *cobra.Command {
	q := osmocli.QueryDescriptor{
		Use:   "lock-reward-receiver",
		Short: "Query lock's reward receiver",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} lock-reward-receiver 1`,
		QueryFnName: "LockRewardReceiver",
	}
	q.Long = osmocli.FormatLongDesc(q.Long, osmocli.NewLongMetadata(types.ModuleName).WithShort(q.Short))
	return osmocli.BuildQueryCli[*types.LockRewardReceiverRequest](&q, types.NewQueryClient)
}

// GetCmdNextLockID returns next lock id to be created.
func GetCmdNextLockID() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.NextLockIDRequest](
		"next-lock-id",
		"Query next lock id to be created",
		`{{.Short}}`, types.ModuleName, types.NewQueryClient)
}

// GetCmdSyntheticLockupsByLockupID returns synthetic lockups by lockup id.
// nolint: staticcheck
func GetCmdSyntheticLockupsByLockupID() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.SyntheticLockupsByLockupIDRequest](
		"synthetic-lockups-by-lock-id",
		"Query synthetic lockups by lockup id (is deprecated for synthetic-lockup-by-lock-id)",
		`{{.Short}}`, types.ModuleName, types.NewQueryClient)
}

// GetCmdSyntheticLockupByLockupID returns synthetic lockup by lockup id.
func GetCmdSyntheticLockupByLockupID() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.SyntheticLockupByLockupIDRequest](
		"synthetic-lockup-by-lock-id",
		"Query synthetic lock by underlying lockup id",
		`{{.Short}}`, types.ModuleName, types.NewQueryClient)
}

// GetCmdAccountLockedLongerDuration returns account locked records with longer duration.
func GetCmdAccountLockedLongerDuration() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AccountLockedLongerDurationRequest](
		"account-locked-longer-duration",
		"Query account locked records with longer duration",
		`{{.Short}}`, types.ModuleName, types.NewQueryClient)
}

// GetCmdAccountLockedLongerDuration returns account locked records with longer duration.
func GetCmdAccountLockedDuration() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AccountLockedDurationRequest](
		"account-locked-duration",
		"Query account locked records with a specific duration",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} account-locked-duration osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3 604800s`, types.ModuleName, types.NewQueryClient)
}

// GetCmdAccountLockedLongerDurationNotUnlockingOnly returns account locked records with longer duration from unlocking only queue.
func GetCmdAccountLockedLongerDurationNotUnlockingOnly() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AccountLockedLongerDurationNotUnlockingOnlyRequest](
		"account-locked-longer-duration-not-unlocking ",
		"Query account locked records with longer duration from unlocking only queue",
		`{{.Short}}`, types.ModuleName, types.NewQueryClient)
}

// GetCmdAccountLockedLongerDurationDenom returns account's locks for a specific denom
// with longer duration than the given duration.
func GetCmdAccountLockedLongerDurationDenom() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.AccountLockedLongerDurationDenomRequest](
		"account-locked-longer-duration-denom",
		"Query locked records for a denom with longer duration",
		`{{.Short}}`, types.ModuleName, types.NewQueryClient)
}

func GetCmdTotalLockedByDenom() (*osmocli.QueryDescriptor, *types.LockedDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "total-locked-of-denom",
		Short: "Query locked amount for a specific denom bigger then duration provided",
		Long: osmocli.FormatLongDescDirect(`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} total-locked-of-denom uosmo --min-duration=0s`, types.ModuleName),
		CustomFlagOverrides: map[string]string{
			"duration": FlagMinDuration,
		},
		Flags: osmocli.FlagDesc{OptionalFlags: []*pflag.FlagSet{FlagSetMinDuration()}},
	}, &types.LockedDenomRequest{}
}

// GetCmdOutputLocksJson outputs all locks into a file called lock_export.json.
func GetCmdOutputLocksJson() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output-all-locks <max lock ID>",
		Short: "output all locks into a json file",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Output all locks into a json file.
Example:
$ %s query lockup output-all-locks <max lock ID>
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

			maxLockID, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			// status
			const (
				doesnt_exist_status = iota
				unbonding_status
				bonded_status
			)

			type LockResult struct {
				Id            int
				Status        int // one of {doesnt_exist, }
				Denom         string
				Amount        osmomath.Int
				Address       string
				UnbondEndTime time.Time
			}
			queryClient := types.NewQueryClient(clientCtx)

			results := []LockResult{}
			for i := 0; i <= int(maxLockID); i++ {
				curLockResult := LockResult{Id: i}
				res, err := queryClient.LockedByID(cmd.Context(), &types.LockedRequest{LockId: uint64(i)})
				if err != nil {
					curLockResult.Status = doesnt_exist_status
					results = append(results, curLockResult)
					continue
				}
				// 1527019420 is hardcoded time well before launch, but well after year 1
				if res.Lock.EndTime.Before(time.Unix(1527019420, 0)) {
					curLockResult.Status = bonded_status
				} else {
					curLockResult.Status = unbonding_status
					curLockResult.UnbondEndTime = res.Lock.EndTime
					curLockResult.Denom = res.Lock.Coins[0].Denom
					curLockResult.Amount = res.Lock.Coins[0].Amount
					curLockResult.Address = res.Lock.Owner
				}
				results = append(results, curLockResult)
			}

			bz, err := json.Marshal(results)
			if err != nil {
				return err
			}
			err = os.WriteFile("lock_export.json", bz, 0o777)
			if err != nil {
				return err
			}

			fmt.Println("Writing to lock_export.json")
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
