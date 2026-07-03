# openai-compat multi-instance HA

Porting ytss commit `ebc1d85` (PR #61, 2026-07-03) into rlss.

## Problem

Every other rlss provider family already gained circuit-breaker/failover
resilience in a prior porting session (`ProviderFallbackStrategy`,
`CircuitBreaker`, empty-response threshold — all mirrored from ytss). But
`openai-compat` is still a single endpoint: if the one configured box goes
down, there is no in-family failover, only whatever cross-provider fallback
chain the user configured (if any). The job: let a user who runs multiple
OpenAI-compatible boxes (e.g. two LM Studio instances on the LAN) treat them
as a named HA pool within the `openai-compat` provider family, the same way
ytss's users already can.

## Users

Anyone running rlss with a local/self-hosted OpenAI-compatible inference
server (oMLX, LM Studio, vLLM) who wants a second box as a hot spare —
currently must either accept single-point-of-failure or configure an
unrelated provider (e.g. `ollama`) as fallback, which doesn't help when the
whole point is redundant *compatible* boxes.

## Smallest End State

Mirror ytss's shape exactly — it's a solved, precedent-approved problem in
the sibling repo with an identical config/summarizer architecture:

- `LLMConfig.OpenAICompat` becomes `map[string]OpenAICompatConfig` (was a
  single `OpenAICompatConfig`).
- `newSingleProvider` resolves `"openai-compat"` → instance `"default"`, and
  `"openai-compat:<name>"` → instance `<name>`, via a plain map lookup that
  errors loudly (`fmt.Errorf("openai-compat: no instance %q configured", ...)`)
  when the instance is missing. No new struct, no new validation layer —
  a missing map key is already the natural "instance not configured" signal
  (covers the empty-instance-name case too: `"openai-compat:"` → instance
  `""` → not in map → same error, no special-casing needed).
- `DefaultConfig()` seeds **no** default instance (`nil` map) — matching
  ytss's explicit choice: bare `openai-compat` with no configured `default`
  key fails loud rather than silently binding `localhost:8000`.
- `config.example.yaml` migrates the single block to a `default:` instance
  plus commented `box1`/`box2` HA examples, with the naming rules ytss
  documents (names may not contain `:`, `default` is reserved for the bare
  form).

## Current State Evidence

- **Forward** (config → summarizer): `config.go:138` declares
  `OpenAICompat OpenAICompatConfig` on `LLMConfig`; `summarizer.go:169-179`
  consumes it directly inside the `switch name` in `newSingleProvider`
  (`case "openai-compat":`).
- **Reverse** (who else reads `cfg.OpenAICompat`): only `summarizer.go` reads
  the field (`grep -rn OpenAICompat` shows no other call sites besides
  `config.go`/`defaults.go`/`summarizer.go`/`openai_compat.go`, and
  `openai_compat.go` only defines the `OpenAICompatSummarizer` struct/method,
  it doesn't touch `config.OpenAICompat`).
- **Error**: today, a bad/empty `openai-compat` config doesn't error at
  resolution time at all — it just builds an `OpenAICompatSummarizer` with
  zero-value fields (empty endpoint), which will fail later at request time
  with an HTTP-level error. Post-port, an unconfigured/misnamed instance
  fails immediately in `newSingleProvider` with a named error, matching
  every other provider's fail-loud-on-missing-config convention (e.g.
  `default: return nil, fmt.Errorf("unknown LLM provider: %q", name)`).
- **Data**: `defaults.go:36-38` seeds `OpenAICompat: OpenAICompatConfig{Timeout: 900}`
  — a single-instance default. Post-port this becomes an intentional `nil`
  map (no seeded instance), matching ytss's precedent and its documented
  rationale.
- **Boundary**: `config.example.yaml:104-108` is the only doc surface for
  this provider; `docs/loom/plans/...` and `docs/loom/specs/...` mention
  `openai-compat` only in passing (original design doc), nothing else
  encodes the single-struct shape as an invariant.

Evidence paths: `internal/config/config.go:128-187`,
`internal/config/defaults.go:1-93`, `internal/summarize/summarizer.go:1-183`,
`config.example.yaml:34-109`.

## Decision

Port ytss's map-based instance resolution verbatim into rlss:

1. `internal/config/config.go`: change `OpenAICompat` field type to
   `map[string]OpenAICompatConfig`; `OpenAICompatConfig` struct itself is
   unchanged.
2. `internal/config/defaults.go`: remove the seeded single-instance default;
   leave `OpenAICompat` as a nil map (matches ytss; a Go nil map reads as
   empty, no explicit `nil` assignment needed — just delete the field
   entry).
3. `internal/summarize/summarizer.go`: replace the `case "openai-compat":`
   branch with the family-prefix check ytss uses (`name ==
   "openai-compat" || strings.HasPrefix(name, "openai-compat:")`), resolve
   the instance name (default `"default"`, or text after `:`), look it up in
   the map, error loudly on miss.
4. `config.example.yaml`: migrate the single block to `default:` plus two
   commented HA box examples and the naming-rule comments, matching ytss's
   example.

This is a deliberate **backward-incompatible break**, same as ytss's own
approved precedent: an existing single-block `openai-compat:` YAML config
(`endpoint:`/`model:`/`api_key:`/`timeout:` at the top level) will fail to
parse into the new map shape — those fields need to be re-nested one level
under a `default:` key. Since this is a solo-maintained repo (not yet
released to other users per the sibling precedent), the break is low-risk;
still called out explicitly so it's not silently discovered on next run.

## Out of Scope

- No changes to `OpenAICompatSummarizer` itself (`openai_compat.go`) — the
  HTTP call logic is untouched, only how it's *constructed* changes.
- No cross-provider HA (i.e. this doesn't change how `openai-compat`
  interacts with the existing `ProviderFallbackStrategy` chain across
  *different* provider families) — that mechanism already exists and
  composes for free once `"openai-compat:box1"`, `"openai-compat:box2"` are
  valid provider-list entries.
- No migration tool/script for existing configs — this repo has no external
  users yet (single-maintainer), so a doc note in the example config is
  sufficient.

## What Becomes Obsolete

- The single-instance `OpenAICompatConfig` field-read path in
  `summarizer.go:169-179` is fully replaced (not left as dead code).
- The single-block example in `config.example.yaml:104-108` is replaced by
  the map form — no dual-format support kept.

## Alternatives Considered

Axis-4 industry research was skipped: this isn't an open design question —
it's porting an already-shipped, already-approved solution from a sibling
repo with an identical architecture, written by the same maintainer, for the
same problem. Re-litigating alternatives (e.g. a list-of-named-structs
instead of a map, or a separate "HA pool" config block) would produce
inconsistency between the two repos for no benefit; the map-keyed-by-name
shape is already validated in production-adjacent use in ytss. The only
real alternative considered was *not* porting default-seeding behavior
change (i.e., keep seeding a `"default"` instance in rlss's `DefaultConfig`
for softer backward compat) — rejected in favor of matching ytss exactly,
since ytss's own commit message explicitly chose fail-loud over a silent
`localhost:8000` default, and that reasoning applies identically to rlss.
