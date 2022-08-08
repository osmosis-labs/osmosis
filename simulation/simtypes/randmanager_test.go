package simtypes

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

var defaultSeed = int64(10)

func getDefaultRandInstance() *rand.Rand {
	return rand.New(rand.NewSource(defaultSeed))
}

func getKDefaultRandManager(k int) []randManager {
	rms := make([]randManager, k)
	for i := 0; i < k; i++ {
		rms[i] = newRandManager(getDefaultRandInstance())
	}
	return rms
}

func randInstancesEqual(rands []*rand.Rand) bool {
	value := rands[0].Int()
	for i := 1; i < len(rands); i++ {
		if rands[i].Int() != value {
			return false
		}
	}
	return true
}

// Test that the rand manager GetRand() are all independent of one another.
func TestRandManagerGetRandIndependence(t *testing.T) {
	rms := getKDefaultRandManager(3)
	expectedEqualRands := []*rand.Rand{}
	// We want to test that in each of the the following three scenarios, r2 is equal:
	// 1) r1 := rm.GetRand(); r2 := rm.GetRand();
	// 2) r1 := rm.GetRand(); r2 := rm.GetRand(); _ = r1.Int()
	// 3) r1 := rm.GetRand(); r1.Int(); r2 := rm.GetRand();
	scenario1RM := rms[0]
	scenario1RM.GetRand()
	r2 := scenario1RM.GetRand()
	expectedEqualRands = append(expectedEqualRands, r2)

	scenario2RM := rms[1]
	r1 := scenario2RM.GetRand()
	r2 = scenario2RM.GetRand()
	r1.Int()
	expectedEqualRands = append(expectedEqualRands, r2)

	scenario3RM := rms[2]
	r1 = scenario3RM.GetRand()
	r1.Int()
	r2 = scenario3RM.GetRand()
	expectedEqualRands = append(expectedEqualRands, r2)
	require.True(t, randInstancesEqual(expectedEqualRands))
}

// Test that the rand manager GetSeededRand() for the same seed are all returning the same rand instance.
func TestRandManagerSameSeedGetSeededRand(t *testing.T) {
	rms := getKDefaultRandManager(3)
	seed := "test seed"
	// We want to test that in each of the the following three scenarios, we generated the same 'trace' of values.
	// 1) r1 := rm.GetSeededRand(seed); r1.Int(); r2 := rm.GetSeededRand(seed); r2.Int();
	// 2) r1 := rm.GetSeededRand(seed); r2 := rm.GetSeededRand(seed); r1.Int(); r2.Int();
	// 3) r1 := rm.GetSeededRand(seed); r2 := rm.GetSeededRand(seed); r2.Int(); r1.Int();
	scenario1RM := rms[0]
	r1 := scenario1RM.GetSeededRand(seed)
	v1 := r1.Int()
	r2 := scenario1RM.GetSeededRand(seed)
	scenario1Trace := []int{v1, r2.Int()}

	scenario2RM := rms[1]
	r1 = scenario2RM.GetSeededRand(seed)
	r2 = scenario2RM.GetSeededRand(seed)
	scenario2Trace := []int{r1.Int(), r2.Int()}

	scenario3RM := rms[2]
	r1 = scenario3RM.GetSeededRand(seed)
	r2 = scenario3RM.GetSeededRand(seed)
	scenario3Trace := []int{r2.Int(), r1.Int()}

	require.Equal(t, scenario1Trace, scenario2Trace)
	require.Equal(t, scenario1Trace, scenario3Trace)
}
