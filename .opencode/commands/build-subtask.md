---
description: BUILD one sub-task (use in Build mode). Example: /build-subtask WP01 1
---
Implement ONE sub-task. Arguments: $ARGUMENTS (format: WPxx N).

1. If a plan for this sub-task exists earlier in this session, follow it. Otherwise read AGENTS.md, the last 40 lines of docs/PROGRESS.md and docs/tasks/<WPxx>.md, and read only the files that brief lists.
2. Contract first (interfaces, sqlc queries, OpenAPI), then tests, then the implementation.
3. Respect every invariant in AGENTS.md. Do not edit anything under docs/srs/. Never weaken or delete a test.
4. Run the targeted tests (make test-module M=<module>) and then make ci. If something fails, fix it. If it still fails after 2 attempts, STOP and explain the problem instead of trying more things.
5. If the SRS is silent on something, add the question to docs/OPEN_QUESTIONS.md and stop that part.
6. Do not start the next sub-task. Do not commit.
7. Finish with: files changed, tests run with results, anything unfinished.
