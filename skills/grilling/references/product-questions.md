# Product question set

Use this list when grilling for a PRD. The rounds method stays the same.
Ask only the topics that are still open. A small story may need only a few.

## Topics and example questions

1. **Problem and evidence**
   - What problem do users have today?
   - How do we know? (support tickets, data, interviews, a request from a partner)
   - What happens if we do nothing?
2. **Users and personas**
   - Who has the problem? Who uses the feature?
   - Who is affected but does not use it (for example an approver or support staff)?
3. **Goals and success metrics**
   - What number must change, from what baseline, to what target?
   - How and when do we measure it?
4. **Scope and out of scope**
   - What is the smallest version that solves the problem?
   - What do we say "no" to in this release?
5. **Main and alternative flows**
   - What are the steps of the main flow, from the user's point of view?
   - What other paths exist (cancel, retry, come back later)?
6. **Business rules and limits**
   - Which rules must always hold? State each one so a test can check it.
   - Which limits apply (amounts, counts, time windows, regions)?
7. **Acceptance criteria**
   - For each rule or flow step: given what, when what, then what?
8. **Edge cases, errors, and empty states**
   - What does the user see when data is missing, late, or wrong?
   - What happens on a double submit, a timeout, or a lost connection?
9. **Permissions and roles**
   - Who can see, create, change, approve, or delete?
10. **Data and reports**
    - Which data must we keep, show, or export? For how long?
    - Who needs a report, and how often?
11. **Dependencies**
    - Which teams, systems, vendors, or legal reviews are needed?
12. **Rollout, flags, migration, and communication**
    - Do we release to everyone, a pilot group, or one region first?
    - Is there existing data to migrate?
    - Who must be told, and when? What makes us roll back?
13. **Open risks**
    - What could make this fail after release?

## What not to ask

- Architecture, services, databases, or APIs.
- Story tiers, test gates, or coverage numbers.
- How the current system behaves. If you cannot confirm it from documents,
  add it to the "Engineering checks" table with status `open`.
