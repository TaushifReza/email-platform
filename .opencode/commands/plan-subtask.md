---
description: PLAN one sub-task (use in Plan mode). Example: /plan-subtask WP01 1
---
You are planning ONE sub-task. Arguments: $ARGUMENTS (format: WPxx N, for example WP01 1).

1. Read AGENTS.md, the last 40 lines of docs/PROGRESS.md, and docs/tasks/<WPxx>.md.
2. Read ONLY the design and SRS files that the brief lists for this work package. Never read docs/SRS_Email_Automation_Platform.md.
3. Take sub-task number N from the brief. Do not plan other sub-tasks.
4. Answer in at most 60 lines with exactly these parts:
   (a) files to create or change, with full paths
   (b) the contract first: Go interfaces, SQL queries for sqlc, OpenAPI changes
   (c) tests to write, named with requirement IDs
   (d) the exact commands to run to prove it works
   (e) risks, and any question for the owner
5. Do not write code. Do not edit any file.
