package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestDownloadAssetList(t *testing.T) {
	testCases := []struct {
		name         string
		statusCode   int
		body         string
		wantError    bool
		wantFileBody string
	}{
		{
			name:         "HTTP error leaves existing asset list unchanged",
			statusCode:   http.StatusInternalServerError,
			body:         `{"error":"upstream failure"}`,
			wantError:    true,
			wantFileBody: `{"assets":[{"base":"old"}]}`,
		},
		{
			name:         "valid JSON replaces existing asset list",
			statusCode:   http.StatusOK,
			body:         `{"assets":[{"base":"new"}]}`,
			wantFileBody: `{"assets":[{"base":"new"}]}`,
		},
		{
			name:         "invalid JSON leaves existing asset list unchanged",
			statusCode:   http.StatusOK,
			body:         `not json`,
			wantError:    true,
			wantFileBody: `{"assets":[{"base":"old"}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			fileName := filepath.Join(t.TempDir(), "assetlist.json")
			const originalBody = `{"assets":[{"base":"old"}]}`
			require.NoError(t, os.WriteFile(fileName, []byte(originalBody), 0644))

			err := downloadAssetList(server.URL, fileName)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			got, readErr := os.ReadFile(fileName)
			require.NoError(t, readErr)
			require.Equal(t, tc.wantFileBody, string(got))
		})
	}
}
