package cli

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
		NewCreateGaugeCmd(),
		NewAddToGaugeCmd(),
		NewCreateGroupCmd(),
	)

	return cmd
}

// NewCreateGaugeCmd broadcasts a CreateGauge message.
func NewCreateGaugeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-gauge [lockup_denom] [reward] [poolId] [flags]",
		Short: "create a gauge to distribute rewards to users. For duration lock gauges set poolId = 0 and for all CL (no-lock) gauges set it to a CL poolId.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom := args[0]

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			var startTime time.Time
			timeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil {
				return err
			}
			if timeStr == "" { // empty start time
				startTime = time.Unix(0, 0)
			} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err == nil { // unix time
				startTime = time.Unix(timeUnix, 0)
			} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err == nil { // RFC time
				startTime = timeRFC
			} else { // invalid input
				return errors.New("invalid start time format")
			}

			epochs, err := cmd.Flags().GetUint64(FlagEpochs)
			if err != nil {
				return err
			}

			perpetual, err := cmd.Flags().GetBool(FlagPerpetual)
			if err != nil {
				return err
			}

			if perpetual {
				epochs = 1
			}

			duration, err := cmd.Flags().GetDuration(FlagDuration)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			var distributeTo lockuptypes.QueryCondition
			// if poolId is 0 it is a guaranteed lock gauge
			// if poolId is > 0 it is a guaranteed no-lock gauge
			if poolId == 0 {
				distributeTo = lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         denom,
					Duration:      duration,
					Timestamp:     time.Unix(0, 0), // XXX check
				}
			} else if poolId > 0 {
				distributeTo = lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.NoLock,
					Duration:      duration,
				}
			}

			msg := types.NewMsgCreateGauge(
				epochs == 1,
				clientCtx.GetFromAddress(),
				distributeTo,
				coins,
				startTime,
				epochs,
				poolId,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateGauge())
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewAddToGaugeCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgAddToGauge](&osmocli.TxCliDesc{
		Use:   "add-to-gauge",
		Short: "add coins to gauge to distribute more rewards to users",
	})
}

func NewCreateGroupCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgCreateGroup](&osmocli.TxCliDesc{
		Use:   "create-group",
		Short: "create a group in order to split incentives between pools",
	})
}

// NewCmdHandleCreateGroupsProposal implements a command handler for the group creation proposal transaction.
func NewCmdHandleCreateGroupsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-groups-proposal [pool-id-pairs] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a create groups proposal",
		Long: strings.TrimSpace(`Submit a create groups proposal.

Passing in pool-id-pairs separated by commas would be parsed automatically to a single set for a single group.
If a semicolon is presented, that would be parsed as pool IDs for separate group.
Don't forget the single quotes around the pool IDs!
Ex) create-groups-proposal '1,2;3,4,5;6,7' ->
Group 1: Pool IDs 1, 2
Group 2: Pool IDs 3, 4, 5
Group 3: Pool IDs 6, 7

		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseCreateGroupArgToContent(cmd, args[0])
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

	return cmd
}

func parseCreateGroupArgToContent(cmd *cobra.Command, arg string) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	createGroupRecords, err := ParseCreateGroupRecords(arg)
	if err != nil {
		return nil, err
	}

	content := &types.CreateGroupsProposal{
		Title:        title,
		Description:  description,
		CreateGroups: createGroupRecords,
	}

	return content, nil
}

func ParseCreateGroupRecords(arg string) ([]types.CreateGroup, error) {
	poolIds2DArray, err := osmocli.ParseStringTo2DArray(arg)
	if err != nil {
		return nil, err
	}

	createGroupRecords := []types.CreateGroup{}

	for _, poolIds := range poolIds2DArray {
		createGroupRecords = append(createGroupRecords, types.CreateGroup{
			PoolIds: poolIds,
		})
	}

	return createGroupRecords, nil
}
