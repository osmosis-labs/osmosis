package cmd

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestInitAppConfigTemplate(t *testing.T) {
	// This test validates templates uses existing fields in config
	appTemplate, appConfig := initAppConfig()
	tmpl := template.New("appDefaultValues")
	tmpl, err := tmpl.Parse(appTemplate)
	require.NoError(t, err)
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, appConfig)
	require.NoError(t, err)
}

// The denoms below have to be listed in testDenomMap() for the rewriting to be
// observable: a denom that is missing from the asset list is written out
// unchanged whether or not the stream happens to be split.
const (
	testIBCDenom     = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	testFactoryDenom = "factory/osmo1z0qrclc8a9lz3t6vz2q0jrhcvfr0pkz0rq8u9z9/uusdc"
	testUnknownDenom = "ibc/0000000000000000000000000000000000000000000000000000000000000000"
)

func testDenomMap() map[string]string {
	return map[string]string{
		testIBCDenom:     "uatom",
		testFactoryDenom: "USDC.f",
	}
}

// writeChunked feeds input to a fresh customWriter in chunks of chunkSize bytes
// and returns everything the writer produced, the flushed tail included.
func writeChunked(t *testing.T, input string, chunkSize int) string {
	t.Helper()
	require.Greater(t, chunkSize, 0)

	var out bytes.Buffer
	w := &customWriter{originalOut: &out, baseMap: testDenomMap()}
	for remaining := input; remaining != ""; {
		n := min(chunkSize, len(remaining))
		written, err := w.Write([]byte(remaining[:n]))
		require.NoError(t, err)
		require.Equal(t, n, written, "Write must report that it accepted every byte of the input")
		remaining = remaining[n:]
	}
	require.NoError(t, w.Flush())
	return out.String()
}

// TestCustomWriterIsChunkingIndependent is a regression test for the denom
// rewriting being done per Write call instead of over the whole stream.
//
// io.Writer makes no guarantee that a token is handed over in a single call, so
// a denom can be split anywhere - including inside its "ibc/" / "factory/"
// marker. Because of that, the output has to depend only on the bytes received
// and not on how they were chunked: every chunking must produce exactly what
// writing the stream at once produces.
func TestCustomWriterIsChunkingIndependent(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "ibc denom",
			input: fmt.Sprintf(`{"denom": "%s", "amount": "100"}`, testIBCDenom),
		},
		{
			name:  "factory denom",
			input: fmt.Sprintf(`{"denom": "%s", "amount": "100"}`, testFactoryDenom),
		},
		{
			name:  "both denoms plus an unknown one",
			input: fmt.Sprintf(`{"a": "%s", "b": "%s", "c": "%s"}`, testIBCDenom, testFactoryDenom, testUnknownDenom),
		},
		{
			name:  "denoms at the very end of the stream",
			input: fmt.Sprintf(`amount: %s and %s`, testFactoryDenom, testIBCDenom),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := writeChunked(t, test.input, len(test.input))
			for chunkSize := 1; chunkSize <= len(test.input); chunkSize++ {
				require.Equal(t, expected, writeChunked(t, test.input, chunkSize),
					"output differs when the stream is split into chunks of %d bytes", chunkSize)
			}
		})
	}
}

// TestCustomWriterRewritesDenoms checks that the denoms of the asset list are
// still replaced by their human readable name, and that anything else is left
// untouched.
func TestCustomWriterRewritesDenoms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ibc denom is replaced",
			input:    fmt.Sprintf(`{"denom": "%s"}`, testIBCDenom),
			expected: `{"denom": "uatom"}`,
		},
		{
			name:     "factory denom is replaced",
			input:    fmt.Sprintf(`{"denom": "%s"}`, testFactoryDenom),
			expected: `{"denom": "USDC.f"}`,
		},
		{
			name:     "denom missing from the asset list is left as is",
			input:    fmt.Sprintf(`{"denom": "%s"}`, testUnknownDenom),
			expected: fmt.Sprintf(`{"denom": "%s"}`, testUnknownDenom),
		},
		{
			name:     "output without a denom is left as is",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, writeChunked(t, test.input, len(test.input)))
		})
	}
}

// TestCustomWriterBuffersPartialDenoms checks that an incomplete denom is held
// back until the rest of it arrives, while Write still reports that every byte
// of the input was accepted.
func TestCustomWriterBuffersPartialDenoms(t *testing.T) {
	var out bytes.Buffer
	w := &customWriter{originalOut: &out, baseMap: testDenomMap()}

	n, err := w.Write([]byte(testIBCDenom[:10]))
	require.NoError(t, err)
	require.Equal(t, 10, n)
	require.Empty(t, out.String(), "a partial denom must not be written out yet")

	_, err = w.Write([]byte(testIBCDenom[10:]))
	require.NoError(t, err)
	require.NoError(t, w.Flush())
	require.Equal(t, "uatom", out.String())
}
