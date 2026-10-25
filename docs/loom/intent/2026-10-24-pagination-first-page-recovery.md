# 從非首頁網址取得完整分頁文章
originator: kouko
kind: product
needs-design: yes — Recovering earlier article pages changes visible completion, navigation and saved content; the existing spec explicitly excludes automatic first-page recovery.
status: confirmed 2026-10-24
publication: automatic — authorized 2026-10-24 by kouko

## Problem
kouko 提供文章中間頁或最後一頁的網址時，目前仍無法取得完整文章，必須改提供第一頁網址。只拒絕保存不完整內容，尚未滿足直接保存這篇完整文章的需求。

## Proposed outcome
對已設定分頁讀取的網站，提供非首頁網址也能取得同一篇文章的完整正文。

## Acceptance
1. 已驗證的 ITmedia、PRESIDENT、Yahoo 與東洋經濟非首頁網址，均能取得與首頁起始相同的完整目標正文。
2. 保存的正文從第一頁開始，順序正確，每頁不重複合併。
3. 無法確認同一篇文章或未能取得完整正文時，不回報完整成功，也不保存部分正文或登入／驗證畫面作為文章。
4. 首頁起始、既有瀏覽器重試設定及失敗時保留舊檔的效果維持正常。

## Constraints
- 正確性優先，目標必須是同一篇文章；不得把其他文章或登入／驗證畫面當成正文。
- 保留既有先無視窗嘗試、失敗後有視窗重試及 headed_on_block: false 的效果。
- 驗證使用隔離設定、瀏覽器資料與輸出，不修改正式設定、既有筆記或正在執行的程式。

## Value case
受益者是從閱讀清單保存文章的 kouko。使用者已明確要求補上非首頁網址的完整擷取能力；目前替代方式是提供第一頁網址。本輪取代這個額外步驟，不擴充其他擷取功能。GO — 已有六個實際非首頁網址可驗證結果，需求具體且直接補足目前能力。

## Out of scope
- 所有網站的自動分頁辨識。
- 登入、付費或驗證限制的繞過。
- 既有筆記的批次補抓、改寫或遷移。

## Open questions
- none
