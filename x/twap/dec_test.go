package twap_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/x/twap/client/queryproto"
)

func TestSDKDecMarshal(t *testing.T) {

	// 1. Try to marshal and then unmarshal a zero Dec -> works fine
	zeroDec := osmomath.ZeroDec()

	bz, err := zeroDec.Marshal()
	require.NoError(t, err)

	var dec osmomath.Dec
	err = dec.Unmarshal(bz)
	require.NoError(t, err)

	require.Equal(t, zeroDec, dec)

	// 2. Try to marshal and unmarshal sdk dec as a field -> works fine.
	twapResponse := queryproto.ArithmeticTwapResponse{
		ArithmeticTwap: zeroDec,
	}

	bz, err = twapResponse.Marshal()
	require.NoError(t, err)

	var twapResponse2 queryproto.ArithmeticTwapResponse
	err = twapResponse2.Unmarshal(bz)

	require.NoError(t, err)

	// 3. Now, try incorrectly initializing sdk.Dec and then marshal and unmarshal
	// -> works fine but initializes to zero.

	twapResponse = queryproto.ArithmeticTwapResponse{
		ArithmeticTwap: osmomath.Dec{}, // incorrectly initialized
	}

	bz, err = twapResponse.Marshal()
	require.NoError(t, err)

	err = twapResponse2.Unmarshal(bz)
	require.NoError(t, err)

	// Response are not equal but panic does not occur.
	require.NotEqual(t, twapResponse, twapResponse2)

	// The unmarshaled one gets intialized to zero.
	require.Equal(t, osmomath.Dec{}, twapResponse.ArithmeticTwap)
	require.Equal(t, osmomath.ZeroDec(), twapResponse2.ArithmeticTwap)

	// 4. Now try not ininitializing sdk.Dec and then marshal and unmarshal
	// -> works fine but initializes unitialized field to zero.

	twapResponse = queryproto.ArithmeticTwapResponse{
		// Not initialized
	}

	bz, err = twapResponse.Marshal()
	require.NoError(t, err)

	err = twapResponse2.Unmarshal(bz)
	require.NoError(t, err)

	// Response are not equal but panic does not occur.
	require.NotEqual(t, twapResponse, twapResponse2)
	require.Equal(t, osmomath.Dec{}, twapResponse.ArithmeticTwap)
	require.Equal(t, osmomath.ZeroDec(), twapResponse2.ArithmeticTwap)

	var emptyBytes []byte
	uninitDec := osmomath.Dec{}
	err = uninitDec.Unmarshal(emptyBytes)
	require.NoError(t, err)

	strExample := "212881620000000000"
	bytes := []byte(strExample)
	fmt.Println(string(bytes))
	fmt.Println(len(bytes))
	uninitDec = osmomath.Dec{}
	err = uninitDec.Unmarshal(bytes)
	require.NoError(t, err)

	fmt.Println("after unmarshal ", uninitDec.String())

	err = uninitDec.Unmarshal(bytes)
	require.NoError(t, err)
}
