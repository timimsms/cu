# cu Context Layer — Design Specification

- **Status:** Accepted
- **Date:** 2026-07-16
- **Scope:** refs (named resources & saved queries), aliases (command macros), packs (shareable declarative bundles), and interfaces (`cu onboard`, pickers, `cu browse`).
- **Supersedes:** PROJECT_SPEC §4.3's `aliases.yml` location (amended to an `aliases:` key in config); confirms §4.3's `cu-<name>` PATH extension model.
- **Decision log:** §9 — final maintainer decisions, security-review foldins, fact-check corrections, and every contradiction with the three input designs.
- **Publishing note:** this file lives in `docs/design/`, deliberately outside `docs/site/` (the mkdocs `docs_dir`), so it does not publish to the documentation site.

---

## 1. Concepts

Four layers, one precedence chain, one config store:

| Layer | What it is | Lives in |
|---|---|---|
| **Refs** | Named resources (`support-bugs` → two list IDs) and saved queries (`my-sprint` → a compiled Filtered-Team-Tasks call). Consumed anywhere a resource is accepted, via bare name or `@name`. | `refs:` key in config layers |
| **Aliases** | gh-identical command macros. `$1..$9` substitution, `!` shell prefix (source-gated, §3.4/§6). `cu bugs` is an alias. | `aliases:` key in config layers |
| **Packs** | Installable, pinned, **declarative-inert** git repos bundling refs + aliases + defaults. A team's shared vocabulary. Packs can never carry executable content (§6). | cloned to XDG data dir, loaded as low-precedence config layers |
| **Interfaces** | `cu onboard` wizard, `cu pick`, disambiguation prompts, and (later) `cu browse`. All veneers over the same resolver — nothing interactive is un-scriptable. | `internal/ui` |

Two audiences, one contract: everything above is inspectable as data (`--json` everywhere, `cu ref resolve`, `cu which`, `cu config list --show-origin`), so terminal users and AI agents consume the identical surface.

