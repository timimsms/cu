package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timimsms/cu/internal/api"
	"github.com/timimsms/cu/internal/auth"
	"github.com/timimsms/cu/internal/config"
	"github.com/zalando/go-keyring"
)

// isolateAuth swaps the OS keyring for go-keyring's in-memory mock and points
// the config at an empty temp directory, so the login/status/logout commands
// run end to end without touching the real keychain or ~/.config/cu.
func isolateAuth(t *testing.T) string {
	t.Helper()
	keyring.MockInit()
	t.Setenv("CU_DEFAULT_WORKSPACE", "")

	viper.Reset()
	cfgDir := t.TempDir()
	old := config.DefaultConfigDir
	config.DefaultConfigDir = cfgDir
	t.Chdir(t.TempDir())
	t.Cleanup(func() { config.DefaultConfigDir = old; viper.Reset() })
	require.NoError(t, config.Init(""))
	return cfgDir
}

func setLoginFlag(t *testing.T, name, value string) {
	t.Helper()
	require.NoError(t, authLoginCmd.Flags().Set(name, value))
	t.Cleanup(func() { _ = authLoginCmd.Flags().Set(name, "") })
}

func TestAuthLoginNamedWorkspaceIsUsedByClient(t *testing.T) {
	cfgDir := isolateAuth(t)
	setLoginFlag(t, "workspace", "prod")
	setLoginFlag(t, "token", "prod-token")

	captureStdout(t, func() { authLoginCmd.Run(authLoginCmd, nil) })

	assert.Equal(t, "prod", auth.CurrentWorkspace())
	written, err := os.ReadFile(filepath.Join(cfgDir, config.ConfigFileName+"."+config.ConfigType))
	require.NoError(t, err)
	assert.Contains(t, string(written), "default_workspace: prod", "the named workspace must persist for later invocations")

	token, err := auth.NewManager().GetCurrentToken()
	require.NoError(t, err)
	assert.Equal(t, "prod-token", token.Value)

	_, err = api.NewClient()
	assert.NoError(t, err, "the API client must authenticate with the token just saved")

	prev := outputFormat
	outputFormat = "json"
	t.Cleanup(func() { outputFormat = prev })
	out := captureStdout(t, func() { authStatusCmd.Run(authStatusCmd, nil) })
	assert.Contains(t, out, `"workspace": "prod"`)
	assert.Contains(t, out, `"authenticated": true`)

	captureStdout(t, func() { authLogoutCmd.Run(authLogoutCmd, nil) })
	_, err = api.NewClient()
	assert.Error(t, err, "logout without -w must remove the token the client was using")
}

func TestAuthLoginDefaultWorkspaceWritesNoConfig(t *testing.T) {
	cfgDir := isolateAuth(t)
	setLoginFlag(t, "token", "default-token")

	captureStdout(t, func() { authLoginCmd.Run(authLoginCmd, nil) })

	assert.Equal(t, auth.DefaultWorkspace, auth.CurrentWorkspace())
	_, err := os.Stat(filepath.Join(cfgDir, config.ConfigFileName+"."+config.ConfigType))
	assert.True(t, os.IsNotExist(err), "a plain login should not create a config file")

	_, err = api.NewClient()
	assert.NoError(t, err)
}

func TestAuthLoginDefaultAfterNamedSwitchesBack(t *testing.T) {
	isolateAuth(t)
	config.Set("default_workspace", "prod")
	setLoginFlag(t, "token", "default-token")

	captureStdout(t, func() { authLoginCmd.Run(authLoginCmd, nil) })

	assert.Equal(t, auth.DefaultWorkspace, auth.CurrentWorkspace())
	token, err := auth.NewManager().GetCurrentToken()
	require.NoError(t, err)
	assert.Equal(t, "default-token", token.Value)
}
