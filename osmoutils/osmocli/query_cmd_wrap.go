package osmocli

import (
	"context"
	"fmt"
	"reflect"

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

func callQueryClientFn[reqP proto.Message, resP proto.Message, querier any](ctx context.Context, fnName string, req reqP, q querier) (res resP, err error) {
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
	res = results[0].Interface().(resP)
	return res, nil
}

func ParseFieldsFromArgs[reqP proto.Message](args []string) (reqP, error) {
	req := osmoutils.MakeNew[reqP]()
	v := reflect.ValueOf(req).Elem()
	t := v.Type()
	if len(args) != t.NumField() {
		return req, fmt.Errorf("Incorrect number of arguments, expected %d got %d", t.NumField(), len(args))
	}

	// Iterate over the fields in the struct
	for i := 0; i < t.NumField(); i++ {
		err := ParseField(v, t, i, args[i])
		if err != nil {
			return req, err
		}
	}
	return req, nil
}

func NewQueryLogicAllFieldsAsArgs[reqP proto.Message, resP proto.Message, querier any](keeperFnName string,
	newQueryClientFn func(grpc1.ClientConn) querier) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		queryClient := newQueryClientFn(clientCtx)
		var req reqP

		req, err = ParseFieldsFromArgs[reqP](args)
		if err != nil {
			return err
		}

		res, err := callQueryClientFn[reqP, resP](cmd.Context(), keeperFnName, req, queryClient)
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}
}
