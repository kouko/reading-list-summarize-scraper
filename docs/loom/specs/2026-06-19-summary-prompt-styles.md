# Built-in summary prompt styles (outline default + article opt-in)

Feasibility brief — 2026-06-19. Status: discovery done (via prior assessment). Scope: FULL
(mechanism + 3 languages + live-prompt harness). Ports youtube-summarize-scraper #58 (+#56).

## Problem
(Axis 1 — JTBD) rlss has exactly one built-in summary shape (a structured "outline": Overview +
`####` sections + key-point lists, gated by an output-scale table). Some articles read better as a
short narrative + tables than as a bullet outline. The job: *let me choose, per config, between the
current outline summary and a more article-style summary (TL;DR + prose + tables) — without
hand-writing a custom prompt file each time.* Mirrors ytss #58.

## Users
(Axis 2) kouko, summarizing reading-list articles → Obsidian. Job story: *When an article is more
narrative than list-shaped, I want to switch the built-in summary to an "article" style globally, so
the output reads like a written piece, while keeping outline as the default for everything else.*
Also: a dev loop to actually tune the 3 prompts against real articles before shipping.

## Smallest End State
(Axis 3 — FULL scope chosen by user)
1. `config.SummaryConfig.Style string` (`yaml:"style"`): `"outline"` (default) | `"article"`.
   Default outline = current behavior unchanged (backward-compatible opt-in).
2. `loadBuiltinPrompt(language, style)`: `style=="article"` → prefix `summary-article`; anything
   else (incl. `""`/unknown) → `summary` (outline). `ResolvePrompt` threads `summaryConfig.Style`.
3. New built-in `summary-article-{en,ja,zh-Hant}.md` — rlss-adapted article-style prompts (NOT
   copied from ytss; ytss's are transcript/subtitle-tuned). Use rlss placeholders
   (`{{title}}`/`{{domain}}`/`{{date_added}}`/`{{content_length}}`/`{{content_tier}}`/`{{content}}`).
   Design (from ytss #58, adapted): lead with `### TL;DR` (heading-safe vs the preamble stripper),
   then `### Overview`; each `####` section "narrate first (1–3 sentences prose) then a list";
   comparative data → Markdown table; drop the rigid output-scale table (keep `{{content_tier}}` as a
   soft hint); preserve concrete detail; faithful (no speculation); per-language output (zh-Hant in
   Traditional Chinese, etc.).
4. **Live-prompt harness** (`RLSS_LIVE_PROMPT=1`, env-gated, NOT in default suite/CI): renders the
   resolved prompt against a sample article through the real chain
   (ResolvePrompt→SubstituteVars→NewSummarizer→Summarize) and prints it, for tuning the 3 prompts.
   Ports ytss #56.

## Current State Evidence
- **Forward** (prompt resolution): `internal/summarize/prompt.go:27` `ResolvePrompt(summaryConfig)`
  3-level cascade (file > inline > `loadBuiltinPrompt(language)` :39) → `keywords.go:28`
  `loadBuiltinPromptByPrefix("summary", lang)` reads `prompts/builtin/summary-<lang>.md` (embedded),
  en-fallback within the prefix. `runner.go` calls `ResolveAndSubstitute(p.config.Summary, vars)`.
- **Reverse** (config SSOT / builtin FS): `internal/config/config.go` `SummaryConfig` (`:208`);
  `prompts/builtin/embed.go` `//go:embed *.md` → `builtin.FS`; files: summary/keywords/mermaid ×
  en/ja/zh-Hant. rlss has NO DefaultConfig Summary block (zero-value Language ""/Style "" → en
  fallback + outline default both hold without a DefaultConfig change).
- **Error**: `loadBuiltinPromptByPrefix` falls back to `<prefix>-en.md`, else errors. So shipping
  `style:article` REQUIRES the `summary-article-*.md` files to exist (selector + content land together).
- **Data**: rlss placeholders are content-oriented (`{{content}}`, `{{content_tier}}`), NOT ytss's
  `{{transcript}}`/`{{transcription_tier}}`; substitution in `prompt.go:58` SubstituteVars.
- **Boundary**: rlss `ResolvePrompt` takes only `SummaryConfig` (no per-channel arg) — 3-level, vs
  ytss's 4-level. Threading `Style` is a one-arg change. PR #11 just added prompt_resolve_test.go
  whose `loadBuiltinPrompt("ja")` calls must update to the new 2-arg signature.

## Decision
Port ytss #58's style-selector mechanism (config `Style` + `loadBuiltinPrompt(lang, style)` +
resolver threading), author rlss-appropriate `summary-article-{en,ja,zh-Hant}.md` (adapt ytss's
article DESIGN to rlss's article-content placeholders — do not copy transcript-tuned content), and
port ytss #56's env-gated live-prompt harness to tune them. Default stays `outline` (current output
unchanged). Do NOT add per-channel style, a new output-scale table for article, or wire the harness
into CI.

## Alternatives Considered
(Axis 4 — assessed in the prior turn)
1. Mechanism only (no article prompts) — REJECTED: ships a selector that errors when set to article
   (files absent) / no-ops otherwise. Zero standalone value (YAGNI scaffold).
2. Mechanism + en only + manual verify — viable minimal, but user chose full 3-language + harness.
3. Full (chosen): mechanism + 3 langs + harness — matches ytss #58+#56; harness is what makes the
   prompt quality tunable rather than blind.

## What Becomes Obsolete
(Axis 5) Nothing removed — purely additive opt-in. The current `summary-<lang>.md` stays as the
outline default. Additive is justified (new user-facing option), bounded by "default unchanged".

## Out of Scope
- Per-channel / per-item style (rlss has no channel concept).
- Output-scale table for the article style (ytss dropped it; article uses tier as a soft hint).
- Wiring the live harness into CI / the default `go test` suite (it makes real LLM calls).
- Changing keywords/mermaid prompts or the outline prompt content.
- A `style` value beyond outline/article.

## Open Questions
- Article prompt wording is iterative/subjective; "done" = renders coherently via the harness on a
  few real articles per language (en/ja/zh-Hant), faithful + correctly-styled. No numeric bar.
