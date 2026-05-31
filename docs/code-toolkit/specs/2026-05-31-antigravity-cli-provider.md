# Brief: add `antigravity-cli` summarization provider

Date: 2026-05-31
Status: ready for writing-plans

## Problem

Google sunsets **Gemini CLI** for individual Google AI Pro/Ultra & free users on **2026-06-18**, replacing it with **Antigravity CLI** (`agy`). rlss's `gemini-cli` provider (`internal/summarize/gemini.go`) invokes the `gemini` binary non-interactively, so it stops working for those users. The job: keep a working "Google-family CLI" summarization option after the sunset by adding support for `agy`.

## Users

rlss users (the author + Homebrew tap users) who run the `gemini-cli` provider authenticated with an individual Google AI account. After 2026-06-18 they must either switch to a paid API key, another provider (`claude-code` etc.), or this new `antigravity-cli` provider.

## Smallest End State

A new `antigravity-cli` provider mirroring the existing CLI-provider pattern (`qwen_code.go` / `claude_code.go`):
- Invoke `agy -p "<combinedPrompt>" --print-timeout <N>m` with **content/prompt via stdin** (stdin is read as context — verified).
- Run agy with `cmd.Dir` set to a per-call **temp directory** (avoids littering the user's cwd with `.antigravitycli/`).
- **Post-process**: strip occasional `（來源：…）` source-injection tail-notes (agy agent-first web-grounding leak, ~1/5 runs) — **inside this provider only**.
- Reuse shared `StripThinkingTags` + `isQuotaMessage`/`QuotaError` so it plugs into the existing fallback chain + circuit breaker unchanged.

## Locked decisions (user sign-off 2026-05-31)

1. **Provider name / yaml key**: `antigravity-cli` (parallels `gemini-cli`; direct successor to Gemini CLI).
2. **Model**: **omit** the `model` field entirely from the config struct — agy headless v1.0.3 has no `-m`/`--model` flag. `SummarizeResult.Model` reported as empty.
3. **Side-effect isolation**: run agy in a per-call temp dir; **no** `--sandbox`. PoC showed pure summarization writes no files; auth uses global keychain so temp cwd is safe.
4. **Source-injection strip**: provider-specific post-process (not shared `StripThinkingTags`) — low blast radius.

## Current State Evidence

- **Forward** (interface providers implement): `internal/summarize/summarizer.go:29` `Summarizer.Summarize`.
- **Reverse** (factory that builds providers from config — the integration point): `internal/summarize/summarizer.go:86-152` `newSingleProvider`, `switch name`.
- **Pattern to mirror**: `internal/summarize/qwen_code.go:20-77`, `internal/summarize/claude_code.go:20-85` (subprocess + stdin pipe + quota detection + `StripThinkingTags`).
- **Config touch points**: `internal/config/config.go:117-127` `LLMConfig`; `:151-161` `GeminiCLIConfig`/`QwenCodeConfig` struct shape; `:314-316` `ExpandPath` for `.Path`.
- **Error/quota plumbing (reuse, do not duplicate)**: `internal/summarize/errors.go:12-51` `QuotaError` + `isQuotaMessage`.
- **Boundary** (PoC verified): agy works via Go `exec.Command` + stdin pipe, no TTY (16s, exit 0). `agy -p ""` errors `flag needs an argument` → prompt must be the `-p` value. No `-m`/`-o`/`--exclude-tools`/`--strict-mcp-config` flags in v1.0.3.
- **Evidence paths**: `internal/summarize/{summarizer,qwen_code,claude_code,gemini,errors}.go`, `internal/config/config.go`.

## What Becomes Obsolete

Nothing removed — purely additive. `gemini-cli` provider **stays** (paid API keys + enterprise licenses keep working post-sunset). README note added pointing individual-account users at this new provider.

## Out of Scope

- Removing or deprecating the `gemini-cli` provider.
- Supporting agy interactive features (subagents, plugins, `/model`, hooks).
- Model selection for agy (not possible in headless v1.0.3).
- Auto-migrating user config from `gemini-cli` to `antigravity-cli`.
- `--sandbox` hardening (deferred; revisit if file-write side effects appear).

## Open Questions

- `SummarizeResult.Model` value when model is unknown: empty string vs a `"default"` sentinel (affects `llm_model` frontmatter). Lean: empty string. Implementer's call during TDD.
- Quota-pattern coverage for agy's specific rate-limit message wording — verify agy's quota error text matches existing `quotaPatterns` (`errors.go:31-39`); extend only if a real mismatch shows up.
