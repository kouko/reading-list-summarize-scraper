---
type: lesson
name: verify-owned-body-before-save
description: When accepting extracted article content, capture expected owned units before cleanup and verify the extracted body before saving; metadata or parser success is insufficient.
sources:
  - resource: facebook.ts
  - resource: extraction.ts
  - resource: internal/extract/browser.go
  - resource: internal/pipeline/extraction.go
---

## Trigger
Changing article extraction, parser cleanup or save validation when a page exposes
metadata, video labels, preview text or a nonempty parser result.

## Correct path
Bind native source identity to the requested URL, join it with target-owned DOM
text and media, capture expected ordered owned units before cleanup, then verify
the extracted body against those units before saving.
Reject ambiguous identity, lost prose or unverified attachment extent explicitly.

## Why
Metadata and video labels can describe a readable text post without representing
its body. Cleanup can remove a whole dated body; the observed Felix failure
returned empty content. Parser success or nonempty output alone also cannot prove
that the saved note contains the requested original body.

## Limits
This is a verification discipline, not a completeness guarantee for every site.
Unavailable source evidence must remain unavailable rather than become a pass.
