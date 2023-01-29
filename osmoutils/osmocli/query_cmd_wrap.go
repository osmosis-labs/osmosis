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
	"github.com/spf13/pflag"
)

// global variable set on index command.
// helps populate Longs, when not set in QueryDescriptor.
var lastQueryModuleName string

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

	ParseQuery func(args []string, flags *pflag.FlagSet) (proto.Message, error)

	ModuleName string
	numArgs    int
}

func QueryIndexCmd(moduleName string) *cobra.Command {
	cmd := IndexCmd(moduleName)
	cmd.Short = fmt.Sprintf("Querying commands for the %s module", moduleName)
	lastQueryModuleName = moduleName
	return cmd
}

func AddQueryCmd[Q proto.Message, querier any](cmd *cobra.Command, newQueryClientFn func(grpc1.ClientConn) querier, f func() (*QueryDescriptor, Q)) {
	desc, _ := f()
	subCmd := BuildQueryCli[Q](desc, newQueryClientFn)
	cmd.AddCommand(subCmd)
}

func (desc *QueryDescriptor) FormatLong(moduleName string) {
	desc.Long = FormatLongDesc(desc.Long, NewLongMetadata(moduleName).WithShort(desc.Short))
}

func prepareDescriptor[reqP proto.Message](desc *QueryDescriptor) {
	if !desc.HasPagination {
		desc.HasPagination = ParseHasPagination[reqP]()
	}
	if desc.QueryFnName == "" {
		desc.QueryFnName = ParseExpectedQueryFnName[reqP]()
	}
	if strings.Contains(desc.Long, "{") {
		if desc.ModuleName == "" {
			desc.ModuleName = lastQueryModuleName
		}
		desc.FormatLong(desc.ModuleName)
	}

	desc.numArgs = ParseNumFields[reqP]() - len(desc.CustomFlagOverrides)
	if desc.HasPagination {
		desc.numArgs = desc.numArgs - 1
	}
}

func BuildQueryCli[reqP proto.Message, querier any](desc *QueryDescriptor, newQueryClientFn func(grpc1.ClientConn) querier) *cobra.Command {
	prepareDescriptor[reqP](desc)
	if desc.ParseQuery == nil {
		desc.ParseQuery = func(args []string, fs *pflag.FlagSet) (proto.Message, error) {
			flagAdvice := FlagAdvice{
				HasPagination:       desc.HasPagination,
				CustomFlagOverrides: desc.CustomFlagOverrides,
				CustomFieldParsers:  desc.CustomFieldParsers,
			}.Sanitize()
			return ParseFieldsFromFlagsAndArgs[reqP](flagAdvice, fs, args)
		}
	}

	cmd := &cobra.Command{
		Use:   desc.Use,
		Short: desc.Short,
		Long:  desc.Long,
		Args:  cobra.ExactArgs(desc.numArgs),
		RunE:  queryLogic(desc, newQueryClientFn),
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
	moduleName string, newQueryClientFn func(grpc1.ClientConn) querier,
) *cobra.Command {
	desc := QueryDescriptor{
		Use:   use,
		Short: short,
		Long:  FormatLongDesc(long, NewLongMetadata(moduleName).WithShort(short)),
	}
	return BuildQueryCli[reqP](&desc, newQueryClientFn)
}

func GetParams[reqP proto.Message, querier any](moduleName string,
	newQueryClientFn func(grpc1.ClientConn) querier,
) *cobra.Command {
	return BuildQueryCli[reqP](&QueryDescriptor{
		Use:         "params [flags]",
		Short:       fmt.Sprintf("Get the params for the x/%s module", moduleName),
		QueryFnName: "Params",
	}, newQueryClientFn)
}

func callQueryClientFn(ctx context.Context, fnName string, req proto.Message, q any) (res proto.Message, err error) {
	qVal := reflect.ValueOf(q)
	method := qVal.MethodByName(fnName)
	if (method == reflect.Value{}) {
		return nil, fmt.Errorf("Method %s does not exist on the querier."+
			" You likely need to override QueryFnName in your Query descriptor", fnName)
	}
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

func queryLogic[querier any](desc *QueryDescriptor,
	newQueryClientFn func(grpc1.ClientConn) querier,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		queryClient := newQueryClientFn(clientCtx)

		req, err := desc.ParseQuery(args, cmd.Flags())
		if err != nil {
			return err
		}

		res, err := callQueryClientFn(cmd.Context(), desc.QueryFnName, req, queryClient)
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}
}