**Name grammar** (refs and aliases): `[a-z0-9][a-z0-9-]*`, max 64 chars, lowercase only (this also makes viper's key case-folding harmless). Alias names are additionally validated against the reserved-word list (§5). `me` is a reserved ref name (`@me` is built in and unshadowable).

---

## 2. Config: files, layers, precedence

### 2.1 File layout (no new dot-dirs)

```
~/.config/cu/                  # existing XDG config dir. --config honored for reads today;
                               # $CU_CONFIG_DIR does NOT exist yet — to be added in the config rework.
  config.yaml                  # the only global file a user edits (the real filename — see note below)
  packs.lock.yml               # machine-written, 0600: sources, pins, revs, manifest hashes, trusted projects
  cache/                       # existing cache subsystem (unchanged)
<project>/.cu.yml              # discovered upward (existing behavior)
~/.local/share/cu/packs/       # $XDG_DATA_HOME/cu/packs — pack clones (reproducible artifacts, not config)
```

**Filename note:** the code writes `config.yaml` (`internal/config/config.go:28-30,81`: `ConfigFileName="config"`, `ConfigType="yaml"`), while the `--config` flag help text in `internal/cmd/root.go:46` says `config.yml` — that help text is wrong today and is a drive-by fix candidate. This spec uses `config.yaml` throughout.

### 2.2 Global `~/.config/cu/config.yaml`

```yaml
# existing keys, unchanged
output: table
default_workspace: "9018012345"
default_space: "Engineering"
default_list: "Sprint Backlog"

refs:
  # Compact form (string) = resource ref: comma-separated type:id.
  # Types: list | folder | space | user. Managed by `cu ref set` or `cu config set refs.<name>`.
  support-bugs: "list:901300123,list:901300456"
  eng: "space:790123"
  qa-team: "user:8823001,user:8823002"

  # Map form = query ref: a saved filter that compiles to GET /team/{id}/task (#25).
  my-sprint:
    from: ["@support-bugs", "list:901300789"]     # refs and raw typed ids mix
    where:
      assignees: ["@me"]
      statuses: ["in progress", "review"]
      tags: ["backend"]
      due: "before:+7d"
      include_closed: false
    sort: due
    order: asc

aliases:                       # gh-identical semantics
  bugs: "task list @support-bugs --status open"
  mine: "task list --assignee @me --status open"
  grab: "task update $1 --status 'in progress' --assignee @me"
  standup: "!cu task list @my-sprint --json id,name --jq '.[].name' | pbcopy"
  # `!` shell aliases are implicitly trusted HERE (you wrote your own global config) — see §3.4/§6.
```

### 2.3 Project `.cu.yml` (committed — the zero-install team pack)

```yaml
default_workspace: "9018012345"
default_space: "Engineering"
default_list: "Sprint Backlog"
refs:
  this-repo: "list:901300001"
aliases:
  ship: "task update $1 --status done"
  # `!` shell aliases ARE legal here, gated by TOFU + content hash at invocation (§6.5).
  # This file is the PR-reviewed home for team shell macros — packs cannot carry them (§6.2).
packs:                          # requirements; installed only via explicit `cu pack sync`
  - acme/payments@v1.2.0
  - acme/eng-common
```

### 2.4 Precedence

**Values** (output, default_*, etc.), one chain for every key:

```
flags > env (CU_WORKSPACE/CU_SPACE/CU_LIST/…) > project .cu.yml > global config.yaml
      > pack defaults (project packs: array order, then global install order) > built-ins
```

Env note: today's `AutomaticEnv` with prefix `CU` yields `CU_OUTPUT`/`CU_DEBUG`/`CU_DEFAULT_*`; `CU_WORKSPACE`/`CU_SPACE`/`CU_LIST` are **new bindings the config rework must add**.

**Name tables** (refs, aliases) merge **per-key**: `project > global > packs (same order)`. A project ref shadows a same-named global/pack ref; it never wipes the map. `cu config list --show-origin` (git-style) prints the winning source per key, including `pack:payments`. Env values pass through the same resolver, so `CU_LIST=@support-bugs` works.

**Credential-key blocklist (security foldin):** `api_token` — and any future credential-bearing keys — are **ignored with a loud warning** when they appear in pack `defaults:` or in a project `.cu.yml` layer. Credentials come only from the OS keyring, environment, or the user's own global config. Rationale: `api_token` is a first-class viper key today (`internal/config/config.go:20`), and without this strip a malicious project file or pack could become a token-substitution confused deputy. The per-layer merge (§2.5) enforces the strip structurally.

### 2.5 Prerequisite rework (load-bearing, blocks everything)

Current plumbing cannot express this chain and must be fixed first:

- `internal/config/config.go:58-64` merges `.cu.yml` via `viper.Set()` — viper's OVERRIDE slot — so project config silently outranks env and bound flags today (inverted precedence).
- `config.Save()` (`config.go:80-83`) is `viper.WriteConfigAs` of the *entire merged state*: running `cu config set` inside a project bakes project values and defaults into the global file. This is a shippable-today bug.
- `internal/cmd/root.go:91` puts `.` on the config search path — a config-injection hole that undermines the entire trust model. Remove (already planned).

Replace with: per-layer stores (packs → global → project) merged in code in documented order (with the credential-key strip from §2.4 applied to project and pack layers), and a **yaml.v3 node-based writer** (yaml.v3 is already a direct dep) that surgically updates only the touched key path, preserving comments and ordering — the gh yamlmap approach. `cu config unset <key>`, `--project` targeting, and `--show-origin` ride this rework. `$CU_CONFIG_DIR` support is added here too (it exists only as an unimplemented promise in PROJECT_SPEC.md today; `--config` is currently honored for reads only while `Save()` writes to the hardcoded dir).

---

## 3. Command surface

New reserved top-level words: `ref`, `alias`, `pack`, `onboard`, `which`, `pick`, `browse`.

### 3.1 `cu ref`

```
cu ref set <name> <value>                                # compact: "list:123,list:456"
cu ref set <name> --type list --id 123 --id 456 [-d TEXT]
cu ref set <name> --pick [--type list]                   # interactive multi-picker (P2)
cu ref set <name> --from @a --from list:123 \
    [--assignee @me] [--status S]... [--tag T]... \
    [--due before:+7d] [--sort due] [--order asc]        # presence of --from/filters => query ref
cu ref list [--source all|project|global|pack:<name>] [--json <fields>] [--jq <expr>]
cu ref get <name> [--json]
cu ref rm <name> [--project]
cu ref resolve <name|@name> [--json]      # prints compiled ids / #25 query params — the scripting seam
cu ref check [<name>|--all]               # validates ids against the API; reports archived/dead ids
```

Writes go to global `config.yaml` by default; `--project` targets the discovered `.cu.yml`. Values are validated on write (type prefix required, ids well-formed, name grammar).

### 3.2 Ref consumption — existing commands gain resolution, not new nouns

Every resource-taking flag accepts, in order: **numeric ID → ref name → live name lookup via API** (refs are one lookup table inside the P1 universal resolver). Existing flags gaining resolution: `--list/-l`, `--space/-s`, `--folder/-f`, `--assignee` (real today on task/export commands). **`-w/--workspace` as a general resource flag is NEW P1 surface**: today `-w` exists only on `auth login`/`auth logout`, where it names the keyring workspace slot, not an API resource scope — the general flag is added, not retrofitted. The `@` sigil **forces** ref-namespace lookup (hard error with near-miss suggestions if absent — never a silent literal); `@@x` passes a literal `@x`. Query refs are accepted positionally.

```
cu task list @my-sprint                        # saved query; flags override its filters:
cu task list @my-sprint --due today
cu task list --list support-bugs --status open # bare name: ref wins over a real list of the same name
cu task search "payment failed" --list @eng    # space ref -> space_ids[]
cu task create --list @support-bugs            # multi-id ref: TTY pick; non-TTY error listing members
cu bulk update --list @support-bugs --status closed
    # NEW SURFACE: bulk today takes task IDs positionally/stdin (internal/cmd/bulk.go) and has no
    # --list flag; this adds --list with ref-fan-out-to-task-IDs semantics to bulk.
cu export tasks @my-sprint --format csv        # --format selects csv|json|markdown; -o/--output is the
    # output FILE path today (shadowing the root -o format flag — the -o/--output-file rename is
    # tracked in the ergonomics roadmap, #28)
cu @mine                                       # root sugar: `cu @x ...` == `cu task list @x ...`
```

**Root `cu @ref` sugar — kept, with an explicit cut line (maintainer decision 5):** if `@`-argv parsing complicates cobra arg/completion handling in practice, the fallback is requiring `cu task list @x`; the sugar is two keystrokes of convenience, not load-bearing surface.

### 3.3 Compilation onto the #25 endpoint (hard dependency)

Resource refs fan out into `GET /team/{id}/task` arrays: list refs → `list_ids[]`, space refs → `space_ids[]`, folder refs → `project_ids[]`; mixed-type refs populate several arrays in one call. Query refs additionally compile `where` to `assignees/statuses/tags/due_date_lt|gt/...` and `sort/order` to `order_by/reverse`. Date grammar (deliberately tiny, shared with `--due`): `today|tomorrow|yesterday|overdue|eow|eom|±Nd|±Nw|YYYY-MM-DD` with `before:`/`after:` prefixes.

**SDK note:** the go-clickup SDK (an ordinary module dependency — `github.com/raksul/go-clickup` in go.mod, **not vendored**) has a `GetFilteredTeamTasks` that reuses `GetTasksOptions`, which lacks `list_ids/space_ids/project_ids`. Do not fork: define a cu-local `FilteredTeamTasksOptions` struct (go-querystring tags) in `internal/api` over the SDK's exported `NewRequest`/`Do`, routed through the existing rate limiter. Upstream a PR in parallel. The #25 query builder must accept **id slices** from day one.

**Read/write split** (taskwarrior lesson): read commands accept any ref; commands needing exactly one container (`task create`) require a single resolved list — multi-id/query refs prompt a pick on a TTY, error with a copy-pasteable fix otherwise. A `write:` field on query/multi refs is reserved for later.

**Caveat to document:** ClickUp statuses are per-list; a `statuses:` filter across lists with disjoint status sets matches unevenly. `cu ref check` warns on this.

### 3.4 `cu alias` — gh-identical (fulfills the P2 / PROJECT_SPEC §4.3 promise)

```
cu alias set <name> <expansion> [--shell] [--clobber] [--project]
cu alias set <name> -                       # expansion from stdin (quoting escape hatch)
cu alias list [--json name,expansion,source]   # flags entries shadowed by higher layers
cu alias delete <name> [--project]
cu alias import <file|-> [--clobber] [--project]
```

Mechanics, exactly gh: expand once (no recursion); shlex-tokenize; substitute `$1..$9` (lingering `$N` = "not enough arguments for alias"); append leftover args; `!`-prefixed expansions run `sh -c <body> -- <args...>` with exit-code passthrough. Aliases may reference `@refs`; aliases may not reference aliases. `cu alias set` refuses builtin/reserved names outright; refuses an existing alias without `--clobber`.

**Shell (`!`) legality is source-based, enforced by the alias engine at invocation:** a `!` alias is legal **iff** its source is (a) the user's own global `config.yaml` (implicitly trusted — you wrote it), or (b) a project `.cu.yml` whose `!` bodies match a TOFU-recorded content hash (§6.5). Pack-sourced content can never carry `!` at all — the pack manifest parser rejects it at install (§6.2). There is no grant flow for packs because there is nothing to grant.

### 3.5 `cu pack`

```
cu pack install <owner>/<name> | <git-url> | <local-path> [--pin <tag|sha>] [--force]
    # shorthand: acme/payments -> https://github.com/acme/cu-pack-payments
cu pack list [--json]                    # name, source, rev, pin
cu pack info <name>                      # manifest, refs/aliases inventory, what's shadowed,
                                         # and a static CAPABILITY REPORT (§6.2): e.g. which aliases
                                         # invoke `cu api`, `export`, `bulk`, or other write verbs —
                                         # possible only because pack content is declarative.
cu pack update [<name>|--all] [--dry-run] [--yes]
    # ANY manifest change => semantic diff + re-consent; --yes/non-TTY fail closed on hash mismatch (§6.3)
cu pack remove <name>
cu pack sync [--yes]                     # install/pin everything the project .cu.yml requires;
                                         # CI-safe = applies only exact, previously-consented,
                                         # hash-pinned content — never unreviewed changes
cu pack init [<dir>]                     # scaffold cu-pack.yml + README + LICENSE
```

### 3.6 Extensions — kubectl-style PATH passthrough (PROJECT_SPEC §4.3 as promised)

If dispatch falls through (§5), cu execs `cu-<word>` found on `PATH` with remaining args verbatim, env augmented with `CU_CONFIG_DIR` (the env var cu sets for the child; the general `$CU_CONFIG_DIR` input var is added in the config rework), `CU_WORKSPACE` (resolved), `CU_PROJECT_CONFIG` (path or empty); exit code passes through. No installer, no managed dir at v1 (a managed `cu extension install` channel is demand-gated future work). Shadowed extensions are runnable directly from the shell (`cu-word ...`) — that is the escape hatch. `cu which` warns about shadowing and non-executables.

### 3.7 `cu onboard` — onboarding wizard (§7), `cu which`, `cu pick` (§8), `cu config` additions

```
cu onboard [--global] [--workspace <id|name>] [--space <name>] [--list <name>] [--yes] [--force]
cu which <word>          # explains exactly what `cu <word>` runs; lists everything it shadows, with sources
cu which --reserved      # machine-readable dump of the reserved-now and future-reserved word lists (§5)
cu config unset <key> [--project]
cu config list --show-origin
```

`--json <fields> --jq <expr> --template <tpl>` (P1 anchor) apply uniformly, including `ref list`, `alias list`, `pack list`.

**Dynamic completions** (P2 anchor): alias names complete automatically once registered as commands; `--list/--space/--folder` complete ref names + cached container names via `RegisterFlagCompletionFunc` (cache-only reads, never API-blocking); `cu ref rm/get <TAB>` and `cu alias delete <TAB>` complete from config.

---

## 4. Dispatch mechanics

Adopt gh's **registration model**, not unknown-command interception: at startup, aliases are materialized as real cobra child commands (`DisableFlagParsing: true`), added after builtins so precedence is structural. This gives help, suggestions, and completions for free.

**Required restructure:** cobra resolves the command (and errors) *before* `OnInitialize`/`PersistentPreRunE` run, but cu loads config only in those hooks (`root.go:28-43`). `Execute()` must eagerly: pre-scan `os.Args` for `--config`/`--debug` (as gh's main does), run `config.Init`, register alias commands, then call `rootCmd.Execute()`. The `__complete` path shares this eager load.

## 5. Dispatch precedence and reserved words

Resolution of `cu <word>`, deterministic, exact-match only:

1. **Builtins.** The live cobra tree; never shadowable. Reserved now: `auth api config completion version me task list space user interactive bulk export comment cache docs help` + new `alias ref pack onboard which pick browse`. **Future-reserved** (approved as an API commitment — maintainer decision 4; published, machine-readable via `cu which --reserved`): `extension ext context init workspace folder doc goal field time webhook template status view search mcp dash upgrade`.
   - Naming note (maintainer decision 1): the wizard is `cu onboard`, so `onboard` is reserved **now** and `init` moved to the future-reserved list. Happy side effect: no confusion with the existing `cu config init` subcommand.
2. **`@` sigil sugar.** `cu @name ...` rewrites to `cu task list @name ...`. Zero collision risk — `@` is illegal in command/alias/extension names. Kept with the §3.2 cut line: if it fights cobra's arg/completion handling, fall back to requiring `cu task list @x`.
3. **Aliases**, merged project > global > packs (project `packs:` array order, then global install order).
4. **PATH extension** `cu-<word>` (checked only after `Find` fails; `exec.LookPath`, ~100 lines, no startup PATH scan).
5. **Error** with did-you-mean suggestions across all namespaces, plus — when the word exists only in a pack the project requires but isn't installed — the hint `run: cu pack sync`.

**Collision policy (validated at write/install time so runtime is always deterministic):**
- `cu alias set`: refuses builtin/future-reserved names ("`X` is already a cu command"); `--clobber` only for alias-over-alias.
- Pack install: entries colliding with builtins/reserved words are **marked inert** with a warning (visible in `cu pack info`); cross-pack collisions: first-in-order wins, every conflict printed with its winner; user/project shadowing a pack entry is not a conflict — it is the chain working, and `--show-origin` / `cu which` display it.
- New builtins in a release shadow same-named aliases: release notes list newly reserved words; a once-per-version check warns about newly shadowed entries instead of failing.

---

## 6. Packs: format and trust model

### 6.1 Format — one file, declarative-inert

Repo `cu-pack-<name>` (auto-prefix, brew-style; full git URLs and local paths accepted for private hosts and development). Everything lives in a single fail-closed manifest:

```yaml
# cu-pack.yml  (required at repo root; unknown keys REJECTED)
name: payments                  # must equal repo suffix; install fails otherwise
version: 1.2.0                  # informational; pinning uses the locked commit SHA (§6.3)
description: Payments team ClickUp shortcuts
min_cu_version: "0.3.0"         # soft check, warn
workspace_hint: "Acme Inc"      # matched at onboard; refs error clearly under the wrong -w

refs:
  pay-bugs: "list:901304567,list:901304890"
  pay-triage:
    from: ["@pay-bugs"]
    where: { statuses: ["open"], tags: ["bug"] }
    sort: due

aliases:
  paybugs: "task list @pay-bugs --status open"
  triage: "task update $1 --list @pay-bugs --priority 1"
  # `!` shell aliases are ILLEGAL in packs. The manifest parser REJECTS any `!` alias at
  # install time — fail-closed, loud error naming the offending alias. See §6.2.

defaults:                       # lowest precedence; user/project/global all beat these.
  output: table                 # credential keys (api_token, ...) are IGNORED with a loud
  default_space: "Payments"     # warning if present (§2.4 blocklist).
```

No hooks, no bundled binaries, no pack dependencies, no registry, and **no signing at v1 — signing/registry are explicitly deferred** (revisit if a pack ecosystem materializes). Discovery = GitHub topic `cu-pack` + explicit `owner/repo` installs; provenance (`owner/repo`, full resolved URL at first install) is always displayed, never bare names.

### 6.2 Trust model — the definitive decision

**Packs are 100% declarative-inert (maintainer decision 2). A pack can never carry an executable surface of any kind.** The manifest parser rejects `!` shell aliases at install — fail-closed, loud error. There is no grant machinery: no `--allow-shell`, no `shell_sha256` for packs, no blocked/granted states, no grant revocation, because there is nothing to grant. The threat we defend against is transitive code execution from shared configuration:

1. **Clone-and-run** — cloning a repo whose `.cu.yml` carries hostile content must never execute code without an explicit, content-bound consent event (the oh-my-zsh dotenv-RCE shape). Defended by the project-file TOFU gate (§6.5) — which stays **independent of packs**.
2. **Upstream compromise** — a compromised pack repo must not turn `cu pack update`/`cu pack sync` into fleet-wide compromise (the asdf/tpm git-pull shape; the reason mise repudiated executable plugins). Defended by rev pinning + the manifest hash (below): changed content is never applied without a reviewed diff.
3. **Shadowing** — shared content must never redefine a builtin (VS Code's "no permission model" lesson). Defended by install-time inert-marking (§5).

We explicitly do **not** defend against: a user knowingly running a malicious `cu-<name>` PATH binary (kubectl posture — their PATH, their trust), or *misdirection within data scope* (a pack aiming a ref at the wrong list — bounded by the user's own API token privileges and visible in diffs).

**The single trust mechanism — manifest-wide content hash:**

- At install, the user consents to the pack's full content (inventory + capability report shown); a **`manifest_sha256` over the entire parsed manifest** is recorded in `packs.lock.yml` alongside the locked commit SHA.
- On update/sync, **ANY manifest change requires re-consent showing the semantic diff** (refs added/removed/changed with old→new ids, alias bodies changed, defaults changed) — the audit advantage no executable-plugin system can offer.
- **`--yes` and non-TTY modes fail closed on hash mismatch**: they apply only lockfile-pinned, hash-verified content. "CI-safe" means "applies only the exact, previously-reviewed, hash-pinned content" — never silently-updated content. (In CI the token is often high-privilege, so unattended application of changed content would be a worse confused deputy than interactive use.)

This replaces the earlier draft's shell-grant machinery entirely and closes both review findings at once: the declarative channel is integrity-pinned (not just `!` bodies), and there is no shell channel in packs to grant.

**Residual, consciously bounded risk:** a declarative alias can still invoke powerful token-backed builtins — `cu api`, `cu export` (writes files via `--output`), `cu bulk`. This is bounded by three properties rather than a grant flow: (a) install-time consent covers the full inventory; (b) the manifest hash makes post-consent mutation impossible without a re-reviewed diff; and (c) **`cu pack info` prints a static capability report** — e.g. exactly which aliases invoke `cu api` or other write verbs — which is *possible only because pack content is declarative* (an executable plugin cannot be statically characterized). Exfiltration via `cu api` is additionally bounded to ClickUp's own API host (the base URL is hardcoded in `internal/cmd/api.go`).

**Escape hatches — where executable team workflows live instead (maintainer decision 2):**

1. **Project `.cu.yml`** — the PR-reviewed home for team shell macros: `!` aliases are legal there under the TOFU content-hash gate (§6.5), and changes arrive through code review like any other repo change.
2. **`cu-<name>` PATH extensions** — the clearly-labeled executable channel (§3.6), never conflated with packs.
3. **A future additive declarative `url:` alias type** — covering the "open the oncall page" class of workflow without shell; additive and compatible.

**Asymmetry rationale:** adding a grant mechanism later is a compatible loosening; removing shell support after packs adopted it would break published packs. Starting inert is the only reversible position.

### 6.3 Pin and fetch integrity

**Pinning — the commit SHA is the sole authority.** Tags are human labels only. `--pin <tag|sha>` resolves to a commit SHA at install; the lockfile records both, but every load and every `sync` **verifies the checked-out tree against the locked `rev`** and fails closed on mismatch. A tag that re-resolves to a different SHA (force-pushed tag — a standard update-hijack vector) is treated as an **explicit re-pin event**: semantic diff + consent, never a silent advance. `cu pack update` never advances a recorded `rev` without an explicitly accepted diff. Pin-by-commit is the preferred recorded form.

**Git fetch hardening** (greenfield code — specified now, before any clone code exists). The fetch step runs git, which is itself an RCE surface regardless of the repo content being declarative:

- Clone with submodules disabled (`--no-recurse-submodules`; malicious `.gitmodules` / `ext::sh -c` shapes).
- Shallow clone (`--depth 1`).
- Protocol allowlist: `https`, `ssh`, and local paths only (`-c protocol.allow` configuration; no `ext::`, no arbitrary schemes).
- Reject any source URL beginning with `-` (argument injection), and always pass `--` to separate arguments from URLs.
- Clone into a temp dir, validate the manifest (fail-closed parse, size caps), then move into `$XDG_DATA_HOME/cu/packs`.
- No symlink escape from the pack dir at load time.

### 6.4 Source identity — TOFU on the pack source

First install records the **full resolved source URL** for the pack name in the lockfile. Any subsequent operation that would use a *different* source for the same pack name requires explicit confirmation — this blunts typosquat/homoglyph swaps (`acme/payment` vs `acme/payments`) behind a familiar name. The `name == repo suffix` check stops self-mislabeling only, not impersonation; full provenance is always displayed at first install and in `cu pack list/info`. Signing and a registry/namespace reservation are **explicitly deferred**; the featured-pack list, when it exists, lives in a maintainer-controlled file rather than raw topic scraping.

### 6.5 Lockfile

```yaml
# ~/.config/cu/packs.lock.yml (machine-written, mode 0600)
version: 1
packs:
  payments:
    source: https://github.com/acme/cu-pack-payments   # TOFU-recorded at first install (§6.4)
    pin: v1.2.0                  # human label only
    rev: a1b2c3d4                # sole authority; verified against the tree on every load/sync (§6.3)
    manifest_sha256: "9f2c..."   # hash of the entire parsed manifest; any drift => re-consent (§6.2)
    installed_order: 3
trusted_projects:                # project .cu.yml TOFU gate — INDEPENDENT of packs; stays as-is
  /Users/tim/dev/acme-app:
    shell_sha256: "be91..."      # content hash over the project file's `!` alias bodies at consent time;
                                 # any change re-prompts with the changed bodies
```

The lockfile is written `0600` and is **tamper-evident, not tamper-proof**: anything already running as the user can edit it — the same trust boundary as the cu binary itself. Defense-in-depth: hashes are re-derived and verified against live content at each use (a recorded hash that no longer matches re-prompts), and detected lockfile inconsistency is treated as revoke-all, not trust-the-file.

### 6.6 Team flow

Lead: `cu pack init`, fill YAML, push, tag, add `packs: [acme/payments@v1.2.0]` to the team repo's `.cu.yml`. Teammate: clone, `cu onboard` (detects the requirement, offers `cu pack sync`), done — `cu paybugs` and `@pay-bugs` mean the same thing to every human and every script on the team. Nothing is ever auto-installed on directory entry or config discovery. Team shell macros, when needed, live in the PR-reviewed `.cu.yml` (§6.2 escape hatch 1), not in the pack.

---

## 7. `cu onboard` — onboarding wizard

Named `cu onboard` (maintainer decision 1; `init` stays future-reserved, and the name avoids any confusion with the existing `cu config init`). Project mode by default (writes `.cu.yml`), `--global` writes `~/.config/cu/config.yaml`. Fully scriptable (`--workspace/--space/--list --yes`); on a non-TTY without `--yes` it errors and prints the exact flag form. Re-runs **merge**: only prompted keys update, other keys and comments preserved (yaml.v3 writer), final file previewed before write.

```
[0/6] Pack check      — if the discovered .cu.yml lists packs: offer `cu pack sync`
                        (install/consent UX per pack, §6); workspace_hint pre-answers step 2.
[1/6] Auth check      — not authenticated? run `cu auth login` inline; greet by name.
[2/6] Workspace       — fuzzy single-select (pre-selects anything already resolvable, saying why:
                        "from CU_WORKSPACE").
[3/6] Space           — fuzzy single-select, skippable.
[4/6] Default list    — fuzzy single-select across the space (folder-aware), skippable.
[5/6] Named refs      — optional loop: multi-select lists -> name the ref (slug suggested, editable).
[6/6] Starter aliases — offered with visible expansions (bugs, mine); never shell aliases.

Preview merged YAML -> confirm -> write.
Close: cheat sheet of what now works (`cu bugs`, `cu @mine`, `cu task list`) and the funnel line:
  project mode: "commit .cu.yml to share these defaults with your team"
  global mode:  "team lead? `cu pack init` turns this setup into a shareable pack"
```

Standalone value: steps 1–6 need no pack anywhere. Pack-aware value: step 0 makes `git clone && cu onboard` the entire new-teammate story. Wizard lookups share the resolver/cache code paths (§3.2) — onboard is a client of the resolver, not a parallel system. No `onboard.yml` DSL in v1; if real packs demonstrate the need, a constrained step schema can ship later without reopening the trust model.

---

## 8. Pickers and TUI

**Decision (maintainer-accepted as specced): bubbletea + bubbles now, behind an interface; no external fzf integration ever; `cu browse` later in core; promptui exits.**

- promptui is disqualified for new work by evidence: no multi-select primitive (onboard is fundamentally multi-select), unmaintained (v0.9.0, 2018 readline pin), substring-search only. Building custom widgets on a dead library to defer a dependency we need for `browse` anyway is a double migration.
- **Now (P2):** `internal/ui` package defining `Pick / PickMulti / Confirm / Input`, backed by bubbletea + bubbles (fuzzy-filtered list, checkbox multi-select). Cost accepted: ~12–15 transitive modules, +2–3 MB binary, actively maintained. Surfaces: the `cu onboard` wizard, `cu ref set --pick`, `cu pick`, TTY disambiguation (multi-id ref on a single-target command), pack consent prompts. Every prompt is flag-bypassable and fails informatively on non-TTY — scriptability is never hostage to interactivity. Existing `interactive.go` promptui screens migrate opportunistically; promptui is then dropped from go.mod.
- **`cu pick`** — the composable primitive: `cu pick <task|list|space|folder|user|status> [@ref|--list X|--space Y] [-m] [--format id|name|url|json]`, selections to stdout one per line. This plain-stdout contract is the deliberate bridge for fzf devotees (`cu task view $(cu pick task @mine)` or pipe through fzf themselves) — cu neither depends on nor shells out to fzf: one picker, one behavior, one support surface, works on a fresh Windows box.
- **Later (post-P2): `cu browse [@ref]`** in core — the one full-screen bubbletea command: left pane refs/starred lists, right pane task table streaming from #25; `/` filter, Enter detail, `c` close, `e` status, `a` assign, `y` yank id, `o` open, `r` refresh. Hard boundary, enforced in review: browse is a veneer over the same resolver and command paths — nothing is doable in the TUI that isn't a CLI one-liner. `cu interactive` is left untouched until browse proves out, then becomes a documented alias.

---

## 9. Decision log

### 9.1 Maintainer decisions (final)

1. **The wizard is `cu onboard`, not `cu init`.** Renamed throughout. `onboard` moves into the reserved-now builtin list; `init` moves to the future-reserved list. Happy side effect: no confusion with the existing `cu config init`.
2. **No shell (`!`) aliases in packs — packs are 100% declarative-inert.** The manifest parser rejects `!` aliases at install (fail-closed, loud error). The grant machinery is deleted entirely (no `--allow-shell`, no pack `shell_sha256`, no blocked/granted states, no grant revocation); the lockfile becomes source/pin/rev/manifest_sha256 (+ `trusted_projects` unchanged). The alias engine rule: `!` is legal iff the source is global config or a TOFU-content-hash-trusted project `.cu.yml` — the project TOFU machinery stays (it defends clone-and-run independent of packs). `cu pack info` prints a static capability report. Escape hatches: project `.cu.yml` (PR-reviewed team shell macros), `cu-<name>` PATH extensions, a future additive declarative `url:` alias type. Asymmetry rationale: adding grants later is compatible; removing shell after adoption would break packs. Pack work resized L → M in sequencing.
3. **bubbletea/bubbles in core is accepted as specced** (§8).
4. **The future-reserved word list is approved** (with the init/onboard swap from decision 1) **and is an API commitment**; `cu which --reserved` stays.
5. **Root-level `cu @ref` sugar is kept**, with an explicit cut line: if `@`-parsing complicates cobra arg/completion handling, the fallback is requiring `cu task list @x`.

### 9.2 Security-review foldins (all incorporated)

- Manifest-wide `manifest_sha256` in the lockfile as the **single trust mechanism**; any manifest change on update/sync requires re-consent with a semantic diff; `--yes`/non-TTY fail closed on hash mismatch, applying only lockfile-pinned, hash-verified content (§6.2, §6.3).
- Commit SHA in the lockfile is the sole pin authority; tags are human labels; every load/sync verifies the tree against the locked rev; a re-resolved tag is an explicit re-pin event with diff + consent (§6.3).
- Git fetch hardening: no submodules, `--` separation, protocol allowlist (https/ssh/local), reject `-`-prefixed source URLs, shallow clone, temp-dir validation, no symlink escape (§6.3).
- Credential-key blocklist: `api_token` (and any future credential keys) ignored with a loud warning in pack `defaults:` and project `.cu.yml` layers; credentials only from keyring/env/global config (§2.4).
- Lockfile written 0600; documented as tamper-evident, not tamper-proof — same trust boundary as the binary itself; inconsistency = revoke-all (§6.5).
- TOFU on pack source identity: first install records the full source URL; a changed source for the same pack name requires explicit confirmation (typosquat/homoglyph blunting). Signing/registry explicitly deferred (§6.4).

### 9.3 Fact-check corrections (all incorporated)

- The global config file is `config.yaml`, not `config.yml` (`internal/config/config.go:28-30,81`); the `root.go:46` help text saying `config.yml` is a drive-by fix candidate (§2.1).
- `$CU_CONFIG_DIR` does not exist today — marked "to be added in the config rework" wherever mentioned (§2.1, §2.5, §3.6).
- `-w/--workspace` as a general resource flag is new P1 surface (today only on `auth login/logout`, naming the keyring slot) (§3.2).
- The bulk example requires adding a `--list` flag with ref-fan-out semantics to `cu bulk` — annotated as new surface (§3.2).
- The export example corrected to `--format csv` (format selector is `--format`, `-o` is the output file path; the `-o/--output-file` rename is tracked in the ergonomics roadmap, #28) (§3.2).
- The go-clickup SDK is an ordinary module dependency, not vendored (§3.3).
- Env bindings `CU_WORKSPACE/CU_SPACE/CU_LIST` are new; today's AutomaticEnv yields `CU_OUTPUT/CU_DEBUG/CU_DEFAULT_*` (§2.4).

### 9.4 Contradictions with the input designs (design-review record)

`cu ref` noun added (vs A); `cu pack` split from extensions (vs A); PATH extensions v1 (vs A); bubbletea in core for pickers (vs A, C); `cu browse` in core (vs A); registration dispatch + Execute() hoist (vs A, B); no sticky contexts v1 (vs B); refs/aliases as config keys, not separate files (vs B; PROJECT_SPEC §4.3 amended); no query-ref params v1 (vs B); no pack priorities (vs B); no onboard.yml DSL v1 (vs C); tag/sha pins not semver ranges (vs C); wizard naming — the synthesis draft chose `cu init` (vs B, C), **reversed by maintainer decision 1: it is `cu onboard`**; first-wins pack collisions (vs A); pack clones in XDG data dir (vs B, C); aliases pulled into the scripting-contract release (vs A's strict P2); **pack shell aliases removed entirely** (vs the synthesis draft's grant model — maintainer decision 2, reinforced by the security review).

### 9.5 Still open (non-blocking)

- `cu interactive` end-state: migrate onto `internal/ui` then alias to `cu browse` post-P2, or deprecate outright once `cu pick` + refs cover it.
- Upstreaming: whether to maintain the cu-local `FilteredTeamTasksOptions` indefinitely if raksul/go-clickup declines the `list_ids[]/space_ids[]/project_ids[]` PR, or move the task-query path to a minimal internal client.

---

## 10. Sequencing

Dependency-ordered, mapped to the existing P0/P1/P2 anchors. Sizes S/M/L. Edges written as `X → Y` (X blocks Y).

**P0 — ships with/before the #25 rebuild**

- 0.1 Remove `.` from viper search path (S). Trust prerequisite; already planned. → 2.5, §6 gating
- 0.2 Config layering rewrite: per-layer stores replacing the `viper.Set()` merge (with the credential-key strip), yaml.v3 surgical writer, `Execute()` config hoist with argv pre-scan, `cu config unset` / `--show-origin` / `--project`, `$CU_CONFIG_DIR` (M). Keystone — also fixes the live project→global config-bleed bug. → everything below
- 0.3 #25 Get Filtered Team Tasks rebuild (already scoped) shaped to accept id slices, plus the cu-local `FilteredTeamTasksOptions` over the SDK's exported `NewRequest`/`Do` (+S delta inside #25). → 1.2 multi-list reads, 2.1, 3.1

**P1 — the scripting-contract release** (the stable surface scripts and AI agents build against; everything here depends on 0.2, and multi-list ref *reads* depend on 0.3, so this release ships after #25):

- 1.1 `internal/resolve` universal resolver (P1 anchor): ID → ref table → live name, one chokepoint for all resource flags; adds the general `-w/--workspace` resource flag (new surface) (M). [0.2, 0.3 → 1.1]
- 1.2 Resource refs: `refs:` schema, `@` sigil + `@@` escape + root `cu @ref` sugar (with the decision-5 cut line), `cu ref set/list/get/rm/resolve/check`, single-target write guard (M). [1.1 → 1.2]
- 1.3 `--json/--jq/--template` (P1 anchor, as scoped) — applied uniformly incl. ref/alias/pack list (M).
- 1.4 `cu alias` engine + CRUD, gh-exact semantics, source-based `!` legality (global implicit, project TOFU gate, packs never), registration-based dispatch (M). Pulled forward from P2: depends only on 0.2, not #25, and packs need it. [0.2 → 1.4]
- 1.5 `cu which` incl. `--reserved` (S). [1.4 → 1.5]

Contract at this release: ref vocabulary, alias vocabulary, `--json/--jq`, `cu ref resolve`, `cu which`, `--show-origin`. Frozen henceforth — packs serialize exactly these schemas.

**P2 — after the scripting-contract release**

- 2.1 Query refs + minimal date grammar, flag-override semantics (M). [1.2, 0.3 → 2.1]
- 2.2 `cu-<name>` PATH extension passthrough + env injection (S). [0.2 → 2.2]
- 2.3 `internal/ui` pickers on bubbletea+bubbles; promptui migration begins (M). → 2.4, 2.6, 3.1
- 2.4 `cu pick` (S). [2.3, 1.1 → 2.4]
- 2.5 `cu pack`: install/list/info/update/remove/sync/init, single-file manifest with install-time `!` rejection, lockfile with manifest-hash pinning + re-consent diff flow, source TOFU, git fetch hardening, pack config layers, collision handling, capability report (M — resized from L per maintainer decision 2: deleting the shell-grant machinery removed the largest trust-UX surface; ships only after 1.2/1.4 schemas are stable). [0.1, 0.2, 1.2, 1.4 → 2.5]
- 2.6 `cu onboard` wizard, standalone + pack-aware step 0 (M). [2.3, 1.2 → 2.6; pack step needs 2.5 but standalone mode can ship before it]
- 2.7 Dynamic completions for aliases (free via registration), refs, extension names; cache-only (S). [1.2, 1.4 → 2.7]

**Post-P2**

- 3.1 `cu browse [@ref]` bubbletea query browser; then alias `cu interactive` to it (L). [2.3, 2.1 → 3.1]
- 3.2 Pack discovery via `cu-pack` GitHub topic / maintainer-controlled featured list (S).
- 3.3 Managed `cu extension install` channel — only if PATH-model demand appears (M).
- 3.4 Demand-gated: onboard step schema for packs, sticky contexts, ref `write:` targets, semver-range pins, alias/query-ref shared param engine, declarative `url:` alias type (§6.2 escape hatch 3).
