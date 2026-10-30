# 將作者串文合併成完整文章
originator: kouko
kind: product
needs-design: yes — This changes which posts belong to a saved article and what counts as complete extraction, including recovery from a middle post and rejection of incomplete or unrelated conversation content.
status: confirmed 2026-10-28
publication: automatic — authorized 2026-10-28 by kouko

## Problem
kouko 的既有閱讀筆記有些只包含作者分段發文中的一部分，缺少前文或續文，後續閱讀與摘要因而不是建立在完整文章上。逐篇查核已找到多組真實案例，而且有些曾成功保存的正文仍缺段。

## Proposed outcome
kouko 能從 X 與 Threads 的作者串文取得正確、完整且按原文順序組合的文章，而非只有指定的單篇貼文。

## Acceptance
1. 已確認的三組 X 與五組 Threads 樣本均能交付完整文章；同一組樣本的首篇與中段網址均對應完整且順序一致的正文。
2. 組合內容只包含同一篇文章的作者正文與續文，保留原有內容順序，同一篇貼文不重複收錄，不混入讀者留言、作者回覆讀者的對話、引用貼文或推薦文章。
3. 無法確認目標、續文關係或完整正文時，不把缺段、預覽、登入或驗證畫面交付為完整文章的成功結果；擷取失敗時保留既有正文與摘要。
4. 既有單篇貼文、長文及一般文章仍正常擷取，原有瀏覽器重試與驗證頁設定效果維持正常。

## Constraints
- 正確性優先，不以放寬成功標準或僅檢查成功訊息代替原文完整性核對。
- 沿用既有閱讀清單工具、先無視窗嘗試及失敗後有視窗重試的既定選擇，保留驗證頁重試開關的效果。
- 驗收以本次逐篇查核確認的八組串文、九個已處理網址為依據，包含從串文中段進入的樣本。
- 開發與驗證使用隔離設定、瀏覽器資料、執行檔及輸出，保留正式設定、筆記與執行中的程序。

## Value case
受益者是使用閱讀清單產生文章筆記的 kouko。本次已逐篇開啟一百零一個已處理網址，確認八組串文且部分舊正文漏段；繼續只取得單篇內容，無法解決已觀察到的缺文。現有替代方式是逐篇開啟原頁與回覆人工核對，本次查核已付出這項工作。本輪優先處理已確認的串文案例，其他平台與其他擷取能力不納入。GO — 已有具體來源、漏段證據及明確的使用者實作要求。

## Out of scope
- X 與 Threads 以外平台的串文。
- 整個帳號、所有留言或任意讀者對話的彙整。
- 繞過登入、付費或驗證限制。
- 既有筆記的自動批次重抓、改寫或遷移。

## Open questions
- none


## Confirmed staged outcome — 2026-10-30

user-decided — Release X specified-post and long-article extraction first.
Content and summaries explicitly exclude follow-up author replies. Complete
X/Threads thread assembly is deferred; a failed complete thread must never be
silently presented as a successful single-post article. The user accepted this
visible behavior with “好” after the recommendation. The spec's staged REQ-1–4
are this release's acceptance contract; original full-thread goals above remain
unmet and deferred, not waived as passing.
