# 依加入時間排序閱讀清單
originator: kouko
kind: product
needs-design: yes — 清單顯示與文章處理順序將改變，尚無涵蓋日期排序的規格
status: confirmed 2026-10-23
publication: automatic — authorized 2026-10-23 by kouko

## Problem
kouko 無法依閱讀清單的加入時間查看與處理文章；合併來源後的順序不代表加入時間，限制每輪筆數時也無法確認選到的是最近加入的項目。

## Proposed outcome
程式提供依加入時間排列的閱讀清單與處理順序，讓 kouko 可以優先查看及處理最近加入的文章。

## Acceptance
1. 清單顯示與批次處理依加入時間由新到舊排列。
2. 同時讀取多個來源時，結果使用一致的加入時間順序。
3. 設定每輪筆數上限時，選取排序後符合篩選條件的項目。

## Constraints
- 沿用目前專案與設定檔。
- 正確性優先。

## Value case
受益者為 kouko。使用者在確認目前來源合併與每輪上限後，要求把日期排序加入程式，而非僅整理一次性的輸出清單。GO：排序可直接解決使用者無法掌握處理順序的現有問題。

## Out of scope
- 網頁擷取及 AI 摘要內容的變更。
- 已完成項目的跳過規則變更。

## Open questions
- none
