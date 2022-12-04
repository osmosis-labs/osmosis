package osmocli

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
)

// Parses arguments 1-1 from args
// makes an exception, where it allows Pagination to come from flags.
func ParseFieldsFromFlagsAndArgs[reqP proto.Message](flags *pflag.FlagSet, args []string) (reqP, error) {
	req := osmoutils.MakeNew[reqP]()
	v := reflect.ValueOf(req).Elem()
	t := v.Type()

	argIndexOffset := 0
	// Iterate over the fields in the struct
	for i := 0; i < t.NumField(); i++ {
		arg := ""
		if len(args) > i+argIndexOffset {
			arg = args[i+argIndexOffset]
		}
		usedArg, err := ParseField(flags, v, t, i, arg)
		if err != nil {
			return req, err
		}
		if !usedArg {
			argIndexOffset -= 1
		}
	}
	return req, nil
}

func ParseNumFields[reqP proto.Message]() int {
	req := osmoutils.MakeNew[reqP]()
	v := reflect.ValueOf(req).Elem()
	t := v.Type()
	return t.NumField()
}

func ParseExpectedFnName[reqP proto.Message]() string {
	req := osmoutils.MakeNew[reqP]()
	v := reflect.ValueOf(req).Elem()
	s := v.Type().String()
	// handle some non-std queries
	var prefixTrimmed string
	if strings.Contains(s, "Query") {
		prefixTrimmed = strings.Split(s, "Query")[1]
	} else {
		prefixTrimmed = strings.Split(s, ".")[1]
	}
	suffixTrimmed := strings.TrimSuffix(prefixTrimmed, "Request")
	return suffixTrimmed
}

func ParseHasPagination[reqP proto.Message]() bool {
	req := osmoutils.MakeNew[reqP]()
	t := reflect.ValueOf(req).Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		fType := t.Field(i)
		if fType.Type.String() == "*query.PageRequest" {
			return true
		}
	}
	return false
}

// returns (increment_arg_index, error)
func ParseField(flags *pflag.FlagSet, v reflect.Value, t reflect.Type, fieldIndex int, arg string) (bool, error) {
	fVal := v.Field(fieldIndex)
	fType := t.Field(fieldIndex)

	// fmt.Printf("Field %d: %s %s %s\n", fieldIndex, fType.Name, fType.Type, fType.Type.Kind())
	switch fType.Type.Kind() {
	// SetUint allows anyof type u8, u16, u32, u64, and uint
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uint:
		u, err := ParseUint(arg, fType.Name)
		if err != nil {
			return true, err
		}
		fVal.SetUint(u)
		return true, nil
	// SetInt allows anyof type i8,i16,i32,i64 and int
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		i, err := ParseInt(arg, fType.Name)
		if err != nil {
			return true, err
		}
		fVal.SetInt(i)
		return true, nil
	case reflect.String:
		s, err := ParseDenom(arg, fType.Name)
		if err != nil {
			return true, err
		}
		fVal.SetString(s)
		return true, nil
	case reflect.Ptr:
		// general strategy very TBD, rn just handle pagination
		typeStr := fType.Type.String()
		if typeStr == "*query.PageRequest" {
			pageReq, err := client.ReadPageRequest(flags)
			if err != nil {
				return false, err
			}
			fVal.Set(reflect.ValueOf(pageReq))
			return false, nil
		}
	case reflect.Struct:
		// Handle struct type
		// ...
	}
	fmt.Println(fType.Type.Kind().String())
	return true, fmt.Errorf("field type not recognized. Got type %v", fType)
}

func ParseUint(arg string, fieldName string) (uint64, error) {
	v, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s as uint for field %s: %w", arg, fieldName, err)
	}
	return v, nil
}

func ParseInt(arg string, fieldName string) (int64, error) {
	v, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s as int for field %s: %w", arg, fieldName, err)
	}
	return v, nil
}

func ParseUnixTime(arg string, fieldName string) (time.Time, error) {
	timeUnix, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse %s as unix time for field %s: %w", arg, fieldName, err)
	}
	startTime := time.Unix(timeUnix, 0)
	return startTime, nil
}

func ParseDenom(arg string, fieldName string) (string, error) {
	return strings.TrimSpace(arg), nil
}
