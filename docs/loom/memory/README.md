---
name: README
description: Repo-native OKF v0.2-compatible Loom memory store — one durable repository lesson per Markdown file, index.md generated and never hand-edited, Git the sole historical log.
type: Memory Store Guide
sources:
  - resource: "docs/loom/intent/2026-10-31-complete-facebook-and-dated-articles.md"
---

# Loom memory store

> Repo-native home for distilled durable practices, habits, processes, and
> recurring gotchas — knowledge that outlives a single commit or change and
> must travel with the repository (any machine, any host, headless agents).
> One lesson per file. `index.md` is generated, never hand-edited — run
> `loom-memory`'s validator to regenerate it after any lesson is added,
> reconciled, or retired. Stale lessons are deleted, not archived — Git
> history is the archive.

## Format

Every non-reserved file in this directory is a Loom memory concept: YAML
frontmatter with `name` (must equal the filename stem), `description` (a
single durable relevance-rule line), `type`, and at least one
`sources[].resource` entry, followed by the lesson body.

## Regenerating the index

Run the `skills/loom-memory/scripts/loom_memory.py` file inside the
`loom-workflow` plugin's own directory
(`${CLAUDE_PLUGIN_ROOT}/skills/loom-memory/scripts/loom_memory.py` on Claude
Code, which substitutes that token; on any other host replace
`${CLAUDE_PLUGIN_ROOT}` with the installed `loom-workflow` plugin directory,
the one holding `skills/loom-memory/SKILL.md`):

```
python3 ${CLAUDE_PLUGIN_ROOT}/skills/loom-memory/scripts/loom_memory.py regenerate-index <this-directory>
python3 ${CLAUDE_PLUGIN_ROOT}/skills/loom-memory/scripts/loom_memory.py validate <this-directory>
```

`validate` fails loudly — naming every offender — on malformed frontmatter,
a missing required field, duplicate concept identity, or index drift. It
never edits the store; `regenerate-index` writes only `index.md`, and
refuses to write anything at all when the concept metadata behind it is
itself broken.
