package osmocli

import (
	"reflect"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type testingStruct struct {
	Int      int64
	UInt     uint64
	String   string
	Float    float64
	Duration time.Duration
	Pointer  *testingStruct
	Slice    sdk.Coins
	Struct   interface{}
	Dec      sdk.Dec
}

func TestParseFieldFromArg(t *testing.T) {
	tests := map[string]struct {
		testingStruct
		arg        string
		fieldIndex int

		expectedStruct testingStruct
		expectingErr   bool
	}{
		"Int value changes from -20 to 10": {
			testingStruct:  testingStruct{Int: -20},
			arg:            "10",
			fieldIndex:     0,
			expectedStruct: testingStruct{Int: 10},
		},
		"Attempt to change Int value 20 to string value": { // does not return error, simply does not change the struct
			testingStruct:  testingStruct{Int: 20},
			arg:            "hello",
			fieldIndex:     0,
			expectedStruct: testingStruct{Int: 20},
		},
		"UInt value changes from 20 to 10": {
			testingStruct:  testingStruct{UInt: 20},
			arg:            "10",
			fieldIndex:     1,
			expectedStruct: testingStruct{UInt: 10},
		},
		"String value change": {
			testingStruct:  testingStruct{String: "hello"},
			arg:            "world",
			fieldIndex:     2,
			expectedStruct: testingStruct{String: "world"},
		},
		"Changing unset value (simply sets the value)": {
			testingStruct:  testingStruct{Int: 20},
			arg:            "hello",
			fieldIndex:     2,
			expectedStruct: testingStruct{Int: 20, String: "hello"},
		},
		"Float value change": {
			testingStruct:  testingStruct{Float: 20.0},
			arg:            "30.0",
			fieldIndex:     3,
			expectedStruct: testingStruct{Float: 30.0},
		},
		"Duration value changes from .Hour to .Second": {
			testingStruct:  testingStruct{Duration: time.Hour},
			arg:            "1s",
			fieldIndex:     4,
			expectedStruct: testingStruct{Duration: time.Second},
		},
		"Attempt to change pointer": { // for reflect.Ptr kind ParseFieldFromArg does nothing, hence no changes take place
			testingStruct:  testingStruct{Pointer: &testingStruct{}},
			arg:            "*whatever",
			fieldIndex:     5,
			expectedStruct: testingStruct{Pointer: &testingStruct{}},
		},
		"Slice change": {
			testingStruct: testingStruct{Slice: sdk.Coins{
				sdk.NewCoin("foo", sdk.NewInt(100)),
				sdk.NewCoin("bar", sdk.NewInt(100)),
			}},
			arg:        "10foo,10bar", // Should be of a format suitable for ParseCoinsNormalized
			fieldIndex: 6,
			expectedStruct: testingStruct{Slice: sdk.Coins{ // swapped places due to lexicographic order
				sdk.NewCoin("bar", sdk.NewInt(10)),
				sdk.NewCoin("foo", sdk.NewInt(10)),
			}},
		},
		"Struct (sdk.Coin) change": {
			testingStruct:  testingStruct{Struct: sdk.NewCoin("bar", sdk.NewInt(10))}, // only supports sdk.Int, sdk.Coin or time.Time, other structs are not recognized
			arg:            "100bar",
			fieldIndex:     7,
			expectedStruct: testingStruct{Struct: sdk.NewCoin("bar", sdk.NewInt(10))},
		},
		"Unrecognizable struct": {
			testingStruct: testingStruct{Struct: testingStruct{}}, // only supports sdk.Int, sdk.Coin or time.Time, other structs are not recognized
			arg:           "whatever",
			fieldIndex:    7,
			expectingErr:  true,
		},
		"Multiple fields in struct are set": {
			testingStruct:  testingStruct{Int: 20, UInt: 10, String: "hello", Pointer: &testingStruct{}},
			arg:            "world",
			fieldIndex:     2,
			expectedStruct: testingStruct{Int: 20, UInt: 10, String: "world", Pointer: &testingStruct{}},
		},
		"All fields in struct set": {
			testingStruct: testingStruct{
				Int:      20,
				UInt:     10,
				String:   "hello",
				Float:    30.0,
				Duration: time.Second,
				Pointer:  &testingStruct{},
				Slice: sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(100)),
					sdk.NewCoin("bar", sdk.NewInt(100)),
				},
				Struct: sdk.NewCoin("bar", sdk.NewInt(10)),
			},
			arg:        "1foo,15bar",
			fieldIndex: 6,
			expectedStruct: testingStruct{
				Int:      20,
				UInt:     10,
				String:   "hello",
				Float:    30.0,
				Duration: time.Second,
				Pointer:  &testingStruct{},
				Slice: sdk.Coins{
					sdk.NewCoin("bar", sdk.NewInt(15)),
					sdk.NewCoin("foo", sdk.NewInt(1)),
				},
				Struct: sdk.NewCoin("bar", sdk.NewInt(10)),
			},
		},
		"Dec struct": {
			testingStruct:  testingStruct{Dec: sdk.MustNewDecFromStr("100")},
			arg:            "10",
			fieldIndex:     8,
			expectedStruct: testingStruct{Dec: sdk.MustNewDecFromStr("10")},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			val := reflect.ValueOf(&tc.testingStruct).Elem()
			typ := reflect.TypeOf(&tc.testingStruct).Elem()

			fVal := val.Field(tc.fieldIndex)
			fType := typ.Field(tc.fieldIndex)

			err := ParseFieldFromArg(fVal, fType, tc.arg)

			if !tc.expectingErr {
				require.Equal(t, tc.expectedStruct, tc.testingStruct)
			} else {
				require.Error(t, err)
			}
		})
	}
}
