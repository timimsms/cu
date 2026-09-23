package cmd

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func testFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("cu", pflag.ContinueOnError)
	fs.String("config", "", "")
	fs.Bool("debug", false, "")
	fs.StringP("output", "o", "table", "")
	return fs
}

func TestFirstCommandWord(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"bare word", []string{"worklog", "plan"}, "worklog"},
		{"no args", nil, ""},
		{"only flags", []string{"--debug"}, ""},

		// A value-taking flag must not have its value mistaken for the
		// command — that would exec cu-x instead of cu-worklog.
		{"long flag with value", []string{"--config", "x", "worklog"}, "worklog"},
		{"long flag with equals", []string{"--config=x", "worklog"}, "worklog"},
		{"shorthand with value", []string{"-o", "json", "worklog"}, "worklog"},
		{"shorthand with equals", []string{"-o=json", "worklog"}, "worklog"},

		// Boolean flags consume nothing.
		{"bool flag", []string{"--debug", "worklog"}, "worklog"},
		{"bool then valued", []string{"--debug", "--config", "x", "worklog"}, "worklog"},

		// An unknown flag is assumed not to take a value, so it cannot
		// swallow the command word.
		{"unknown flag", []string{"--mystery", "worklog"}, "worklog"},

		{"double dash ends the search", []string{"--", "worklog"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, firstCommandWord(tc.args, testFlags()))
		})
	}
}

func TestArgsAfter(t *testing.T) {
	t.Run("forwards everything after the word", func(t *testing.T) {
		got := argsAfter([]string{"worklog", "plan", "--since", "1d"}, "worklog")
		assert.Equal(t, []string{"plan", "--since", "1d"}, got)
	})

	t.Run("cu's own preceding flags are not forwarded", func(t *testing.T) {
		got := argsAfter([]string{"--output", "json", "worklog", "plan"}, "worklog")
		assert.Equal(t, []string{"plan"}, got)
	})

	t.Run("word with no trailing args", func(t *testing.T) {
		assert.Empty(t, argsAfter([]string{"worklog"}, "worklog"))
	})

	t.Run("absent word", func(t *testing.T) {
		assert.Nil(t, argsAfter([]string{"task", "list"}, "worklog"))
	})
}

func TestConfigFileFrom(t *testing.T) {
	assert.Equal(t, "/tmp/c.yaml", configFileFrom([]string{"--config", "/tmp/c.yaml", "worklog"}))
	assert.Equal(t, "/tmp/c.yaml", configFileFrom([]string{"--config=/tmp/c.yaml", "worklog"}))
	assert.Equal(t, "", configFileFrom([]string{"worklog", "plan"}))
	assert.Equal(t, "", configFileFrom([]string{"--config"}), "a dangling --config must not panic")
}

func TestTryExtensionLeavesBuiltinsAlone(t *testing.T) {
	// tryExtension exits the process when it dispatches, so the safe
	// assertion is that a resolvable builtin returns normally — proving
	// builtins are never routed to an extension.
	tryExtension([]string{"version"})
	tryExtension([]string{"task", "list"})
	tryExtension([]string{})
}

func TestExtensionNameRejectsPaths(t *testing.T) {
	// exec.LookPath treats a word containing a separator as a path rather than
	// a PATH search, so without this constraint `cu ../../tmp/evil` would run
	// an executable that was never on PATH. Extensions are found on PATH, by
	// name, or not at all.
	for _, bad := range []string{
		"../evil", "../../tmp/evil", "/abs/evil", "dir/evil",
		".hidden", "-leading-dash", "", "has space", "semi;colon", "dot.dot",
	} {
		assert.False(t, extensionName.MatchString(bad), "must reject %q", bad)
	}

	for _, good := range []string{"worklog", "trailer", "my-ext", "my_ext", "ext2"} {
		assert.True(t, extensionName.MatchString(good), "must accept %q", good)
	}
}
