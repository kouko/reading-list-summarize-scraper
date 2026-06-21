# Plan: Built-in summary prompt styles (outline default + article opt-in)

Source brief: docs/loom/specs/2026-06-19-summary-prompt-styles.md
Total tasks: 5
Critical-path depth: 3 (≤5)  ← longest: T1→T2→{T3|T4}
Execution order: parallel-where-possible (T3 & T4 are independent leaves after T2)
Plan-document-reviewer verdict: PASS (2026-06-19, 14/14)

## Task 1 — SummaryConfig.Style field
- Description: Add `Style string` (`yaml:"style"`) to `SummaryConfig`. No DefaultConfig change needed (zero value "" resolves to the outline default at the resolver); document the default as "outline" in the field comment.
- Module: internal/config
- Files touched: internal/config/config.go, internal/config/config_test.go
- Context paths:
  - internal/config/config.go  (SummaryConfig ~:208; TestLoad_* parse-test template)
- Acceptance:
  - RED: TestLoad_SummaryStyle — load YAML `summary: { style: article }`; fails to compile (field undefined).
  - GREEN: parsed `cfg.Summary.Style == "article"`; and an omitted style yields `""`.
- Dependencies: none
- Independent: true
- Brief item covered: "config.SummaryConfig.Style string (yaml:\"style\"): outline (default) | article"

## Task 2 — Thread style into prompt resolution
- Description: Change `loadBuiltinPrompt(language string)` → `loadBuiltinPrompt(language, style string)`: `prefix := "summary"; if style == "article" { prefix = "summary-article" }`; call `loadBuiltinPromptByPrefix(prefix, language)`. `ResolvePrompt` passes `summaryConfig.Style`. Update the PR-#11 test `TestLoadBuiltinPrompt_KnownAndFallback` call sites to the 2-arg signature.
- Module: internal/summarize
- Files touched: internal/summarize/prompt.go, internal/summarize/prompt_resolve_test.go
- Context paths:
  - internal/summarize/prompt.go  (ResolvePrompt :27, loadBuiltinPrompt :52)
  - internal/summarize/keywords.go  (loadBuiltinPromptByPrefix :28)
- Acceptance:
  - RED: TestLoadBuiltinPrompt_StyleSelectsPrefix — assert that with a (temp/stub) article file present, `style=="article"` selects the article prefix and any other value (""/"outline"/"xxx") selects the plain summary prefix. Fails to compile (loadBuiltinPrompt is 1-arg).
  - GREEN: 2-arg loadBuiltinPrompt + ResolvePrompt threading compile; new test + updated PR-#11 tests pass; `go test ./internal/summarize/` green.
- Dependencies: Task 1 completes first
- Independent: false
- Brief item covered: "loadBuiltinPrompt(language, style): style==\"article\" → prefix summary-article; ... ResolvePrompt threads summaryConfig.Style"

## Task 3 — Author article-style built-in prompts (en/ja/zh-Hant)
- Description: Create `prompts/builtin/summary-article-{en,ja,zh-Hant}.md`, adapting ytss #58's article DESIGN to rlss's article-content placeholders (NOT copying transcript-tuned content). Each: lead with `### TL;DR` then `### Overview`; `####` sections "narrate first then list"; comparative data → table; drop the output-scale table (use `{{content_tier}}` as a soft hint); rlss placeholders only (`{{title}}`/`{{domain}}`/`{{date_added}}`/`{{content_length}}`/`{{content_tier}}`/`{{content}}`); per-language output (zh-Hant → Traditional Chinese).
- Module: prompts/builtin
- Files touched: prompts/builtin/summary-article-en.md, prompts/builtin/summary-article-ja.md, prompts/builtin/summary-article-zh-Hant.md, internal/summarize/summary_article_test.go
- Context paths:
  - prompts/builtin/summary-en.md  (current outline baseline for tone/placeholders)
