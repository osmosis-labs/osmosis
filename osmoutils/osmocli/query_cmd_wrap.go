package osmocli

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	grpc1 "github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
)

func QueryIndexCmd(moduleName string) *cobra.Command {
	return &cobra.Command{
		Use:                        moduleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", moduleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       indexRunCmd,
	}
}

func indexRunCmd(cmd *cobra.Command, args []string) error {
	usageTemplate := `Usage:{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}
  
{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	cmd.SetUsageTemplate(usageTemplate)
	return cmd.Help()
}

type QueryDescriptor struct {
	Use   string
	Short string
	Long  string

	HasPagination bool

	QueryFnName string
}

func SimpleQueryFromDescriptor[reqP proto.Message, querier any](desc QueryDescriptor, newQueryClientFn func(grpc1.ClientConn) querier) *cobra.Command {
	numArgs := ParseNumFields[reqP]()
	if desc.HasPagination {
		numArgs = numArgs - 1
	}
	cmd := &cobra.Command{
		Use:   desc.Use,
		Short: desc.Short,
		Long:  desc.Long,
		Args:  cobra.ExactArgs(numArgs),
		RunE: NewQueryLogicAllFieldsAsArgs[reqP](
			desc.QueryFnName, newQueryClientFn),
	}
	flags.AddQueryFlagsToCmd(cmd)
	if desc.HasPagination {
		cmdName := strings.Split(desc.Use, " ")[0]
		flags.AddPaginationFlagsToCmd(cmd, cmdName)
	}

	return cmd
}

// SimpleQueryCmd builds a query, for the common, simple case.
// It detects that the querier function name is the same as the ProtoMessage name,
// with just the "Query" and "Request" args chopped off.
// It expects all proto fields to appear as arguments, in order.
func SimpleQueryCmd[reqP proto.Message, querier any](use string, short string, long string,
	moduleName string, newQueryClientFn func(grpc1.ClientConn) querier) *cobra.Command {
	desc := QueryDescriptor{
		Use:           use,
		Short:         short,
		Long:          FormatLongDesc(long, NewLongMetadata(moduleName).WithShort(short)),
		HasPagination: ParseHasPagination[reqP](),
		QueryFnName:   ParseExpectedFnName[reqP](),
	}
	return SimpleQueryFromDescriptor[reqP](desc, newQueryClientFn)
}

type ParamGetter[reqP proto.Message, resP proto.Message] interface {
	Params(context.Context, reqP, ...grpc.CallOption) (resP, error)
}

func GetParams[reqP proto.Message, resP proto.Message, querier ParamGetter[reqP, resP]](
	moduleName string,
	newQueryClientFn func(grpc1.ClientConn) querier) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params [flags]",
		Short: fmt.Sprintf("Get the params for the x/%s module", moduleName),
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := newQueryClientFn(clientCtx)

			req := osmoutils.MakeNew[reqP]()
			res, err := queryClient.Params(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func callQueryClientFn[reqP proto.Message, querier any](ctx context.Context, fnName string, req reqP, q querier) (res proto.Message, err error) {
	qVal := reflect.ValueOf(q)
	method := qVal.MethodByName(fnName)
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(req),
	}
	results := method.Call(args)
	if len(results) != 2 {
		panic("We got something wrong")
	}
	if !results[1].IsNil() {
		//nolint:forcetypeassert
		err = results[1].Interface().(error)
		return res, err
	}
	//nolint:forcetypeassert
	res = results[0].Interface().(proto.Message)
	return res, nil
}

func NewQueryLogicAllFieldsAsArgs[reqP proto.Message, querier any](keeperFnName string,
	newQueryClientFn func(grpc1.ClientConn) querier) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		queryClient := newQueryClientFn(clientCtx)
		var req reqP

		req, err = ParseFieldsFromFlagsAndArgs[reqP](cmd.Flags(), args)
		if err != nil {
			return err
		}

		res, err := callQueryClientFn(cmd.Context(), keeperFnName, req, queryClient)
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}
}
