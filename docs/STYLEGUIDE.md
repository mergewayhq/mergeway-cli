# Mergeway Documentation Style Guide

This guide distills the authoring standards from `docs-bck/good-documentation.md` and Instruction 03 so writers have a single reference while working in mdBook.

## 1. Voice & Tone

- Use active voice and direct, outcome-focused verbs (`Run`, `Validate`, `Inspect`).
- Avoid filler words (`just`, `simply`, `obviously`) and state preconditions and side effects explicitly.
- Prefer short paragraphs (2–3 sentences). Break complex ideas into bullet points or ordered lists.
- Address the reader directly (`You can…`, `Run…`).

## 2. Structure & Headings

- Each page begins with an `# H1` that mirrors the entry in `SUMMARY.md`.
- Follow a consistent heading hierarchy (`##`, then `###`). Do not skip levels.
- Include a brief "Why this matters" lead paragraph after the title.
- Finish guides with a "Next steps" or "Related reading" section to aid navigation.

## 3. Code & Examples

- Fence code blocks with language tags (```` ```bash ````, ```` ```json ````, etc.) so mdBook enables syntax highlighting and copy buttons.
- Provide both minimal and realistic examples:

  ```markdown
  > **Example – Minimal setup**
  > ```bash
  > mw init
  > ```
  >
  > **Example – Advanced configuration**
  > ```bash
  > mw validate --summary --output table
  > ```
  ```

- Inline commands with backticks (`mw validate --summary`) and link to the corresponding reference entry.

## 4. Images & Diagrams

- Store editable sources in `docs/assets-src/architecture/` and exported formats in `docs/src/assets/images/architecture/`.
- Embed diagrams using standard Markdown: `![Workspace flow](../assets/images/architecture/workspace-flow.png)`.
- Provide concise captions that explain the key takeaway.

## 5. Content Types

Refer to the templates under `docs/templates/` when drafting:

- Overview pages: problem statement → capabilities → quick next steps.
- Installation: prerequisites, commands, verification, and optional source build.
- Getting Started: first validation walkthrough with expected output.
- Concepts: definitions, relationships, and validation flow.
- CLI Reference: usage, flags/options table, examples, exit codes, CLI version.
- Schema Format: minimal example, required keys, common field attributes, and reference usage.
- Troubleshooting: symptom → cause → fix, plus FAQ entries.
- Changelog: version/date table with short highlights and any upgrade notes.

## 6. Style Automation

- Include a "Last updated: YYYY-MM-DD" line near the top of each page (automation can fill this in later).
- If introducing alternative linters, document how to run them in contributor guidance.

## 7. Maintenance Rituals

- Every feature PR should update documentation or explicitly explain why no update is needed.
- Review key chapters quarterly. Owners and the review checklist are documented in `docs/README.md`.
- When refactoring sections, update `SUMMARY.md` and ensure legacy links redirect or are noted in release notes.

Following these guidelines keeps the docs clear, trustworthy, and easy to evolve alongside the Mergeway CLI.
