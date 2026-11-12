---
description: >-
  Plain English for Jarvis docs read by second-language readers. Applies to
  TRD, TRD-Testing, review text, PR bodies, chat replies, and other Jarvis
  docs. Based on ASD-STE100 Issue 9 writing rules (not the dictionary).
---

# Plain English (based on ASD-STE100)

Readers use English as a second language. Write so they read it once.

## Sentences

1. Instructions and test steps: max 20 words. Descriptions: max 25.
   A number, code identifier, quoted text, or name counts as one word.
2. One instruction per sentence. Write it as a command ("Call", "Check").
3. Condition first, then a comma, then the command ("If the flag is off, ...").
4. Active voice. Name who does it (the app, the service, the reviewer).
5. Keep "the", "a", "this". Do not drop words to save space. No contractions.
6. No semicolons. Split the sentence or use a list.
7. More than two actions or results: use a numbered or bulleted list.

## Words

8. One word = one meaning. Use the term in the TRD, Jira, or glossary.
   Do not use synonyms for the same thing (toast vs snackbar vs alert).
9. Max 3 nouns in a row. Rewrite "stable request identity persistence".
10. Use a verb for the action ("validate", not "perform validation").
11. Keep code, error strings, and UI copy exact, in quotes.

## Paragraphs

12. One topic per paragraph, max 6 sentences.

## Scope

- Applies to: TRD, TRD-Testing, review text, PR bodies, chat replies, other Jarvis docs.
- Chat replies: apply all rules 1 to 12 as written, including rule 5 (no contractions).
- Not: code, code comments, commit subjects, quoted third-party test text.
- Review: style problems are Nits only, never blockers (a step too unclear to run is already a normal blocker).
- Source: ASD-STE100 Issue 9, https://www.asd-ste100.org/ (free copy). The text is ASD copyright. Do not vendor the PDF or the dictionary. Say 'based on ASD-STE100', not 'STE-compliant'.
