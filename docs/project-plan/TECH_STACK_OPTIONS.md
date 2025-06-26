# cu – Tech‑Stack Matrix & Evaluation

## 1. Purpose
This document compares candidate implementation stacks for the **`cu` ClickUp CLI** across a common set of criteria—including packaging, performance, ecosystem support, and developer experience—to guide an informed decision for the MVP and long‑term roadmap.

## 2. Evaluation Criteria
| # | Criterion                           | Rationale                                                                                             |
|:-:| ----------------------------------- | ------------------------------------------------------------------------------------------------------ |
| 1 | **Binary distribution**             | Ability to ship single‑file executables for macOS, Linux, Windows; Homebrew formula ease.             |
| 2 | **npm publish path**                | Feasibility of distributing via `npm` (either source or packaged binary).                              |
| 3 | **Extensibility / plugin model**    | Support for user‑authored sub‑commands (`cu‑foo`) & dynamic loading.                                   |
| 4 | **Existing ClickUp SDK**            | Availability & maturity of community libraries to reduce API boilerplate.                             |
| 5 | **CLI framework maturity**          | Ecosystem, stability, learning curve (e.g., Cobra, oclif, Thor…).                                      |
| 6 | **Cross‑compile workflow**          | CI pipelines for multi‑platform builds; static linking hurdles.                                        |
| 7 | **Performance & footprint**         | Runtime speed and binary size (cold start on CI/agent VMs).                                           |
| 8 | **Team familiarity**                | Alignment with in‑house expertise (Ruby strong, Go moderate, Node strong, etc.).                      |
| 9 | **Community & longevity**           | Project activity, corporate backing, future‑proofing.                                                 |
|10 | **Security & supply chain**         | Risk surface (runtime CVEs, native dependencies, signing, SBOM ease).                                 |

## 3. Side‑by‑Side Matrix
| Criterion \ Stack            | **Go (Cobra + Viper)** | **Node (oclif)** | **Ruby (Thor / CmdKit)** | **Rust (clap + cargo)** | **Python (Typer + rich‑cli)** |
|------------------------------|------------------------|------------------|--------------------------|------------------------|-------------------------------|
| **1. Binary distribution**   | ✅ Static, ~7 MB; built‑in cross‑compile; easy Homebrew bottle | ⚠️ Requires pkg/​nexe for single file; ~25 MB; Brew via tarball | ⚠️ Ruby runtime needed or embed via ruby‑app bundle; formula possible | ✅ Static MUSL possible; ~2‑4 MB; Homebrew accepted | 🔸 PyOxidizer / PEX possible but heavy; less common in Brew |
| **2. npm publish**           | ✅ Ship prebuilt binaries as tarballs like `gh` | ✅ Native; simplest | ⚠️ Uncommon; possible via executable gem & node wrapper | ⚠️ Needs pre‑built bins; uncommon | 🔸 Possible via packaged bin; heavy |
| **3. Extensibility model**   | ✅ `cu‑xyz` on PATH auto‑loads; mimic kubectl | ✅ oclif plugins & hooks | ✅ Thor sub‑command loading; extensions via gems | 🔸 Custom; clap supports sub‑command plugins with extra work | 🔸 Typer not opinionated; requires plug‑in design |
| **4. ClickUp SDK**           | 🟢 `go-clickup`, `clickup-client-go` mature | 🟢 `@clickup/rest-client`, `clickup-node` | 🔸 No maintained gem | 🔸 No first‑party crate | 🔸 Unofficial `clickup.py` low activity |
| **5. CLI framework**         | 🟢 Cobra adopted by `gh`, `kubectl`; strong docs | 🟢 oclif backed by Heroku/Salesforce; active | 🟡 Thor stable but aging; CmdKit new | 🟢 clap 4.x is fast, ergonomic | 🟡 Typer popular but young |
| **6. Cross‑compile CI**      | ✅ `goreleaser` templates; multi‑arch | ⚠️ pkg builders per‑arch; slower | ⚠️ rubyc or Traveling Ruby; brittle | ✅ cross + musl; cargo‑zigbuild | 🔸 PyInstaller w/ manylinux wheels |
| **7. Perf / footprint**      | 🟢 Fast start; small RAM | 🟡 Node warm‑up; bigger RAM | 🟡 Interpreted; slower; medium RAM | 🟢 Native; fastest; tiny RAM | 🟡 Interpreted; slower; bigger RAM |
| **8. Team familiarity**      | 🟡 Moderate | 🟢 High (JS/TS common) | 🟢 Very high | 🟠 Low‑moderate | 🟡 Moderate |
| **9. Community**             | 🟢 Large; longstanding | 🟢 Large; Salesforce | 🟡 Ruby shrinking | 🟢 Growing; crates.io | 🟢 Large; Python |
| **10. Security supply chain**| 🟢 Static deps; easy SBOM | 🟡 Node CVEs frequent; supply chain risk | 🟡 Ruby CVEs fewer but gem signing rare | 🟢 Cargo audit; reproducible | 🟡 Many linux wheels; CVEs frequent |

Legend: ✅ best‑in‑class • 🟢 strong • 🟡 adequate • 🔸 weak • 🟠 limited • ⚠️ trade‑off.

## 4. Deep‑Dive Notes
### 4.1 Go + Cobra
* **Packaging**: Single binary per OS/arch; Homebrew core already ships `cobra-cli` formula, demonstrating compatibility (see Formula page).
* **Existing SDKs**: Unofficial `go-clickup` covers most API surface and is actively maintained.
* **Release automation**: `goreleaser` can cross‑compile & push Homebrew tap & npm tarballs in one workflow.

### 4.2 Node + oclif
* **Developer UX**: Fast iteration in TypeScript with hot‑reload; rich generator & plugin API.
* **Distribution**: Need `pkg` to bundle; Homebrew guides exist but larger binary size.
* **ClickUp SDKs**: Official `@clickup/rest-client` plus community wrappers provide full API coverage.

### 4.3 Ruby + Thor / CmdKit
* **Strength**: Highest team familiarity; DSL expressive.
* **Risk**: Shipping self‑contained binaries is still experimental; end‑users may need Ruby installed unless using `traveling‑ruby` or AppImage.

### 4.4 Rust + clap
* **Performance**: Best runtime and smallest static binaries; memory‑safe.
* **Complexity**: Steeper learning curve and minimal ClickUp‑specific ecosystem; extra work to design plugin host.

### 4.5 Python + Typer
* **Rapid prototyping**, rich ecosystem for HTTP + config.
* **Distribution pain**: PyInstaller output can exceed 50 MB and hits antivirus false‑positives; Homebrew bottles rare.

## 5. Recommendation
**Primary stack for MVP: Go 1.22 + Cobra/Viper**—mirrors GitHub CLI precedent, provides optimal distribution with Homebrew/npm via `goreleaser`, and leverages existing `go-clickup` SDK. Node + oclif is a strong secondary candidate for rapid plugin development; consider exposing “extension” commands via Node wrappers post‑1.0.

---

*Last updated: 2025‑06‑25*