- Acceptance:
  - RED: TestArticlePrompts_LoadAndShape (internal/summarize/summary_article_test.go) — for each of en/ja/zh-Hant, `loadBuiltinPrompt(lang, "article")` returns non-empty, contains `{{content}}` and `### TL;DR`, and contains NO ytss-only placeholder (`{{transcript}}`/`{{transcription_tier}}`). Fails: files absent.
  - GREEN: 3 files authored + embedded; test passes. (Wording quality is verified manually via the Task-4 harness, not asserted here.)
- Dependencies: Task 2 completes first
- Independent: true
- Brief item covered: "New built-in summary-article-{en,ja,zh-Hant}.md — rlss-adapted article-style prompts"

## Task 4 — Env-gated live-prompt harness
- Description: Add `internal/summarize/prompt_live_test.go` porting ytss #56: `TestSummaryPromptLive` skips unless `RLSS_LIVE_PROMPT=1`; otherwise loads config (`RLSS_CONFIG`, default `../../config.yaml` relative to package dir — verify path), resolves the prompt (`RLSS_PROMPT_FILE` override else `ResolveAndSubstitute(cfg.Summary, sampleVars)`), builds the provider chain (`NewSummarizer`), runs `Summarize` against a sample article (`RLSS_ARTICLE` override else a built-in sample), and prints the result. Add `prompts/my-summary.md` to .gitignore.
- Module: internal/summarize
- Files touched: internal/summarize/prompt_live_test.go, .gitignore
- Context paths:
  - internal/summarize/prompt.go  (ResolveAndSubstitute, PromptVars)
  - internal/summarize/summarizer.go  (NewSummarizer, Summarizer.Summarize)
- Acceptance:
  - RED: `go test ./internal/summarize/` — the new test file fails to compile until written (undefined sample/symbols).
  - GREEN: with `RLSS_LIVE_PROMPT` unset, `go test ./internal/summarize/ -run TestSummaryPromptLive -v` SKIPS (no LLM call) and the package suite stays green; the harness compiles and goes through ResolvePrompt→SubstituteVars→NewSummarizer→Summarize when opted in.
- Dependencies: Task 2 completes first
- Independent: true
- Brief item covered: "Live-prompt harness (RLSS_LIVE_PROMPT=1, env-gated, NOT in default suite/CI) ... ports ytss #56"

## Task 5 — Document summary.style in config.example.yaml
- Description: Add a commented `style:` key under the `summary:` block in config.example.yaml documenting `outline` (default) | `article`, mirroring the existing summary-block style.
- Module: docs (config.example.yaml)
- Files touched: config.example.yaml
- Context paths:
  - config.example.yaml  (existing summary: block)
- Acceptance:
  - RED: grep for `style:` under summary in config.example.yaml returns nothing.
  - GREEN: `style:` documented (outline default / article opt-in); key name matches Task 1's yaml tag.
- Dependencies: Task 1 completes first  (doc-mirrors-code: key name must match the struct yaml tag — see Notes)
- Independent: false
- Brief item covered: Smallest End State — the user-facing `summary.style` config surface.

## Notes
- Depth-3, 5 tasks. T1 (config) is the root; T2 (resolver) depends on it; T3 (article prompts) and
  T4 (harness) are independent leaves after T2 (disjoint files: prompts/builtin/* + summary_article_test.go
  vs prompt_live_test.go + .gitignore; no shared symbol). T5 (doc) depends on T1 (doc-mirrors-code,
  kept Independent:false though files are disjoint).
- Testability: T1/T2/T3 carry real unit tests (config parse / style-prefix selection / article-file
  shape). T4 is an env-gated dev harness — its automated assertion is "skips by default + compiles +
  package stays green"; actual prompt-wording quality is tuned manually through it (brief Open Question).
- Out of scope (brief): per-channel style, article output-scale table, CI-wiring the harness,
  touching keywords/mermaid/outline prompt content.
