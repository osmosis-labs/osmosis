package cmd

import (
	"bytes"
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
