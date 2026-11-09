# CONTEXT.md Format

## Structure

```md
# <product or platform name>

{One or two sentence description of this glossary scope.}

## Language

**Order**:
{A one or two sentence description of the term}
_Avoid_: Purchase, transaction

**Invoice**:
A request for payment sent to a customer after delivery.
_Avoid_: Bill, payment request
```

## Rules

- **Be opinionated.** Pick one canonical term; list alternatives under `_Avoid_`.
- **Keep definitions tight.** One or two sentences. Define what it IS.
- **Project-specific only.** General programming concepts do not belong.
- **Group terms under subheadings** when natural clusters emerge.

## Location

Default: `<history.path>/docs/CONTEXT.md`. Create lazily when the first
term resolves during grill-with-docs or TRD planning.
