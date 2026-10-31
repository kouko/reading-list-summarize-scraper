# 修復 Facebook 與 Felix 的完整正文擷取
originator: kouko
kind: product
needs-design: yes — Previously failing pages will produce saved article content, and target identity, completeness rejection, and failure reporting need a specification for this scope.
evidence: [/private/tmp/rlss-content-research-verification-20261030.json, /private/tmp/rlss-live-original-comparison.json]
status: confirmed 2026-10-31
publication: automatic — authorized 2026-10-31 by kouko

## Problem
kouko 的閱讀清單中，已確認可以閱讀原文的 Facebook 貼文與 Felix 文章仍持續擷取失敗，無法產生需要的文章筆記。取得一些文字或顯示成功，也不足以確認沒有漏掉正文或混入留言、登入介面等其他內容。

## Proposed outcome
這兩個失敗樣本能取得正確、完整的目標正文，並能以實際原文核對結果；無法確認正確性或完整性時，保留明確的失敗原因。

## Acceptance
1. 已研究的 Facebook 貼文可取得完整目標正文，保留原有段落、標籤、項目符號、正文連結與圖片，不混入留言、其他貼文或登入介面。
2. 已研究的 Felix 文章可取得完整正文，保留原有段落順序、四個章節、正文連結與圖片，不混入訂閱表單或導覽文字。
3. 目標不符、正文截斷、來源矛盾或僅有登入、驗證、錯誤畫面的內容，不會被保存並回報為成功文章；擷取失敗時保留既有正文與摘要。
4. 擷取失敗的紀錄足以區分載入失敗、目標無法確認、正文解析失敗及完整性不足，讓使用者可以辨識失敗原因。
5. 既有 X 指定貼文與長文、一般文章及分頁文章維持正常，既有瀏覽器重試與設定效果維持一致。
6. 兩個修復樣本都有逐篇與實際來源比對的驗證結果，涵蓋正文順序、章節、連結、標籤、符號及正文圖片，不只依成功訊息或字數判定。

## Constraints
- 正確性優先，不把錯誤或部分正文當成成功，也不宣稱所有網站都能確認全文完整。
- 保留先無視窗嘗試、失敗後有視窗重試，以及既有 headed_on_block 設定的效果。
- 使用隔離的瀏覽器資料、設定、執行檔與輸出進行實測，保留正式設定、筆記及目前執行中的程序。
- 這輪先處理已研究的 Facebook 與 Felix 失敗樣本，沿用現有工具。

## Value case
受益者是以閱讀清單產生文章筆記的 kouko。兩個失敗樣本已持續重試，仍需自行開啟原文閱讀；隔離研究已重現失敗原因，並在未登入的兩種瀏覽器模式取得可核對的正文。使用者已選擇先修復這兩個樣本並建立完整性驗證，將更多網站與多種解析器的擴充留在本輪範圍之外。GO — 現有重試無法解決，且已有實際來源證據支持有限範圍的修復。

## Out of scope
- 所有網站的完整擷取保證，以及未列入本輪的網站支援擴充。
- X 或 Threads 的作者串文組合。
- 登入、付費或驗證限制的繞過。
- 既有筆記的批次重抓、覆寫或遷移。
- 多種解析器自動切換框架。

## Open questions
- none
