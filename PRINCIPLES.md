# Product principles

ratified-by: kouko 2026-10-22

## Who
kouko：從閱讀清單取得文章正文並產生筆記；目前部分頁面需要手動切換瀏覽器模式重試。

## Non-negotiables (ordered)
1. Prioritize correctness: delivered content must correspond to the target article body.
2. Never treat a login or verification page as the actual article body.
3. Never report extraction success when the actual article body was not obtained.

## Won't do
- 不將錯誤頁面摘要成文章筆記。

## Failure we must avoid
- 無法取得目標頁面正文。
- 取得錯誤頁面內容，例如登入頁面，而非實際內文。

## Fixed choices
- 沿用現有閱讀清單工具。
- 所有頁面先用 headless 嘗試抓取，失敗後改用 headed。
