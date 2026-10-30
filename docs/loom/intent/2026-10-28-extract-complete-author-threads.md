# 作者串文計畫：先釋出 X 指定貼文擷取
originator: kouko
kind: product
needs-design: yes — This changes which posts belong to a saved article and what counts as complete extraction, including recovery from a middle post and rejection of incomplete or unrelated conversation content.
status: confirmed 2026-10-28
publication: automatic — authorized 2026-10-28 by kouko

## Problem
kouko 的既有閱讀筆記有些只包含作者分段發文中的一部分，缺少前文或續文，後續閱讀與摘要因而不是建立在完整文章上。逐篇查核已找到多組真實案例，而且有些曾成功保存的正文仍缺段。

## Proposed outcome
先讓 kouko 能取得指定 X 貼文或長文的正確完整內文，正文與摘要清楚標示未包含作者後續回覆；完整 X／Threads 串文留待後續。

## Acceptance
1. 指定 X／Twitter 貼文網址時，取得該篇正確且完整的內文，包含長文、原有換行與作者連結；從首篇或中段進入都只取得指定的那一篇。
2. 正文與產生的摘要清楚標示僅擷取指定貼文、未包含作者後續回覆，保留來源連結，不混入其他貼文、讀者留言、引用內文或推薦內容；本階段不開放 X／Threads 完整串文擷取。
3. 目標或完整內文無法確認、遇到登入／驗證頁、預覽或空白內容，以及處理暫緩支援的 Threads 貼文時，均回報失敗，不拿舊快取繞過驗證、不啟動摘要，並保留既有正文與摘要。
4. 一般文章、YouTube、略過完成項目、明確重抓、先無視窗再有視窗重試及驗證頁重試設定維持既有效果；驗證全程隔離，不批次改寫既有筆記。

## Constraints
- 正確性優先，不以放寬成功標準或僅檢查成功訊息代替原文完整性核對。
- 沿用既有閱讀清單工具、先無視窗嘗試及失敗後有視窗重試的既定選擇，保留驗證頁重試開關的效果。
- 分段釋出驗收沿用已確認的樣本：X 首篇與中段各自核對指定貼文，Threads 確認暫緩支援時不寫入；不宣稱原本完整串文條件已通過。
- 開發與驗證使用隔離設定、瀏覽器資料、執行檔及輸出，保留正式設定、筆記與執行中的程序。

## Value case
受益者是使用閱讀清單產生文章筆記的 kouko。先前逐篇查核確認完整串文擷取仍有缺文與文章歸屬問題；目前已可從 X 的指定貼文與長文取得可核對的內文。先交付清楚標示範圍的單篇內容，讓 X 可以使用，同時避免把未完成的完整串文當成成功。GO — 使用者接受先釋出此範圍，完整串文需求保留至後續。

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


## Deferred original acceptance (not satisfied by this stage)

1. 已確認的三組 X 與五組 Threads 樣本均能交付完整文章；同一組樣本的首篇與中段網址均對應完整且順序一致的正文。
2. 組合內容只包含同一篇文章的作者正文與續文，保留原有內容順序，同一篇貼文不重複收錄，不混入讀者留言、作者回覆讀者的對話、引用貼文或推薦文章。
3. 無法確認目標、續文關係或完整正文時，不把缺段、預覽、登入或驗證畫面交付為完整文章的成功結果；擷取失敗時保留既有正文與摘要。
4. 既有單篇貼文、長文及一般文章仍正常擷取，原有瀏覽器重試與驗證頁設定效果維持正常。
