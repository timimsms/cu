# cu – ClickUp Command‑Line Interface

## 1. Vision & Motivation

Develop a first‑class command‑line interface for ClickUp users—``—that mirrors the developer‑centric ergonomics of GitHub CLI (`gh`). The tool should let engineers and agentic workflows create, query, and manipulate ClickUp tasks, lists, spaces, and other resources as easily as they handle repositories, issues, and pull‑requests on GitHub. This specification (“PROJECT\_SPEC”) defines scope, functionality, design, and milestones for an initial open‑source release intended for distribution via **npm** and **Homebrew**.

---

## 2. Goals

1. **Parity with GitHub CLI** – Provide analogous commands for the most‑used `gh` sub‑commands (e.g., `gh issue list` → `cu task list`).
2. **Low‑friction Auth** – Support both Personal API tokens and full OAuth 2.0 device‑flow in a manner similar to `gh auth login`.
3. **Project‑Scoped Defaults** – Allow a project to declare a default Space / Folder / List (via `.cu` file or `cu config set …`).
4. **Extensibility** – Encourage user‑authored sub‑commands and scripting hooks (via an `extensions/` directory or npm packages).
5. **Agent‑friendly Output** – Offer JSON, TSV, and human‑readable formats, with machine‑parseable output as first‑class.
6. **Cross‑platform Packaging** – Ship binaries via npm (Node wrapper) *and* Homebrew formula, plus standalone archives for CI jobs.

### Non‑Goals (v1)

- Managing ClickUp Docs or Whiteboards.
- UI rendering of views (Kanban, Calendar, etc.).
- Real‑time sync / watch mode; polling is acceptable initially.

---

## 3. Domain Model & Terminology

| ClickUp Hierarchy         | Equivalent GH Concept                | `cu` Handle | Notes                                                   |
| ------------------------- | ------------------------------------ | ----------- | ------------------------------------------------------- |
| Workspace (Team)          | GitHub Org                           | `team`      | Auth scope; rarely needed in everyday commands.         |
| Space                     | GH Project (classic)/Repo Group      | `space`     | Logical top‑level container.                            |
| Folder (legacy *Project*) | GH Repo Project ‑or‑ Label Milestone | `folder`    | Optional layer.                                         |
| List                      | GH Issue Milestone                   | `list`      | Tasks live here; `list‑id` required for most write ops. |
| Task                      | Issue/PR                             | `task`      | Core unit.                                              |
| Subtask                   | Issue Checklist Item                 | `subtask`   | Supported in API v2.                                    |

---

## 4. Command Taxonomy

> **Naming**: Short alias `cu` preferred, but fallback `clickup` or `ck` MUST be supported via install flag because of the *POSIX* `cu` utility conflict (see §11).

### 4.1 Top‑level Commands (MVP)

