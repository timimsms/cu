package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/auth/mock"
	cuerrors "github.com/tim/cu/internal/errors"
)

func TestMockAuthProvider_BasicOperations(t *testing.T) {
	provider := mock.NewAuthProvider()

	t.Run("not authenticated initially", func(t *testing.T) {
		assert.False(t, provider.IsAuthenticated("default"))
		_, err := provider.GetToken("default")
		assert.ErrorIs(t, err, cuerrors.ErrNotAuthenticated)
	})

	t.Run("save and retrieve token", func(t *testing.T) {
		token := &auth.Token{
			Value:     "test-token",
			Workspace: "default",
			Email:     "test@example.com",
		}

		err := provider.SaveToken("default", token)
		assert.NoError(t, err)
		assert.True(t, provider.IsAuthenticated("default"))

		retrieved, err := provider.GetToken("default")
		assert.NoError(t, err)
		assert.Equal(t, token.Value, retrieved.Value)
		assert.Equal(t, token.Email, retrieved.Email)
	})

	t.Run("delete token", func(t *testing.T) {
		err := provider.DeleteToken("default")
		assert.NoError(t, err)
		assert.False(t, provider.IsAuthenticated("default"))
	})

	t.Run("error simulation", func(t *testing.T) {
		provider.SetGetError(errors.New("simulated error"))
		_, err := provider.GetToken("default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated error")
	})
}

func TestMockScenarios(t *testing.T) {
	provider := mock.NewAuthProvider()
	scenarios := mock.NewScenarios(provider)

	t.Run("authenticated scenario", func(t *testing.T) {
		auth := scenarios.Authenticated()
		assert.True(t, auth.IsAuthenticated("default"))
	})

	t.Run("not authenticated scenario", func(t *testing.T) {
		auth := scenarios.NotAuthenticated()
		assert.False(t, auth.IsAuthenticated("default"))
	})

	t.Run("expired token scenario", func(t *testing.T) {
		auth := scenarios.ExpiredToken()
		_, err := auth.GetToken("default")
		assert.ErrorIs(t, err, cuerrors.ErrTokenExpired)
	})

	t.Run("multiple workspaces scenario", func(t *testing.T) {
		auth := scenarios.MultipleWorkspaces()
		assert.True(t, auth.IsAuthenticated("default"))
		assert.True(t, auth.IsAuthenticated("production"))
		assert.True(t, auth.IsAuthenticated("staging"))

		workspaces, err := auth.ListWorkspaces()
		assert.NoError(t, err)
		assert.Len(t, workspaces, 3)
	})
}