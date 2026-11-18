<!--
Adapted from google/elephant-goldfish-model at
210977c1f76f07961d11d42b34bfd768c9267327
(resources/document_standards/implementation-readiness-standard.md,
agents/implementation_readiness.md), Apache-2.0.
Modified: rewritten in Jarvis's own words; GO/NO-GO verdict and TRD
template sections added. See LICENSES/egm-Apache-2.0.txt.
-->

# Ready to build (TRD go/no-go)

Use this prompt for stage 3, after the common opening in
[lenses.md](lenses.md):

> Decide whether an engineer who does not know this codebase could build the
> change from this TRD alone, on the first try, without asking questions and
> without making a design choice. Answer **GO** or **NO-GO** on the first
> line, then list findings.

## The bar

A TRD is **GO** only when all of these hold:

1. **No design choices left to the implementer.** Data shapes, API paths and
   payloads, error handling, retries, defaults, config keys, and new
   dependencies are decided in the text. "Handle errors appropriately" or
   "update the service" is a NO-GO.
2. **Files to change.** Every file to create, modify, or delete is listed
   with its path (repo-relative) and one line on what changes and why. Tests,
   config, schema, and docs count. A glob is allowed only for generated
   output.
3. **Phases.** The work is split into steps, and each step leaves the code
   building and its tests passing.
4. **Verification.** Named test cases (happy path and edge cases), and how to
   check the change locally. For Tier A, `TRD-Testing.md` exists or the plan
   says where the cases live.
5. **Open questions.** None left that change the design. Remaining open
   questions are marked as not blocking, with an owner.
6. **Links.** Related story and (for an implementation plan) the
   architecture TRD are linked; terms are in the glossary or linked.

Any Blocker in this stage means **NO-GO**. Majors alone may still be GO if
the reviewer says why.
