package cmd

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveFieldID(t *testing.T) {
	newCmd := func() *cobra.Command {
		c := &cobra.Command{}
		c.Flags().StringP("list", "l", "", "")
		return c
	}

	t.Run("a field id is used as given, even with a default list", func(t *testing.T) {
		prev := viper.Get("default_list")
		viper.Set("default_list", "some-other-list")
		t.Cleanup(func() { viper.Set("default_list", prev) })

		// A nil client proves no list lookup happens for an id.
		id, err := resolveFieldID(context.Background(), nil, newCmd(), "a5013cb5-9244-49be-a6a2-90e1cc8a1d31")
		require.NoError(t, err)
		assert.Equal(t, "a5013cb5-9244-49be-a6a2-90e1cc8a1d31", id)
	})

	t.Run("a name without any list is a clear error, not a raw API call", func(t *testing.T) {
		prev := viper.Get("default_list")
		viper.Set("default_list", "")
		t.Cleanup(func() { viper.Set("default_list", prev) })

		_, err := resolveFieldID(context.Background(), nil, newCmd(), "Repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--list")
	})

	t.Run("find does not print usage on API errors", func(t *testing.T) {
		assert.True(t, fieldFindCmd.SilenceUsage)
	})
}
