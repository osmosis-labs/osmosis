package cache_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/cache"
)

// This is a basic ChatGPT generates test case for the cache.
func TestCache(t *testing.T) {
	cache := cache.New()

	const defaultKeyToSet = "key1"

	// Test cases
	testCases := []struct {
		name       string
		key        string
		value      interface{}
		expiration time.Duration
		sleep      time.Duration
		expected   interface{}
	}{
		{"ValidKey", defaultKeyToSet, "value1", time.Second * 5, 0, "value1"},
		{"ExpiredKey", defaultKeyToSet, "value2", time.Nanosecond, time.Millisecond * 10, nil},
		{"NonExistentKey", "key2", "value3", time.Second * 5, 0, nil},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache.Set(defaultKeyToSet, tc.value, tc.expiration)

			// Sleep if necessary to simulate expiration
			time.Sleep(tc.sleep)

			val, exists := cache.Get(tc.key)

			// Check existence and value
			if exists != (tc.expected != nil) {
				t.Errorf("Expected existence: %v, Got: %v", tc.expected != nil, exists)
			}

			// Check value if it exists
			if exists && val != tc.expected {
				t.Errorf("Expected value: %v, Got: %v", tc.expected, val)
			}
		})
	}
}

// This test does basic validation against concurrency.
// That is, it tests that there are no deadlocks.
func TestConcurrentCache(t *testing.T) {
	cache := cache.New()

	seed := int64(10)
	rand := rand.New(rand.NewSource(seed))

	// Number of goroutines
	numGoroutines := 10
	numRunsPerRoutine := 15
	maxKeyNumRand := 10
	expirationMaxMs := 100

	// Wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Channel for goroutines to communicate errors
	errCh := make(chan error, numGoroutines*numRunsPerRoutine)

	// Mutex protecting the random number generator preventing
	// data races.
	randMx := sync.Mutex{}

	// Run goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			for i := 0; i < numRunsPerRoutine; i++ {

				randMx.Lock()
				randKey := rand.Intn(maxKeyNumRand)
				randMx.Unlock()

				// Random key and value
				key := fmt.Sprintf("key%d", randKey)
				value := "does not matter"

				// Random expiration time
				randMx.Lock()
				randDurarion := rand.Intn(expirationMaxMs)
				randMx.Unlock()

				expiration := time.Millisecond * time.Duration(randDurarion)

				// Set value in cache
				cache.Set(key, value, expiration)

				// Simulate some random work in the goroutine
				randMx.Lock()
				randDuarion := rand.Intn(expirationMaxMs)
				randMx.Unlock()

				time.Sleep(time.Millisecond * time.Duration(randDuarion))

				// Retrieve value from the cache
				val, exists := cache.Get(key)

				// Check if the retrieved value matches the expected value
				if exists && val != value {
					errCh <- fmt.Errorf("Goroutine %d: Expected value %s, Got %s", index, value, val)
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the error channel to signal the end of errors
	close(errCh)

	// Collect errors from goroutines
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	// Check for errors
	if len(errors) > 0 {
		t.Errorf("Concurrent Cache Test failed with %d errors:", len(errors))
		for _, err := range errors {
			t.Error(err)
		}
	}
}
