package auth

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timimsms/cu/internal/config"
	cuerrors "github.com/timimsms/cu/internal/errors"
	"github.com/zalando/go-keyring"
)

// isolate swaps the OS keyring for go-keyring's in-memory mock and points the
// config at an empty temp directory, so nothing touches the real keychain or
// ~/.config/cu.
func isolate(t *testing.T) {
	t.Helper()
	keyring.MockInit()
	t.Setenv("CU_DEFAULT_WORKSPACE", "")

	viper.Reset()
	old := config.DefaultConfigDir
	config.DefaultConfigDir = t.TempDir()
	t.Chdir(t.TempDir())
	t.Cleanup(func() { config.DefaultConfigDir = old; viper.Reset() })
	require.NoError(t, config.Init(""))
}

func TestCurrentWorkspace(t *testing.T) {
	t.Run("unset config falls back to default", func(t *testing.T) {
		isolate(t)
		assert.Equal(t, DefaultWorkspace, CurrentWorkspace())
	})

	t.Run("honours default_workspace", func(t *testing.T) {
		isolate(t)
		config.Set("default_workspace", "prod")
		assert.Equal(t, "prod", CurrentWorkspace())
	})

	t.Run("environment overrides config", func(t *testing.T) {
		isolate(t)
		config.Set("default_workspace", "prod")
		viper.Reset()
		t.Setenv("CU_DEFAULT_WORKSPACE", "staging")
		require.NoError(t, config.Init(""))
		assert.Equal(t, "staging", CurrentWorkspace())
	})
}

func TestGetCurrentToken(t *testing.T) {
	t.Run("named workspace login is used", func(t *testing.T) {
		isolate(t)
		m := NewManager()
		require.NoError(t, m.SaveToken("prod", &Token{Value: "prod-token", Workspace: "prod"}))
		config.Set("default_workspace", "prod")

		token, err := m.GetCurrentToken()
		require.NoError(t, err)
		assert.Equal(t, "prod-token", token.Value)
	})

	t.Run("named token is not read without default_workspace", func(t *testing.T) {
		isolate(t)
		m := NewManager()
		require.NoError(t, m.SaveToken("prod", &Token{Value: "prod-token"}))

		_, err := m.GetCurrentToken()
		assert.ErrorIs(t, err, cuerrors.ErrNotAuthenticated)
	})

	t.Run("unset config falls back to default account", func(t *testing.T) {
		isolate(t)
		m := NewManager()
		require.NoError(t, m.SaveToken("", &Token{Value: "default-token"}))
		require.NoError(t, m.SaveToken("prod", &Token{Value: "prod-token"}))

		token, err := m.GetCurrentToken()
		require.NoError(t, err)
		assert.Equal(t, "default-token", token.Value)
	})
}
