package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		want int
	}{
		{"bare word", []string{"worklog", "plan"}, 0},
		{"no args", nil, -1},
		{"only flags", []string{"--debug"}, -1},

		// A value-taking flag must not have its value mistaken for the
		// command — that would exec cu-x instead of cu-worklog.
		{"long flag with value", []string{"--config", "x", "worklog"}, 2},
		{"long flag with equals", []string{"--config=x", "worklog"}, 1},
		{"shorthand with value", []string{"-o", "json", "worklog"}, 2},
		{"shorthand with equals", []string{"-o=json", "worklog"}, 1},

		// Boolean flags consume nothing.
		{"bool flag", []string{"--debug", "worklog"}, 1},
		{"bool then valued", []string{"--debug", "--config", "x", "worklog"}, 3},

		// The word itself can appear earlier as a flag value; the index must
		// point at the command, not the first string that matches it.
		{"word repeated as flag value", []string{"--config", "worklog", "worklog", "plan"}, 2},

		// An unknown flag is assumed not to take a value, so it cannot
		// swallow the command word.
		{"unknown flag", []string{"--mystery", "worklog"}, 1},

		{"double dash ends the search", []string{"--", "worklog"}, -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, firstCommandWord(tc.args, testFlags()))
		})
	}
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

// TestExtensionHelperProcess is not a real test: runCu re-executes the test
// binary into it so the full Execute path — including os.Exit — runs in a
// child process the test can observe.
func TestExtensionHelperProcess(t *testing.T) {
	if os.Getenv("CU_EXT_HELPER") != "1" {
		t.Skip("helper process for exec-based extension tests")
	}
	args := []string{"cu"}
	for i, a := range os.Args {
		if a == "--" {
			args = append(args, os.Args[i+1:]...)
			break
		}
	}
	os.Args = args
	if err := Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// fakeExtension is a cu-hello that prints each argument it received, then the
// config dir cu handed it, then exits with $FAKE_EXIT (or kills itself with
// SIGTERM when asked).
const fakeExtension = `#!/bin/sh
for a in "$@"; do printf 'arg:%s\n' "$a"; done
printf 'dir:%s\n' "$CU_CONFIG_DIR"
if [ "$FAKE_SIGNAL" = "1" ]; then kill -TERM $$; fi
exit "${FAKE_EXIT:-0}"
`

// runCu runs `cu args...` in a child process with only binDir on PATH and an
// isolated HOME, returning its stdout and exit status.
func runCu(t *testing.T, binDir string, env []string, args ...string) (string, int) {
	t.Helper()
	home := t.TempDir()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestExtensionHelperProcess$", "--"}, args...)...)
	cmd.Env = append([]string{
		"CU_EXT_HELPER=1",
		"PATH=" + binDir,
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
	}, env...)
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(out), exitErr.ExitCode()
	}
	require.NoError(t, err)
	return string(out), 0
}

func writeScript(t *testing.T, dir, name, body string) {
	t.Helper()
	// #nosec G306 -- the test fixture must be executable to be found on PATH.
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o700))
}

func TestExtensionDispatchExec(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell script as the fake extension")
	}
	bin := t.TempDir()
	writeScript(t, bin, "cu-hello", fakeExtension)

	argsOf := func(out string) []string {
		var got []string
		for _, line := range strings.Split(out, "\n") {
			if v, ok := strings.CutPrefix(line, "arg:"); ok {
				got = append(got, v)
			}
		}
		return got
	}

	t.Run("forwards everything after the word verbatim", func(t *testing.T) {
		out, code := runCu(t, bin, nil, "hello", "plan", "--since", "1d", "-o", "json")
		assert.Equal(t, 0, code)
		assert.Equal(t, []string{"plan", "--since", "1d", "-o", "json"}, argsOf(out))
	})

	t.Run("cu's own preceding flags are not forwarded", func(t *testing.T) {
		out, code := runCu(t, bin, nil, "--debug", "-o", "json", "hello", "plan")
		assert.Equal(t, 0, code)
		assert.Equal(t, []string{"plan"}, argsOf(out))
	})

	t.Run("word repeated as a flag value", func(t *testing.T) {
		out, code := runCu(t, bin, nil, "--config", "hello", "hello", "real")
		assert.Equal(t, 0, code)
		assert.Equal(t, []string{"real"}, argsOf(out))
	})

	t.Run("a --config after the word belongs to the extension", func(t *testing.T) {
		out, code := runCu(t, bin, nil, "hello", "--config", "/elsewhere/c.yaml")
		assert.Equal(t, 0, code)
		assert.Equal(t, []string{"--config", "/elsewhere/c.yaml"}, argsOf(out))
		assert.NotContains(t, out, "dir:/elsewhere", "cu must not adopt the extension's --config")
	})

	t.Run("exit code passes through", func(t *testing.T) {
		_, code := runCu(t, bin, []string{"FAKE_EXIT=7"}, "hello")
		assert.Equal(t, 7, code)
	})

	t.Run("signal exit maps to 128+signal", func(t *testing.T) {
		_, code := runCu(t, bin, []string{"FAKE_SIGNAL=1"}, "hello")
		assert.Equal(t, 128+15, code)
	})

	t.Run("cu-help on PATH does not replace cu help", func(t *testing.T) {
		hijack := t.TempDir()
		writeScript(t, hijack, "cu-help", "#!/bin/sh\necho HIJACKED\nexit 3\n")
		out, code := runCu(t, hijack, nil, "help")
		assert.Equal(t, 0, code)
		assert.NotContains(t, out, "HIJACKED")
		assert.Contains(t, out, "Usage:")
	})
}
