# 修復九個失敗樣本的完整正文擷取
originator: kouko
kind: product
needs-design: yes — This changes saved article content and successful extraction for previously failing ordinary articles and X content; the nine-sample completion and rejection behavior needs a specification.
status: confirmed 2026-10-27
publication: automatic — authorized 2026-10-27 by kouko

## Problem
kouko 的閱讀清單目前有八篇 X 內容與一篇 Felix 文章持續擷取失敗，反覆重試後仍無法產生所需的文章筆記。只確認程式持續運作，不能確認這些文章已被完整取得。

## Proposed outcome
這九個失敗樣本都能取得正確、完整的目標正文，並有逐篇核對原始內容的驗證結果。

## Acceptance
1. 八篇已列出的 X 樣本均取得同一目標貼文或長文的完整正文與原有章節，不混入其他貼文、回覆或登入介面。
2. 已列出的 Felix 文章取得完整正文與原有章節，不因頁面附帶資訊而遺失正文。
3. 無法確認目標身分或完整內容時仍回報失敗，不將部分正文、預覽、錯誤頁或登入／驗證畫面保存為成功文章；失敗時保留舊正文與摘要。
4. 既有可成功取得的一般文章與 X 內容，以及原有瀏覽器重試與設定效果，均維持正常。

## Constraints
- 正確性優先；必須以實際來源正文核對九篇樣本，不能只依成功訊息判定通過。
- 保留先無視窗嘗試、失敗後有視窗重試，以及既有驗證頁重試設定的效果。
- 測試使用隔離設定、瀏覽器資料、執行檔與輸出；保留正式設定、筆記及目前執行中的程序。

## Value case
受益者是使用閱讀清單產生筆記的 kouko。九個樣本在目前執行中持續失敗，等待既有重試尚未解決問題；隔離研究已確認其中兩篇存在正文且有可行的修正方向。本輪集中修復並逐篇驗證這九篇，不擴充其他擷取能力。GO — 問題已實際重現，且使用者明確要求把新增失敗項目一併納入驗收。

## Out of scope
- 其他平台或所有網站的完整擷取保證。
- 登入、付費或驗證限制的繞過。
- 既有筆記的批次重抓、改寫或遷移。

## Open questions
- none
