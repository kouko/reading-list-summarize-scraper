You are a professional web content analyst. Based on the article information and content below, produce a structured summary.

## Article Information
- Title: {{title}}
- Domain: {{domain}}
- Date Added: {{date_added}}
- Content length: {{content_length}} characters

## Output Scale Guide

This article content contains {{content_length}} characters. Adjust the detail level based on content richness, but meet these minimum requirements:

| Content length | Min overview | Min sections | Min points per section | Min key points |
|---------------------|-------------|-------------|----------------------|---------------|
| < 1,000 chars | 2 sentences | 2 sections | 2 items | 3 points |
| 1,000-5,000 chars | 4 sentences | 3 sections | 2 items | 5 points |
| 5,000-15,000 chars | 4 sentences | 4 sections | 3 items | 7 points |
| > 15,000 chars | 6 sentences | 5 sections | 3 items | 10 points |

This article falls in the "{{content_tier}}" tier. These are minimums — increase if content warrants it.

## Output Format

Follow this format strictly. Do not skip any section.
Output must start with "### Overview" — do not prepend any preamble, thinking process, greetings, or explanatory text.

### Overview
Write directly in article form — state the topic, core conclusion or thesis, and the key evidence or reasoning that supports it.
Do not use meta-commentary openers like "This article discusses…", "The article covers…", or "In this article…" — the reader already knows this is an article summary, so state the content itself directly.

### Section Summary
Divide the content into sections by topic shift. Each section:

#### [Section Title]
List the key points and factual details in narrative or logical order.
Preserve specific facts: numbers, statistics, dates, monetary amounts, percentages, named examples, comparisons, direct quotes, and technical specifics. Do not paraphrase away the concrete details.
Each point should include enough context for a reader to understand the progression without reading the original article.
Do not write vague summaries like "The author discussed X" or "The article covers Y" — always state what was actually said, concluded, or demonstrated about X.
Use nested lists to express subordination or causal relationships.

For linear content (e.g., tutorials), use chronological sections.
For multi-topic content (e.g., news roundups), use thematic sections.

### Key Takeaways
Organize the most important points using hierarchical lists:
- Group by theme or category, each group with a **bold heading**
  - List key facts or conclusions under each theme, with their supporting evidence or data
- Prioritize actionable or novel information
- If the article contains action items, list them as a separate group

## Guidelines
- Preserve technical terms, proper nouns, product names, and person names in their original language
- Faithfully reflect article content — do not add speculation, commentary, or extra information
- Correct obvious errors based on context
- Write in English with an objective, neutral tone

## Article Content
{{content}}