| Command            | Description                                                                                                     | Notes     |                                    |                                                                                            |
| ------------------ | --------------------------------------------------------------------------------------------------------------- | --------- | ---------------------------------- | ------------------------------------------------------------------------------------------ |
| \`cu auth \<login  | logout                                                                                                          | status>\` | Manage tokens / OAuth device flow. |                                                                                            |
| \`cu config \<list | get                                                                                                             | set       | edit>\`                            | Per‑user and per‑project settings stored in `$CU_CONFIG_DIR` (defaults to `~/.config/cu`). |
| `cu task <sub>`    | CRUD and query tasks. Sub‑commands: `list`, `create`, `view`, `comment`, `close`, `reopen`, `assign`, `delete`. |           |                                    |                                                                                            |
| `cu list <sub>`    | Work with Lists: `list`, `create`, `view`, `archive`, `default`.                                                |           |                                    |                                                                                            |
| `cu space <sub>`   | View or create Spaces; set project default.                                                                     |           |                                    |                                                                                            |
| `cu folder <sub>`  | Manage Folders if the Workspace uses them.                                                                      |           |                                    |                                                                                            |
| `cu me`            | Show current user info (parity with `gh api user`).                                                             |           |                                    |                                                                                            |
| `cu api`           | Raw passthrough to any ClickUp REST endpoint (advanced).                                                        |           |                                    |                                                                                            |
| `cu completion`    | Shell completion scripts (bash                                                                                  | zsh       | fish                               | powershell).                                                                               |
| `cu version`       | Print version / update notice.                                                                                  |           |                                    |                                                                                            |

### 4.2 Flag Conventions

- `--json <fields>` ⇒ explicit machine output.
- `--team`, `--space`, `--folder`, `--list` ⇒ override defaults.
- `--assignee`, `--status`, `--tag`, `--due`, `--priority` for task filters.
- Global `--format` (`json`, `yaml`, `table`, `short`).

### 4.3 Alias & Extension System

- Aliases stored in `~/.config/cu/aliases.yml`.
- Executables on `$PATH` matching pattern `cu-<name>` auto‑loaded (`kubectl` model).

---

## 5. Configuration Strategy

1. **Runtime discovery**
   - `$CU_CONFIG_DIR` env var overrides default path.
2. **Project Dotfile**
   - `.cu.yml` at repo root may specify:
     ```yaml
     default_space: "Engineering"
     default_folder: "Sprint Backlog"
     default_list: "Bugs"
     output: json
     ```
3. **Interactive **``** Wizard** (optional v1.x)

---

## 6. Authentication & Authorization

| Token Type               | Flow                                                             | Command UX              |
| ------------------------ | ---------------------------------------------------------------- | ----------------------- |
| Personal API Token       | Prompt or `--token` flag; stored encrypted in keychain/key‑ring. | `cu auth login --token` |
| OAuth Device Code        | Open device URL; poll token.                                     | `cu auth login --web`   |
| Re‑auth / multiple teams | Host‑scoped secrets with `--team-id` flag.                       |                         |

- Respect ClickUp rate limits (100 RPM on Free plan). Implement exponential back‑off.

---

## 7. Implementation Overview

### 7.1 Language & Runtime

| Option                    | Pros                                                            | Cons                                                                      |
| ------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------- |
| **Ruby (Thor or CmdKit)** | Aligns with team preference; easy gem build; clean DSL for CLI. | Requires packaging Ruby runtime for Homebrew binary; concurrency limited. |
| Go (Cobra)                | Static binary, used by `gh`; cross‑platform; easy Homebrew.     | Larger learning curve; binary size.                                       |
| Node (oclif)              | Native npm; extension system via JS; rapid dev.                 | Requires Node env for brew unless packaging with pkg.                     |

**Recommendation**: Pilot in **Go** for parity with `gh`, then expose lightweight Ruby wrappers via FFI if desired.

### 7.2 Core Modules

1. **API Client** – Typed wrapper over v2 endpoints.
2. **Config Manager** – Merges env vars, dotfile, global config.
3. **Formatter** – Table/JSON/YAML renderers.
4. **Prompt UI** – Interactive menus (Cobra‑survey or Inquirer).
5. **Cache** – Local LRU for ID↔Name lookups (reduces API calls).

### 7.3 Error Handling

- Map HTTP errors to friendly messages.
- For 429, show wait timer.

---

## 8. Output Specification

- **Human** default: concise tables (trunc names, relative dates).
- **Machine**: `--json` returns canonical structures identical to API payload (snake\_case).

Sample JSON for `cu task view 123`:

```json
{
  "id": "123",
  "name": "Fix login bug",
  "status": "open",
  "url": "https://app.clickup.com/t/123",
  "assignees": ["42"],
  "tags": ["bug", "login"],
  "due_date": "2025-07-01T00:00:00Z"
}
```

---

## 9. Security & Privacy

- Store tokens in OS‑native keychain (`security`, `secret-tool`, `wincred`).
- Obfuscate token in logs; redact PII.
- Opt‑in telemetry for crash reports only.

---

## 10. Packaging & Distribution

| Target       | Strategy                                                                                                                  |
| ------------ | ------------------------------------------------------------------------------------------------------------------------- |
| **Homebrew** | Publish formula `clickup-cli` in a tap `yourorg/homebrew-clickup`. Support rename symlink `cu` unless conflict detected.  |
| **npm**      | Bundle with `pkg` to ship pre‑built binaries; publish as `@yourorg/clickup-cli`. Avoid package name `cu` (already taken). |
| **Docker**   | Alpine‑based image with `cu` pre‑installed for CI.                                                                        |

---

## 11. Naming Collision Risk

- *POSIX* utility `` (Call UNIX) is present in macOS and Linux distributions. Shipping an identically named binary will shadow the system command. Mitigations:
  1. Install as `clickup` and create opt‑in symlink `cu` (with warning).
  2. Provide `--install-symlink` flag for brew/npm post‑install script.
- **npm** package `cu` already exists (Copper Config). Reserve `clickup-cli` or `cu-cli`.

---

## 12. Roadmap & Milestones

| Version   | Target Date | Scope                                                         |
| --------- | ----------- | ------------------------------------------------------------- |
| **0.1.0** | Aug 2025    | Auth flow, `task list/view/create`, config file, JSON output. |
| **0.2.0** | Sep 2025    | `list` management, shell completions, CI Docker image.        |
| **0.3.0** | Oct 2025    | Extensions API, interactive prompts, rate‑limit back‑off.     |
| **1.0.0** | Dec 2025    | Full parity with defined MVP commands, extensive tests, docs. |

---

## 13. Open Questions

1. Support ClickUp “Custom Fields” in JSON output? How to select columns?
2. Should `cu pull` mirror `gh pr` with a concept of “merge request” (ClickUp doesn’t have PRs)?
3. Potential integration with Git hooks to auto‑open tasks on branch creation.

---

## 14. Acceptance Criteria

- Successfully create/list tasks from terminal with default config.
- Pass end‑to‑end tests against ClickUp demo workspace.
- Publish on npm & brew with an automated GitHub Actions release pipeline.

---

**END OF SPEC**

