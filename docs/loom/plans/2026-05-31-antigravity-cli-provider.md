# Plan: antigravity-cli provider

Source brief: docs/loom/specs/2026-05-31-antigravity-cli-provider.md
Total tasks: 6
Critical-path depth: 4 (≤5)   ← longest chain: Task1 → Task3 → Task4 → Task6
Execution order: parallel-where-possible
Plan-document-reviewer verdict: PASS (2026-05-31, 14/14 checks)

## Task 1 — source-injection strip helper
- Description: Create `internal/summarize/antigravity.go` with a pure function `stripSourceInjection(s string) string` that removes `（來源：…）` tail-notes (full-width parens, "來源" prefix, non-greedy to closing `）`) and trims surrounding whitespace, leaving all other text intact.
- Module: internal/summarize
- Files touched: internal/summarize/antigravity.go, internal/summarize/antigravity_test.go
- Context paths:
  - internal/summarize/summarizer.go (StripThinkingTags style/regex idiom to mirror)
- Acceptance:
  - RED: `TestStripSourceInjection` in antigravity_test.go — asserts a string ending `（來源：[ai-souken.com]、[antigravity.google]）` has the note removed; a string containing legitimate inline `（來源：原文）` mid-sentence and a string with no note are handled per spec. Fails to compile (function undefined).
  - GREEN: `go test ./internal/summarize -run TestStripSourceInjection` passes.
- Dependencies: none
- Independent: true
- Brief item covered: "Post-process: strip occasional `（來源：…）` source-injection tail-notes … inside this provider only."

## Task 2 — AntigravityCLIConfig struct + config wiring
- Description: Add `AntigravityCLIConfig{ Path string `yaml:"path"`; Timeout int `yaml:"timeout"` }` (NO Model field) to `internal/config/config.go`; add `AntigravityCLI AntigravityCLIConfig `yaml:"antigravity-cli"`` to `LLMConfig`; add `cfg.LLM.AntigravityCLI.Path = ExpandPath(...)` alongside the existing ExpandPath calls.
- Module: internal/config
- Files touched: internal/config/config.go, internal/config/config_test.go
- Context paths:
  - internal/config/config.go (LLMConfig:117-127, GeminiCLIConfig:151-161, ExpandPath block:314-316)
- Acceptance:
  - RED: `TestLoad_AntigravityCLIConfig` in config_test.go — loads YAML with an `antigravity-cli: { path: "~/bin/agy", timeout: 600 }` block under `llm:`, asserts `cfg.LLM.AntigravityCLI.Timeout == 600` and `Path` is tilde-expanded (absolute). Fails (field does not exist).
  - GREEN: `go test ./internal/config -run TestLoad_AntigravityCLIConfig` passes.
- Dependencies: none
- Independent: true
- Brief item covered: Locked decision #1 (yaml key `antigravity-cli`) + #2 (omit model field).

## Task 3 — AntigravityCLISummarizer type + Summarize method
- Description: In `internal/summarize/antigravity.go` add `AntigravityCLISummarizer{ binaryPath string; timeout time.Duration }` and a `Summarize` method implementing `Summarizer`: resolve `agy` via `binaryPath`/`exec.LookPath`; build args `["-p", combinedPrompt, "--print-timeout", "<timeout-minutes>m"]`; set `cmd.Dir` to a per-call `os.MkdirTemp` dir (defer RemoveAll) to isolate `.antigravitycli/`; pipe `combinedPrompt` via stdin; on error wrap + `isQuotaMessage`→`QuotaError{Provider:"antigravity-cli"}`; on success return `StripThinkingTags` then `stripSourceInjection` of stdout, `Provider:"antigravity-cli"`, `Model:""`.
- Module: internal/summarize
- Files touched: internal/summarize/antigravity.go, internal/summarize/antigravity_test.go
- Context paths:
  - internal/summarize/qwen_code.go (subprocess + stdin + quota pattern to mirror)
  - internal/summarize/claude_code.go (LookPath + timeout-default idiom)
  - internal/summarize/errors.go (QuotaError, isQuotaMessage)
- Acceptance:
  - RED: `TestAntigravityCLI_InterfaceConformance` — `var _ Summarizer = (*AntigravityCLISummarizer)(nil)`. Fails to compile until the type + method exist.
  - GREEN: `go test ./internal/summarize -run TestAntigravityCLI` passes (compiles + conformance holds).
