---
type: lesson
name: verify-owned-body-before-save
description: When accepting extracted article content, prove requested identity and owned body completeness before cleanup and saving; metadata or nonempty output is insufficient.
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
text and media, and compare ordered owned units before cleanup and before saving.
Reject ambiguous identity, lost prose or unverified attachment extent explicitly.

## Why
Metadata and video labels can describe a readable text post without representing
its body. A parser can return plausible nonempty content after dropping a whole
dated body. Neither observation proves that the saved note contains the original.

## Limits
This is a verification discipline, not a completeness guarantee for every site.
Unavailable source evidence must remain unavailable rather than become a pass.
