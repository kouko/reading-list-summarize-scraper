# Recover Felix and X article content — plan
intent: 2026-10-27-recover-felix-and-x-article-content@02b94b0
spec: docs/loom/2026-10-27-recover-felix-and-x-article-content/spec.md@f169cbf
charter: 1.1

## Task DAG
### Wave 1 — Restore public X target bodies
**W1-01 Recognize complete target posts and articles**  after: none  acceptance: 1, 3, 4
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs
- Test: A1 positive: x-public-target-body; negative: x-wrong-target. A3 positive: x-complete-only; negative: x-preview-or-reply. A4 positive: legacy-social-bodies; boundary: challenge-policy.
- Risk: agent-decided — REQ-1/3/4 Design decision: preserve legacy routes while adding observed target-scoped layouts; extraction.test.mjs and extraction-boundaries.test.mjs coverage widens without dropping existing cases. Inspect all eight actual samples before declaring support.

### Wave 2 — Prevent ordinary article metadata from deleting the body
**W2-01 Preserve substantive article containers during cleanup**  after: W1-01  acceptance: 2, 3, 4
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs
- Test: A2 positive: ordinary-dated-body; negative: ordinary-date-only. A3 positive: ordinary-complete-body; negative: ordinary-empty-or-challenge. A4 positive: ordinary-existing-cleanup; boundary: original-dom-preserved.
- Risk: agent-decided — REQ-2/3/4 Design decision: narrow cleanup adaptation rather than disabling parser filters; extraction.test.mjs coverage widens, existing extraction-boundaries.test.mjs cases remain intact. Avoid domain-specific bypasses and preserve source DOM.

### Wave 3 — Integrate the embedded browser entry
**W3-01 Regenerate the committed browser bundle**  after: W2-01  acceptance: 1, 2, 3, 4
- Files: embed/defuddle.min.js (entry: window.extractArticle)
- Test: A1 positive: bundled-x-target; negative: bundled-wrong-target. A2 positive: bundled-dated-body; negative: bundled-date-only. A3 positive: bundled-complete-body; negative: bundled-empty. A4 positive: clean-rebuild; boundary: legacy-bundle-path.
- Risk: agent-decided — REQ-1/2/3/4 require the shipped Go embedded entry, not only TypeScript source. Regenerate using the existing build script; verify deterministic rebuild and focused tests. No handwritten bundle changes.

## Simplicity check
- none found

## Questions asked
① — what — Original confirmation question unavailable in retained context; user confirmation: 「ＯＫ 開始實作吧」. No reconstructed quotation.
② — behaviour — 以上行為符合你的期待嗎？

## Risks
1. X can serve different layouts or deny a fresh browser. All eight exact URLs require independent live body and section comparisons; inaccessible or partial samples are not completed acceptance evidence.
2. Source body availability is not proven by canonical identity alone. Adversarial tests must cover wrong identity, nested quoted or reply bodies, previews, ambiguity and login or verification shells.
3. Preserve production configuration, browser data, running executable and notes. All browser captures and executable outputs use isolated temporary paths; committed evidence excludes full copyrighted article captures.
4. agent-decided — No paid API, new account requirement, dependency upgrade, file-format migration or bulk rewrite; preserve existing retry policy and settings. The current scope requires no new obligation or irreversible data operation.
5. Live completeness evidence records source URLs, target identity, paragraph and section coverage, and contamination checks per sample; full package and adversarial checks run after both implementation tasks.
