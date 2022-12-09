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
)

func QueryIndexCmd(moduleName string) *cobra.Command {
	cmd := IndexCmd(moduleName)
	cmd.Short = fmt.Sprintf("Querying commands for the %s module", moduleName)
	return cmd
}

type QueryDescriptor struct {
	Use   string
	Short string
	Long  string

	HasPagination bool

	QueryFnName string

	Flags FlagDesc
	// Map of FieldName -> FlagName
	CustomFlagOverrides map[string]string
	// Map of FieldName -> CustomParseFn
	CustomFieldParsers map[string]CustomFieldParserFn
}

func SimpleQueryFromDescriptor[reqP proto.Message, querier any](desc QueryDescriptor, newQueryClientFn func(grpc1.ClientConn) querier) *cobra.Command {
	numArgs := ParseNumFields[reqP]() - len(desc.CustomFlagOverrides)
	if desc.HasPagination {
		numArgs = numArgs - 1
	}
	flagAdvice := FlagAdvice{
		HasPagination:       desc.HasPagination,
		CustomFlagOverrides: desc.CustomFlagOverrides,
		CustomFieldParsers:  desc.CustomFieldParsers,
	}.Sanitize()
	cmd := &cobra.Command{
		Use:   desc.Use,
		Short: desc.Short,
		Long:  desc.Long,
		Args:  cobra.ExactArgs(numArgs),
		RunE: NewQueryLogicAllFieldsAsArgs[reqP](
			flagAdvice, desc.QueryFnName, newQueryClientFn),
	}
	flags.AddQueryFlagsToCmd(cmd)
	AddFlags(cmd, desc.Flags)
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
		QueryFnName:   ParseExpectedQueryFnName[reqP](),
	}
	return SimpleQueryFromDescriptor[reqP](desc, newQueryClientFn)
}

func GetParams[reqP proto.Message, querier any](moduleName string,
	newQueryClientFn func(grpc1.ClientConn) querier) *cobra.Command {
	return SimpleQueryFromDescriptor[reqP](QueryDescriptor{
		Use:         "params [flags]",
		Short:       fmt.Sprintf("Get the params for the x/%s module", moduleName),
		QueryFnName: "Params",
	}, newQueryClientFn)
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

func NewQueryLogicAllFieldsAsArgs[reqP proto.Message, querier any](flagAdvice FlagAdvice, keeperFnName string,
	newQueryClientFn func(grpc1.ClientConn) querier) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		queryClient := newQueryClientFn(clientCtx)
		var req reqP

		req, err = ParseFieldsFromFlagsAndArgs[reqP](flagAdvice, cmd.Flags(), args)
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
