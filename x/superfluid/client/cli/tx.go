package cli

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/superfluid/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
		NewSuperfluidDelegateCmd(),
		NewSuperfluidUndelegateCmd(),
		NewSuperfluidUnbondLockCmd(),
		// NewSuperfluidRedelegateCmd(),
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

func NewSuperfluidUndelegateCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgSuperfluidUndelegate](&osmocli.TxCliDesc{
		Use:   "undelegate [lock_id] [flags]",
		Short: "superfluid undelegate a lock from a validator",
	})
}

func NewSuperfluidUnbondLockCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgSuperfluidUnbondLock](&osmocli.TxCliDesc{
		Use:   "unbond-lock [lock_id] [flags]",
		Short: "unbond lock that has been superfluid staked",
	})
}

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

func NewCmdUnPoolWhitelistedPool() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgUnPoolWhitelistedPool](&osmocli.TxCliDesc{
		Use:   "unpool-whitelisted-pool [pool_id] [flags]",
		Short: "unpool whitelisted pool",
	})
}

// NewCmdUpdateUnpoolWhitelistProposal defines the command to create a new update unpool whitelist proposal command.
func NewCmdUpdateUnpoolWhitelistProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-unpool-whitelist [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Update unpool whitelist proposal",
		Long: "This proposal will update the unpool whitelist if passed. " +
			"Every pool id must be valid. If the pool id is invalid, the proposal will not be submitted. " +
			"If the flag to overwrite is set, the whitelist is completely overridden. Otherwise, it is appended to the existing whitelist, having all duplicates removed.",
		Example: "osmosisd tx gov submit-proposal update-unpool-whitelist --pool-ids \"1, 2, 3\" --title \"Title\" --description \"Description\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			content, err := parseUpdateUnpoolWhitelistArgsToContent(cmd.Flags())
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
	cmd.Flags().String(FlagPoolIds, "", "The new pool id whitelist to set")
	cmd.Flags().Bool(FlagOverwrite, false, "The flag indicating whether to overwrite the whitelist or append to it")

	return cmd
}

func parseUpdateUnpoolWhitelistArgsToContent(flags *flag.FlagSet) (govtypes.Content, error) {
	title, err := flags.GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := flags.GetString(govcli.FlagDescription)
	if err != nil {
		return nil, err
	}

	poolIdsStr, err := flags.GetString(FlagPoolIds)
	if err != nil {
		return nil, err
	}

	poolIds, err := osmoutils.ParseUint64SliceFromString(poolIdsStr, ",")
	if err != nil {
		return nil, err
	}

	isOverwrite, err := flags.GetBool(FlagOverwrite)
	if err != nil {
		return nil, err
	}

	content := &types.UpdateUnpoolWhiteListProposal{
		Title:       title,
		Description: description,
		Ids:         poolIds,
		IsOverwrite: isOverwrite,
	}
	return content, nil
}
