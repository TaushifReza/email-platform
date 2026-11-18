---
description: REVIEW uncommitted changes (use in Plan mode). Example: /review-diff WP01 1
---
Review the uncommitted changes for sub-task $ARGUMENTS. Do not edit any file.

1. Run git status --short and git diff --stat, then read the diff in pieces.
2. Check against the invariants in AGENTS.md: tenant isolation (WithOrg, organization_id), SQL only through sqlc, secrets never logged or returned, no InsecureSkipVerify, no text/template or html/template for user templates, guarded state updates, GET never unsubscribes, no words delivered/opened/bounced.
3. Check that tests exist for every new behavior, that no test was deleted, skipped or weakened, and that nothing outside the sub-task was changed.
4. Check the plan and the brief (docs/tasks/<WPxx>.md): is anything missing or invented?
5. Answer with: PASS or FAIL, then a numbered list of problems (file and line), most serious first. Keep it under 40 lines.