- External surfaces: `os/exec` (spawns `agy` binary), `os.MkdirTemp` (temp dir lifecycle); CLI-flag surface `agy -p <prompt> --print-timeout <N>m` — grounding: brief §Boundary PoC (2026-05-31, agy v1.0.3).
- Dependencies: Task 1 completes first
- Independent: false
- Brief item covered: "Invoke `agy -p … --print-timeout …` with content via stdin … run agy with cmd.Dir set to a per-call temp directory … reuse shared StripThinkingTags + isQuotaMessage/QuotaError."

## Task 4 — wire antigravity-cli into provider factory
- Description: Add `case "antigravity-cli":` to `newSingleProvider` in `internal/summarize/summarizer.go`, constructing `&AntigravityCLISummarizer{ binaryPath: cfg.AntigravityCLI.Path, timeout: time.Duration(cfg.AntigravityCLI.Timeout)*time.Second }` with the same 15-min default-timeout fallback as the other CLI providers.
- Module: internal/summarize
- Files touched: internal/summarize/summarizer.go, internal/summarize/summarizer_test.go
- Context paths:
  - internal/summarize/summarizer.go (newSingleProvider switch:86-152)
  - internal/summarize/summarizer_test.go (TestNewSingleProvider_AllProviders providers slice:42)
- Acceptance:
  - RED: add `"antigravity-cli"` to the `providers` slice in `TestNewSingleProvider_AllProviders` — test fails with "unknown LLM provider" until the case is added.
  - GREEN: `go test ./internal/summarize -run TestNewSingleProvider` passes.
- Dependencies: Tasks 2, 3 complete first
- Independent: false
- Brief item covered: Smallest End State "a new `antigravity-cli` provider mirroring the existing CLI-provider pattern" — factory integration point.

## Task 5 — document antigravity-cli in config.example.yaml
- Description: Add a commented `antigravity-cli:` block under `llm:` in `config.example.yaml` mirroring the `gemini-cli` block (path, timeout) with a note that `model` is NOT supported (agy headless has no model flag) and that it requires the `agy` binary.
- Module: docs/config
- Files touched: config.example.yaml
- Context paths:
  - config.example.yaml (existing gemini-cli / qwen-code blocks)
  - internal/config/config.go (AntigravityCLIConfig field names — must match)
- Acceptance:
  - RED diagnostic: `grep -q "antigravity-cli" config.example.yaml` returns non-zero (block absent).
  - GREEN: block present; `go build ./...` unaffected; field names match the struct from Task 2.
- Dependencies: Task 2 completes first
- Independent: true
- Brief item covered: Users section — migration path; surfacing the new provider as a config option.

## Task 6 — README note for gemini-cli sunset → antigravity-cli
- Description: In `README.md`, add a short note (near the LLM provider list / requirements) that Gemini CLI sunsets for individual accounts on 2026-06-18 and that `antigravity-cli` (`agy`) is the successor provider; list `agy` alongside `claude`/`gemini`/`qwen` in Requirements.
- Module: docs
- Files touched: README.md
- Context paths:
  - README.md (Requirements:20-26, provider mentions)
- Acceptance:
  - RED diagnostic: `grep -q "antigravity-cli" README.md` returns non-zero.
  - GREEN: note present and accurate (date 2026-06-18, binary `agy`).
- Dependencies: Task 4 completes first
- Independent: false
- Brief item covered: What Becomes Obsolete — "README note added pointing individual-account users at this new provider."

## Notes

- Tasks 1 and 2 are level-1 parallel leaves (disjoint files, no semantic dependency). Tasks 3 and 5 can also run in parallel (3 dep 1, 5 dep 2; disjoint files antigravity.go vs config.example.yaml).
- gemini/qwen/claude_code providers have no per-provider unit tests (they shell out); Task 3 follows that convention — exec behavior is covered by the manual PoC + P0 spike (brief §Boundary), unit coverage is the strip helper (Task 1) + factory conformance (Tasks 3–4).
- `SummarizeResult.Model` reported as empty string per brief Open Question (model unknowable in agy headless).
- Post-PASS amendment (2026-05-31): stamped verdict PASS; added CLI-flag External-surfaces bullet to Task 3. Both additive + schema-safe (no field removed, DAG unchanged) — reviewer re-run skipped per writing-plans §Amending a PASS plan.
