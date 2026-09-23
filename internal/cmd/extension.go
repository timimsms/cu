package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
	"github.com/timimsms/cu/internal/config"
)

// extensionPrefix is the naming convention for external subcommands:
// `cu worklog` runs `cu-worklog` from PATH, the same shape kubectl and gh use.
const extensionPrefix = "cu-"

// extensionName constrains a word to a bare command name. Without this,
// exec.LookPath treats anything containing a separator as a *path* rather than
// a PATH search, so `cu ../../tmp/evil` would resolve and run an executable
// that was never on PATH at all. Extensions are found on PATH, by name, or not
// at all.
var extensionName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// tryExtension runs an external subcommand when cobra cannot resolve the word,
// and never returns if it does. It is deliberately the *last* resort: builtins
// always win, so an extension can never shadow a cu command.
//
// There is no PATH scan at startup — one exec.LookPath only once a word is
// known to be unresolvable — so this costs nothing on the normal path.
func tryExtension(args []string) {
	if len(args) == 0 {
		return
	}

	// Only consider a word cobra itself could not resolve.
	if _, _, err := rootCmd.Find(args); err == nil {
		return
	}

	word := firstCommandWord(args, rootCmd.PersistentFlags())
	if word == "" || strings.HasPrefix(word, extensionPrefix) || !extensionName.MatchString(word) {
		return
	}

	path, err := exec.LookPath(extensionPrefix + word)
	if err != nil {
		// No such extension: fall through so cobra reports the unknown
		// command with its did-you-mean suggestions.
		return
	}

	runExtension(path, word, args)
}

// runExtension execs the extension and exits with its status. It does not
// return.
func runExtension(path, word string, args []string) {
	// Pass every argument after the extension word through verbatim, including
	// flags that look like cu's own — they belong to the extension.
	rest := argsAfter(args, word)

	// #nosec G702 G204 -- the resolved path is not attacker-controlled: the word
	// is constrained to ^[a-zA-Z0-9][a-zA-Z0-9_-]*$ so it cannot contain a path
	// separator, meaning LookPath can only return an executable already on the
	// user's own PATH. The arguments are the user's own command line, forwarded
	// verbatim, and cross no privilege boundary — this is the same trust model
	// as the shell that invoked cu.
	cmd := exec.Command(path, rest...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = extensionEnv(args)

	err := cmd.Run()
	if err == nil {
		os.Exit(0)
	}

	// Propagate the child's exit status; a cu-shaped failure code from an
	// extension should reach the caller's shell unchanged.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}

	fmt.Fprintf(os.Stderr, "failed to run %s: %v\n", filepath.Base(path), err)
	os.Exit(1)
}

// extensionEnv hands the child the context it would otherwise have to
// rediscover: which config cu resolved, and which workspace and project file
// are in play.
func extensionEnv(args []string) []string {
	// The extension runs before PersistentPreRunE, so load config here. A
	// failure is not fatal — the extension simply gets less context.
	_ = config.Init(configFileFrom(args))

	env := append(os.Environ(),
		"CU_CONFIG_DIR="+filepath.Dir(config.GlobalConfigPath()),
		"CU_WORKSPACE="+config.GetString("default_workspace"),
		"CU_PROJECT_CONFIG="+config.GetProjectConfigPath(),
	)
	return env
}

// configFileFrom extracts a --config value so the child sees the same config
// cu would have used.
func configFileFrom(args []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if v, ok := strings.CutPrefix(a, "--config="); ok {
			return v
		}
		if a == "--config" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// argsAfter returns everything following the extension word, so cu's own
// global flags that preceded it are not forwarded.
func argsAfter(args []string, word string) []string {
	for i, a := range args {
		if a == word {
			return args[i+1:]
		}
	}
	return nil
}

// firstCommandWord returns the first positional argument — the candidate
// command — skipping global flags and the values they consume. Getting this
// wrong would treat a flag's value as a command name, so flag arity is read
// from the flag set rather than guessed.
func firstCommandWord(args []string, flags *pflag.FlagSet) string {
	for i := 0; i < len(args); i++ {
		a := args[i]

		switch {
		case a == "--":
			return ""

		case strings.HasPrefix(a, "--"):
			if strings.Contains(a, "=") {
				continue // --flag=value consumes nothing further
			}
			name := strings.TrimPrefix(a, "--")
			if takesValue(flags.Lookup(name)) {
				i++
			}

		case strings.HasPrefix(a, "-") && len(a) > 1:
			if strings.Contains(a, "=") {
				continue
			}
			// In a shorthand cluster only the final flag can take a value.
			last := a[len(a)-1:]
			if takesValue(flags.ShorthandLookup(last)) {
				i++
			}

		default:
			return a
		}
	}
	return ""
}

// takesValue reports whether a flag consumes the following argument. An
// unknown flag is assumed not to, so an unrecognised flag cannot swallow the
// command word.
func takesValue(f *pflag.Flag) bool {
	return f != nil && f.Value.Type() != "bool"
}
