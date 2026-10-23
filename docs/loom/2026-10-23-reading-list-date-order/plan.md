# Reading list date order — plan
intent: 2026-10-23-reading-list-date-order@232216c
spec: docs/loom/2026-10-23-reading-list-date-order/spec.md@048a0c4
charter: 1.1

## Task DAG
### Wave 0 — shared collection order
**W0-01 Order qualifying reading items before limiting** after: none acceptance: 1, 2, 3
- Files: cmd/rlss/process.go (entry: applyFilters), cmd/rlss/process_test.go, README.md
- Test: A1 positive: precise-newest-first; boundary: stable-ties-undated. A2 positive: mixed-sources; boundary: first-url-occurrence. A3 positive: filtered-top-n; boundary: unlimited-empty.
- Risk: agent-decided — REQ-1/REQ-3 use a stable standard-library timestamp comparison in the existing shared filtering entry before limit; preserve deduplication and skip handling. Explain unchanged completed-item behavior in existing documentation.

## Simplicity check
- none found

## Questions asked
① — what — 確認這樣符合你的需求嗎？確認後會完成測試與審查，再依授權建立 PR；合併另行確認。
② — behaviour — 依 Loom 行為確認流程，請確認以上行為是否符合預期？

## Risks
1. Unknown dates sort after actual current reading-list dates; missing-date tests protect this boundary despite both audited live sources currently having complete timestamps.
2. Same-instant records keep source collection order. Existing duplicate URL selection and completed-item skipping remain unchanged, so sorting alone does not backfill older unprocessed items beyond the limit.
3. The running user process remains untouched. Verification uses isolated output/configuration; deploying requires rebuilding and restarting the executable separately.
