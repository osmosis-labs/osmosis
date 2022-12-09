package osmocli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagAdvice struct {
	HasPagination bool

	// Map of FieldName -> FlagName
	CustomFlagOverrides map[string]string
	CustomFieldParsers  map[string]CustomFieldParserFn

	// Tx sender value
	IsTx              bool
	TxSenderFieldName string
	FromValue         string
}

type FlagDesc struct {
	RequiredFlags []*pflag.FlagSet
	OptionalFlags []*pflag.FlagSet
}

type FieldReadLocation = bool

const (
	UsedArg  FieldReadLocation = true
	UsedFlag FieldReadLocation = false
)

// CustomFieldParser function.
type CustomFieldParserFn = func(arg string, flags *pflag.FlagSet) (valueToSet any, usedArg FieldReadLocation, err error)

func (f FlagAdvice) Sanitize() FlagAdvice {
	// map CustomFlagOverrides & CustomFieldParser keys to lower-case
	// initialize if uninitialized
	newFlagOverrides := make(map[string]string, len(f.CustomFlagOverrides))
	for k, v := range f.CustomFlagOverrides {
		newFlagOverrides[strings.ToLower(k)] = v
	}
	f.CustomFlagOverrides = newFlagOverrides
	newFlagParsers := make(map[string]CustomFieldParserFn, len(f.CustomFieldParsers))
	for k, v := range f.CustomFieldParsers {
		newFlagParsers[strings.ToLower(k)] = v
	}
	f.CustomFieldParsers = newFlagParsers
	return f
}

func FlagOnlyParser[v any](f func(fs *pflag.FlagSet) (v, error)) CustomFieldParserFn {
	return func(_arg string, fs *pflag.FlagSet) (any, FieldReadLocation, error) {
		t, err := f(fs)
		return t, UsedFlag, err
	}
}

// AddFlags from desc to cmd.
// Required flags are marked as required.
func AddFlags(cmd *cobra.Command, desc FlagDesc) {
	for i := 0; i < len(desc.OptionalFlags); i++ {
		cmd.Flags().AddFlagSet(desc.OptionalFlags[i])
	}
	for i := 0; i < len(desc.RequiredFlags); i++ {
		fs := desc.RequiredFlags[i]
		cmd.Flags().AddFlagSet(fs)

		// mark all these flags as required.
		fs.VisitAll(func(flag *pflag.Flag) {
			err := cmd.MarkFlagRequired(flag.Name)
			if err != nil {
				panic(err)
			}
		})
	}
}
