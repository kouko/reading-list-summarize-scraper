# 修正 ITmedia 分頁文章不完整卻回報成功
originator: kouko
kind: product
needs-design: yes — Multi-page extraction changes user-visible saved content and success/failure outcomes; no existing spec covers reliable completion of this flow.
status: confirmed 2026-10-24
publication: automatic — authorized 2026-10-24 by kouko

## Problem
kouko 在實際驗證中發現，一篇五頁的 ITmedia 文章只保存第一頁，卻顯示成功。這會讓後續筆記使用不完整的原文，無法信任擷取結果。

## Proposed outcome
能完整取得已驗證的 ITmedia 分頁文章，並讓擷取結果如實反映文章是否完整。

## Acceptance
1. 已驗證的 ITmedia 五頁文章可取得全部五頁的目標正文，順序正確且不重複合併頁面。
2. 已知文章尚未完整取得時，不回報完整擷取成功，也不把部分正文保存為成功結果。
3. 既有的瀏覽器重試設定維持效果，先前可成功取得的其他新聞分頁案例仍能正確取得正文。
4. 收集更多實際日本新聞網站的分頁文章案例，逐項記錄頁數、正文核對及成功或失敗結果。

## Constraints
- 正確性優先；不得以登入頁、驗證頁或其他文章取代目標正文。
- 保留既有先無視窗嘗試、失敗後有視窗重試的設定與 headed_on_block: false 效果。
- 驗證使用隔離設定、瀏覽器資料與輸出，不修改使用者正式設定、筆記或正在執行的程式。

## Value case
受益者是以本工具保存原文並產生筆記的 kouko。使用者已要求處理實際重現的錯誤；目前替代方式是逐頁手動核對，或暫不信任成功狀態。本輪聚焦此修正，全面自動辨識分頁另案評估。GO — 已有實站不完整卻成功的證據，直接影響原文正確性。

## Out of scope
- 所有網站的自動分頁辨識。
- 其他網站的專屬新擷取功能。
- 既有筆記的批次補抓、改寫或遷移。

## Open questions
- none
