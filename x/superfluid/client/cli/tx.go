package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
		NewSuperfluidDelegateCmd(),
		NewSuperfluidUndelegateCmd(),
		NewSuperfluidUnbondLockCmd(),
		NewSuperfluidUndelegateAndUnbondLockCmd(),
		// NewSuperfluidRedelegateCmd(),
		NewCmdLockAndSuperfluidDelegate(),
		NewCmdUnPoolWhitelistedPool(),
		NewUnbondConvertAndStake(),
	)
	osmocli.AddTxCmd(cmd, NewCreateFullRangePositionAndSuperfluidDelegateCmd)
	osmocli.AddTxCmd(cmd, NewAddToConcentratedLiquiditySuperfluidPositionCmd)
	osmocli.AddTxCmd(cmd, NewUnlockAndMigrateSharesToFullRangeConcentratedPositionCmd)

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

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

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
		Use:   "undelegate",
		Short: "superfluid undelegate a lock from a validator",
	})
}

func NewSuperfluidUnbondLockCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgSuperfluidUnbondLock](&osmocli.TxCliDesc{
		Use:   "unbond-lock",
		Short: "unbond lock that has been superfluid staked",
	})
}

func NewSuperfluidUndelegateAndUnbondLockCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgSuperfluidUndelegateAndUnbondLock](&osmocli.TxCliDesc{
		Use:   "undelegate-and-unbond-lock",
		Short: "superfluid undelegate and unbond lock for the given amount of coin",
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
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseSetSuperfluidAssetsArgsToContent(cmd)
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
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
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseRemoveSuperfluidAssetsArgsToContent(cmd)
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
	cmd.Flags().String(FlagSuperfluidAssets, "", "The superfluid asset array")

	return cmd
}

func parseSetSuperfluidAssetsArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
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
		var assetType types.SuperfluidAssetType
		if strings.HasPrefix(asset, gammtypes.GAMMTokenPrefix) {
			assetType = types.SuperfluidAssetTypeLPShare
		} else if strings.HasPrefix(asset, cltypes.ConcentratedLiquidityTokenPrefix) {
			assetType = types.SuperfluidAssetTypeConcentratedShare
		} else {
			return nil, fmt.Errorf("Invalid asset prefix: %s", asset)
		}

		superfluidAssets = append(superfluidAssets, types.SuperfluidAsset{
			Denom:     asset,
			AssetType: assetType,
		})
	}

	content := &types.SetSuperfluidAssetsProposal{
		Title:       title,
		Description: description,
		Assets:      superfluidAssets,
	}
	return content, nil
}

func parseRemoveSuperfluidAssetsArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
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

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

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
		Use:   "unpool-whitelisted-pool",
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
		Example: "osmosisd tx gov submit-proposal update-unpool-whitelist --pool-ids \"1, 2, 3\" --title \"Title\" --summary \"Description\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseUpdateUnpoolWhitelistArgsToContent(cmd.Flags())
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
	cmd.Flags().String(FlagPoolIds, "", "The new pool id whitelist to set")
	cmd.Flags().Bool(FlagOverwrite, false, "The flag indicating whether to overwrite the whitelist or append to it")

	return cmd
}

func NewCreateFullRangePositionAndSuperfluidDelegateCmd() (*osmocli.TxCliDesc, *types.MsgCreateFullRangePositionAndSuperfluidDelegate) {
	return &osmocli.TxCliDesc{
		Use:     "create-full-range-position-and-sf-delegate",
		Short:   "creates a full range concentrated position and superfluid delegates it to the provided validator",
		Example: "create-full-range-position-and-sf-delegate 100000000uosmo,10000udai 45 --from val --chain-id osmosis-1",
	}, &types.MsgCreateFullRangePositionAndSuperfluidDelegate{}
}

func parseUpdateUnpoolWhitelistArgsToContent(flags *flag.FlagSet) (govtypesv1beta1.Content, error) {
	title, err := flags.GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := flags.GetString(govcli.FlagSummary)
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

func NewAddToConcentratedLiquiditySuperfluidPositionCmd() (*osmocli.TxCliDesc, *types.MsgAddToConcentratedLiquiditySuperfluidPosition) {
	return &osmocli.TxCliDesc{
		Use:     "add-to-superfluid-cl-position",
		Short:   "add to an existing superfluid staked concentrated liquidity position",
		Example: "add-to-superfluid-cl-position 10 1000000000uosmo 10000000uion",
	}, &types.MsgAddToConcentratedLiquiditySuperfluidPosition{}
}

func NewUnlockAndMigrateSharesToFullRangeConcentratedPositionCmd() (*osmocli.TxCliDesc, *types.MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition) {
	return &osmocli.TxCliDesc{
		Use:     "unlock-and-migrate-to-cl",
		Short:   "unlock and migrate gamm shares to full range concentrated position",
		Example: "unlock-and-migrate-cl 10 25000000000gamm/pool/2 1000000000uosmo,10000000uion",
	}, &types.MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition{}
}

func NewUnbondConvertAndStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unbond-convert-and-stake [lock-id] [valAddr] [min-amount-to-stake](optional) [shares-to-convert](optional)",
		Short:   "instantly unbond any locked gamm shares convert them into osmo and stake",
		Example: "unbond-convert-and-stake 10 osmo1xxx 100000uosmo",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			sender := clientCtx.GetFromAddress()
			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			valAddr := args[1]

			var minAmtToStake osmomath.Int
			// if user provided args for min amount to stake, use it. If not, use empty coin struct
			var sharesToConvert sdk.Coin
			if len(args) >= 3 {
				convertedInt, ok := osmomath.NewIntFromString(args[2])
				if !ok {
					return errors.New("Conversion for osmomath.Int failed")
				}
				minAmtToStake = convertedInt
				if len(args) == 4 {
					coins, err := sdk.ParseCoinNormalized(args[3])
					if err != nil {
						return err
					}
					sharesToConvert = coins
				}
			} else {
				minAmtToStake = osmomath.ZeroInt()
				sharesToConvert = sdk.Coin{}
			}

			msg := types.NewMsgUnbondConvertAndStake(sender, uint64(lockId), valAddr, minAmtToStake, sharesToConvert)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
