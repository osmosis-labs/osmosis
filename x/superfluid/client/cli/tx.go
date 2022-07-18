package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/osmosis-labs/osmosis/v10/x/superfluid/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewSuperfluidDelegateCmd(),
		NewSuperfluidUndelegateCmd(),
		NewSuperfluidUnbondLockCmd(),
		// NewSuperfluidRedelegateCmd(),
		NewCmdSubmitSetSuperfluidAssetsProposal(),
		NewCmdSubmitRemoveSuperfluidAssetsProposal(),
		NewCmdLockAndSuperfluidDelegate(),
		NewCmdUnPoolWhitelistedPool(),
	)

	return cmd
}

// NewSuperfluidDelegateCmd broadcast MsgSuperfluidDelegate.
func NewSuperfluidDelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate [lock_id] [val_addr] [flags]",
		Short: "superfluid delegate a lock to a validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgSuperfluidDelegate(
				clientCtx.GetFromAddress(),
				uint64(lockId),
				valAddr,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSuperfluidUndelegateCmd broadcast MsgSuperfluidUndelegate.
func NewSuperfluidUndelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate [lock_id] [flags]",
		Short: "superfluid undelegate a lock from a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSuperfluidUndelegate(
				clientCtx.GetFromAddress(),
				uint64(lockId),
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSuperfluidUnbondLock broadcast MsgSuperfluidUndelegate and.
func NewSuperfluidUnbondLockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond-lock [lock_id] [flags]",
		Short: "unbond lock that has been superfluid staked",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSuperfluidUnbondLock(
				clientCtx.GetFromAddress(),
				uint64(lockId),
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSuperfluidRedelegateCmd broadcast MsgSuperfluidRedelegate
// func NewSuperfluidRedelegateCmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "redelegate [lock_id] [val_addr] [flags]",
// 		Short: "superfluid redelegate a lock to a new validator",
// 		Args:  cobra.ExactArgs(2),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			clientCtx, err := client.GetClientTxContext(cmd)
// 			if err != nil {
// 				return err
// 			}

// 			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

// 			lockId, err := strconv.Atoi(args[0])
// 			if err != nil {
// 				return err
// 			}

// 			valAddr, err := sdk.ValAddressFromBech32(args[1])
// 			if err != nil {
// 				return err
// 			}

// 			msg := types.NewMsgSuperfluidRedelegate(
// 				clientCtx.GetFromAddress(),
// 				uint64(lockId),
// 				valAddr,
// 			)

// 			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
// 		},
// 	}

// 	flags.AddTxFlagsToCmd(cmd)
// 	return cmd
// }

// NewCmdSubmitSetSuperfluidAssetsProposal implements a command handler for submitting a superfluid asset set proposal transaction.
func NewCmdSubmitSetSuperfluidAssetsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-superfluid-assets-proposal [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a superfluid asset set proposal",
		Long:  "Submit a superfluid asset set proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			content, err := parseSetSuperfluidAssetsArgsToContent(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(govcli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(govcli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(govcli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().String(FlagSuperfluidAssets, "", "The superfluid asset array")

	return cmd
}

// NewCmdSubmitRemoveSuperfluidAssetsProposal implements a command handler for submitting a superfluid asset remove proposal transaction.
func NewCmdSubmitRemoveSuperfluidAssetsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-superfluid-assets-proposal [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a superfluid asset remove proposal",
		Long:  "Submit a superfluid asset remove proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			content, err := parseRemoveSuperfluidAssetsArgsToContent(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(govcli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(govcli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(govcli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().String(FlagSuperfluidAssets, "", "The superfluid asset array")

	return cmd
}

func parseSetSuperfluidAssetsArgsToContent(cmd *cobra.Command) (govtypes.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagDescription)
	if err != nil {
		return nil, err
	}

	assetsStr, err := cmd.Flags().GetString(FlagSuperfluidAssets)
	if err != nil {
		return nil, err
	}

	assets := strings.Split(assetsStr, ",")

	superfluidAssets := []types.SuperfluidAsset{}
	for _, asset := range assets {
		superfluidAssets = append(superfluidAssets, types.SuperfluidAsset{
			Denom:     asset,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})
	}

	content := &types.SetSuperfluidAssetsProposal{
		Title:       title,
		Description: description,
		Assets:      superfluidAssets,
	}
	return content, nil
}

func parseRemoveSuperfluidAssetsArgsToContent(cmd *cobra.Command) (govtypes.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagDescription)
	if err != nil {
		return nil, err
	}

	assetsStr, err := cmd.Flags().GetString(FlagSuperfluidAssets)
	if err != nil {
		return nil, err
	}

	assets := strings.Split(assetsStr, ",")

	content := &types.RemoveSuperfluidAssetsProposal{
		Title:                 title,
		Description:           description,
		SuperfluidAssetDenoms: assets,
	}
	return content, nil
}

// NewCmdLockAndSuperfluidDelegate implements a command handler for simultaneous locking and superfluid delegation.
func NewCmdLockAndSuperfluidDelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-and-superfluid-delegate [tokens] [val_addr] [flags]",
		Short: "lock and superfluid delegate",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			sender := clientCtx.GetFromAddress()

			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgLockAndSuperfluidDelegate(sender, coins, valAddr)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdUnPoolWhitelistedPool implements a command handler for unpooling whitelisted pools.
func NewCmdUnPoolWhitelistedPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpool-whitelisted-pool [pool_id] [flags]",
		Short: "unpool whitelisted pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			sender := clientCtx.GetFromAddress()

			poolId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgUnPoolWhitelistedPool(sender, uint64(poolId))

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
