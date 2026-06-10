---
name: implement-issue
description: Implement a vedanta/asctl GitHub issue end to end — verify dependencies, implement with tests derived from the acceptance criteria, self-review against the design quality gates, open a PR that closes the issue, and enable auto-merge gated on CI. Trigger: /implement-issue <issue-number> [more issue numbers]
---

# Implement a GitHub Issue

Take one issue from `vedanta/asctl` from open → merged PR. If multiple issue
numbers are given, process them strictly in the given order, finishing (or
parking) one before starting the next.

## 1. Load context

1. `gh issue view <N> -R vedanta/asctl --comments` — the issue body is the spec.
2. Read the pinned design reference once per session:
   `gh issue view 1 -R vedanta/asctl` (full DESIGN.md). The local working copy
   may also have `DESIGN.md`, `APICLISPEC.md`, `CLIHELP.md` (gitignored) — prefer
   local copies when present.
3. Parse from the issue body:
   - **Scope** — what to build
   - **Acceptance criteria** — the checklist; this doubles as the test plan
   - **Depends on #N** — prerequisite issues

## 2. Preflight — refuse early instead of failing late

- Every dependency issue must be CLOSED (`gh issue view <dep> --json state`).
  If any is open: comment on the issue explaining which dependencies block it,
  do not start, and report back.
- If the issue is already closed or has an open linked PR, stop and report.
- `git fetch origin && git switch main && git pull` — start from current main.
- If the scope is ambiguous or contradicts the design, comment the question on
  the issue and stop. Do not guess on spec conflicts.

## 3. Branch

```bash
git switch -c issue-<N>-<short-slug>
```

One issue per branch. Never commit to main.

## 4. Implement

- Follow the architecture in design §7: layering `cli → domain → asc → client`;
  command handlers never build HTTP requests; `client` never formats output.
- Match existing code style, naming, and helper usage. Reuse the shared helpers
  (printer, confirm, resolve, page iterator) — do not reimplement them locally.
- Help text follows CLIHELP rules (design §6): outcome-first short help,
  tiered long help, progressive commented examples, safety notes, Next: pointer
  on core workflow commands. Help strings live in `internal/cli/helptext/`.

## 5. Tests — derived from acceptance criteria

For each acceptance criterion, decide how it is verified:

| Criterion type | Test type |
| --- | --- |
| Pure logic (planner, importer, config precedence, JWT claims, resolution) | Unit table-tests |
| Command behavior against the API | `httptest.Server` integration with fixtures in `testdata/fixtures/` |
| Table or `--help` rendering | Golden files in `testdata/golden/` (regenerate with `go test -update`) |
| Safety behavior (confirm, dry-run) | Integration test asserting prompts and that dry-run issues zero write requests |

Rules:

- A criterion without a covering test is not done — either write the test or
  state in the PR body why it cannot be tested (live-API-only behavior).
- If the issue needs no new tests at all (pure infra, e.g. CI config), the PR
  body must say so explicitly with the reason.
- New fixtures are recorded JSON:API shapes — keep them minimal and realistic.

## 6. Verify locally

```bash
go build ./...
golangci-lint run        # if installed; otherwise go vet ./...
go test -race ./...
```

All green before review. Fix failures — do not skip or weaken tests to pass.

## 7. Self-review

Review the full diff (`git diff main`) against the design §9 quality gates:

- [ ] Resource-first command shape, standard verb vocabulary
- [ ] Table default; `-o json` is clean, stable, never raw JSON:API
- [ ] Destructive paths confirm by default and accept `-y`
- [ ] Bulk paths support `--dry-run`
- [ ] Pagination handled (`--all`/`--limit` where applicable)
- [ ] API errors mapped to Error/Try with next actions
- [ ] Help has copyable examples; golden tests exist for new help/table output
- [ ] No layering violations; no auth/token material in logs or output

Fix everything found. For substantive changes, re-run step 6.

## 8. Pull request

```bash
git push -u origin HEAD
gh pr create -R vedanta/asctl --title "<issue title>" --body "..."
```

PR body must contain:

- `Closes #<N>`
- Short summary of the approach (a paragraph, not a diff narration)
- The issue's acceptance criteria as a checked checklist, each with how it is
  verified (test name or rationale)
- Test evidence: the `go test` summary line
- If no new tests: the explicit justification

## 9. Merge

```bash
gh pr merge --auto --squash
```

- Auto-merge fires when required checks pass. **Caveat:** until CI (issue #9)
  exists and is a required check on main, auto-merge merges immediately —
  acceptable for the M0 bootstrap issues only; flag it in the report.
- If CI fails: fix on the branch, push, let auto-merge retry. Do not merge over
  a red build, and do not use `--admin` to bypass checks.
- After merge: `git switch main && git pull` before the next issue.

## 10. Report

End with: issue number, PR number and state (merged / awaiting checks /
blocked), test summary, and anything punted (with the follow-up issue or
comment created for it).
