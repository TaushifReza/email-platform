# Owner's Guide: building the platform with OpenCode, step by step

This guide is for **you**, not for the AI. Follow it from top to bottom. Every step says **Do**, **You should see**, and **If not**.

**The big picture in four lines**
1. You open OpenCode inside the project folder.
2. You give it one small job (a "sub-task") using a short command.
3. The AI writes code and tests. **You** run the tests yourself and check the result.
4. You save the work (commit) and move to the next sub-task.

You are the manager. The AI writes. The tests prove. You decide.

> Commands in this guide were checked against the OpenCode documentation. If your OpenCode version looks different, type `/help` inside OpenCode.

---

## PART A: One-time setup (about 1 to 2 hours)

### A1. First, check where your AI models really run (important)

Your config points to `http://127.0.0.1:11434`, which means "this computer". But your models are big: a 24B to 30B model needs very roughly **16 to 20 GB of memory** at a common size, plus more for the 64k to 131k context. A machine with about 12 GB of RAM cannot run them. So one of these must be true:

| Situation | What it means |
|---|---|
| **A.** The models run on a stronger machine (GPU PC or server) and `127.0.0.1` is a tunnel to it | Fine. Nothing to change |
| **B.** The models run on this machine's GPU (enough VRAM) | Fine, but keep other programs closed |
| **C.** The models run on this 12 GB machine only | **It will not work properly.** Stop and tell me. We will use a hosted model or another machine |

**Do** (inside your WSL terminal):
```bash
curl -s http://127.0.0.1:11434/api/tags | head -c 600
```
**You should see** JSON listing your models (names such as `qwen3-coder-64k:latest`).
**If not:** if Ollama runs on Windows and you are inside WSL, WSL may not reach it. Fix: create or edit `C:\Users\<you>\.wslconfig` on Windows and add
```
[wsl2]
networkingMode=mirrored
memory=8GB
swap=8GB
```
then run `wsl --shutdown` in Windows PowerShell and open WSL again. Try the `curl` again.

**Do** next (if `ollama` is available where you are):
```bash
ollama list
```
**You should see** the exact model names. Compare them with `opencode.json` in step A6. They must match **exactly**, including the part after the colon.

**Do** one speed test: ask a model something small and time it.
```bash
time curl -s http://127.0.0.1:11434/api/generate -d '{"model":"qwen3-coder-64k:latest","prompt":"Say hello","stream":false}' | head -c 300
```
If this takes more than about 30 seconds, the models are too slow for this work. Tell me the number.

### A2. Put the project in the right place

Because you work in WSL, **keep the project inside the Linux side**, not under `/mnt/c/...` (that is very slow and causes strange errors).

**Do**
```bash
mkdir -p ~/projects
# If your folder is on Windows (for example C:\Users\you\email-platform), copy it:
cp -r /mnt/c/Users/<you>/email-platform ~/projects/email-platform
cd ~/projects/email-platform
ls
```
**You should see:** `AGENTS.md  docs  .github` (and `docs` contains `tasks`, `technical-design`, `srs`, `review`).
**If not:** the zip was unpacked one level too deep. Move the folder contents up so that `AGENTS.md` is directly inside `email-platform`.

### A3. Add the setup files I gave you

You received a second zip, `opencode-setup-files.zip`. It contains: `opencode.json`, `.gitignore`, `.opencode/commands/*`, `tools/redflags.sh`, and this guide.

**Do:** unpack it **into the project root** (the same folder as `AGENTS.md`), then check:
```bash
cd ~/projects/email-platform
ls -a
ls .opencode/commands
```
**You should see:** `.gitignore  .opencode  opencode.json  OWNER_GUIDE.md  tools` and six command files (`plan-subtask.md`, `build-subtask.md`, `review-diff.md`, `handoff.md`, `explain.md`, `stuck.md`).

### A4. Install the tools you need

**Do:** check each tool. Each command should print a version.
```bash
git --version
go version
node --version
docker --version
docker compose version
make --version
```
**If one is missing:** install it. In WSL Ubuntu, the basics are:
```bash
sudo apt update && sudo apt install -y git make curl build-essential
```
- Go: install the current stable Go from go.dev (the AI will tell you the minimum version in WP01).
- Node: use the current LTS version (for example with `nvm`).
- Docker: either Docker Desktop with WSL integration turned on, or Docker Engine inside WSL. **On a 12 GB machine, Docker Engine inside WSL uses less memory.**

**Memory tip for a 12 GB machine:** close the browser tabs you do not need, never run the staging stack locally, and keep the `.wslconfig` limits from step A1.

### A5. Start Git (do this before the AI writes anything)

OpenCode's `/undo` command only works in a Git project, so this is not optional.

**Do**
```bash
cd ~/projects/email-platform
git init
git branch -M main
git config user.name "Your Name"
git config user.email "you@example.com"
git add -A
git commit -m "Add requirements, design and task documents"
git log --oneline
```
**You should see** one commit.

**Create the GitHub repository** (do this when you are ready, it is not needed for the first sub-tasks):
1. On github.com click **New repository**. Name it `email-platform`. Choose **Private**. Do **not** add a README or .gitignore.
2. Copy the commands GitHub shows under "push an existing repository", which look like:
```bash
git remote add origin git@github.com:<your-username>/email-platform.git
git push -u origin main
```
3. Open `.github/CODEOWNERS` and replace every `@OWNER` with `@<your-github-username>`. Commit that change.

### A6. Check the OpenCode configuration

The file `opencode.json` in the project root tells OpenCode which model to use in which mode:

| Mode (press Tab to switch) | Model | Used for |
|---|---|---|
| **Plan** | `qwen3.8-132k` (big context, read-only) | Planning a sub-task, reviewing a change |
| **Build** | `qwen3-coder-64k` | Writing code and running tests |
| **explore** (helper) | `gemma4-64k` | Read-only questions about the code |

Your `provider` block (the part that defines the Ollama models) can stay in your global OpenCode config. OpenCode merges the global file with the project file.

**Do:** open `opencode.json` and make sure each model name matches your `ollama list` output exactly (`ollama/<name>:<tag>`). Fix any difference.

### A7. Start OpenCode and test it

**Do**
```bash
cd ~/projects/email-platform
opencode
```
**You should see** the OpenCode screen with a text box at the bottom.

**Check 1: models.** Type `/models` and press Enter. **You should see** your Ollama models. Press Escape to close.

**Check 2: modes.** Press **Tab**. **You should see** the mode name near the text box change between **Build** and **Plan**. The model name shown may change too.

**Check 3: the AI can read the project.** Switch to **Plan** with Tab, then type:
```
Read AGENTS.md and list the non-negotiable invariants in your own words, in 8 bullet points. Do not read any other file.
```
**You should see** a list that mentions tenant isolation, sqlc, secrets, no InsecureSkipVerify, no text/template, never send twice, GET never unsubscribes. If it does, the setup works.

> **Do NOT run `/init`.** It is meant to create an AGENTS.md. You already have a carefully written one and `/init` could change it.

Press `Ctrl+C` or type `/exit` when you are done.

---

## PART B: The loop you repeat for every sub-task

Each work package (WP01 to WP11) is divided into numbered **sub-tasks** in its brief (`docs/tasks/WP01.md`, and so on). You do the loop below once per sub-task. A normal pass takes **20 to 60 minutes**.

### The loop

**Step 1. Start fresh.**
In OpenCode type `/new`. (A fresh session keeps the small context clean. Never carry one session across many sub-tasks.)

**Step 2. Plan (read-only).**
Press **Tab** until the mode says **Plan**. Type:
```
/plan-subtask WP01 1
```
(Replace `WP01 1` with the work package and sub-task number.)
**You should see** a short plan (under 60 lines): files, contracts first, tests, commands, risks.
**Read it.** Ask yourself: is it only about this one sub-task? Does it list tests? If it is vague, type:
```
Be more specific. List exact file paths and the exact test function names. Do not write code.
```
If it wants to read the big SRS file, type:
```
Do not read docs/SRS_Email_Automation_Platform.md. Use only the files listed in the brief.
```

**Step 3. Build.**
Press **Tab** to switch to **Build**. Type:
```
/build-subtask WP01 1
```
Wait. OpenCode may ask permission to run commands or edit files. For now choose **allow once**. After you trust a command such as `make` or `go test`, you may choose **always** for those.
**You should see** it write files, run tests, and finish with a short summary.
Press **Escape** if you need to interrupt it.

**Step 4. Check it yourself. Do not trust the summary.**
Open a **second terminal** (a new WSL window) and run:
```bash
cd ~/projects/email-platform
git status --short
git diff --stat
make ci
bash tools/redflags.sh
```
- `git status` and `git diff --stat` show **what changed**. It should match the plan. If a lot of unexpected files changed, something is wrong.
- `make ci` runs all the tests. **It must pass.** (In the very first sub-tasks, `make ci` may not exist yet. Then run the command the plan told you.)
- `redflags.sh` is a quick scan for common dangerous mistakes. **It must say "No red flags found."**

**Step 5. Second opinion.**
Back in OpenCode, press **Tab** to **Plan** and type:
```
/review-diff WP01 1
```
**You should see** PASS or FAIL with a list. If FAIL, go back to **Build** and type: `Fix these problems from the review: <paste the list>. Then run make ci.` Repeat Step 4.

**Step 6. Understand it.**
Type `/explain`. Read the 3 riskiest places. For critical sub-tasks (see Part E) open those files and read them yourself.

**Step 7. Save the work.**
Type `/handoff WP01 1`. Then in the second terminal:
```bash
git add -A
git commit -m "WP01.1: short description of what was done"
git log --oneline | head -3
```

**Step 8. Next.** Type `/new` and begin again at Step 2 with the next number.

### The two-strike rule (very important)

If a sub-task is not working after **two fix attempts**, or takes more than **60 minutes**, do **not** keep pushing.
1. Type `/stuck` and read the report.
2. Type `/undo` (repeat if needed) to throw away the changes. Or run `git restore . && git clean -fd` in the second terminal.
3. Type `/new`.
4. Try again with one of: a **smaller** version of the sub-task, the **other build model** (`/models`, pick `Devstral Small 2`), or ask me for help with the exact error.

### Branches and pull requests (keep `main` clean)

One branch per work package, one commit per sub-task.
```bash
git switch -c wp01-foundations        # at the start of a work package
# ... do the loop for every sub-task, committing each ...
git push -u origin wp01-foundations   # at the end of the work package
```
Then on GitHub open a **Pull request** from your branch into `main`. When GitHub Actions shows **green**, click **Merge**. (Before WP01 sub-task 4 there is no CI yet, so just push.)
After the first CI run works, protect `main`: GitHub, Settings, Branches, add a rule requiring the CI check to pass before merging.

---

## PART C: Your very first run (WP01, sub-task 1), keystroke by keystroke

1. Terminal: `cd ~/projects/email-platform && git switch -c wp01-foundations`
2. Terminal: `opencode`
3. Type `/new` Enter.
4. Press Tab until **Plan**.
5. Type `/plan-subtask WP01 1` Enter. Wait.
6. Read the plan. It should create the backend folder layout, a Makefile with the targets listed in `docs/technical-design/06-test-strategy.md`, lint config, `sqlc.yaml`, goose setup.
7. Press Tab to **Build**.
8. Type `/build-subtask WP01 1` Enter. Allow commands when asked.
9. When it finishes, in the second terminal run `git status --short`, `git diff --stat`, `ls backend`, and `make ci` if the Makefile exists.
10. Back in OpenCode: Tab to **Plan**, `/review-diff WP01 1`.
11. If PASS: `/handoff WP01 1`, then `git add -A && git commit -m "WP01.1: repo skeleton and Makefile"`.
12. `/new`, then start sub-task 2.

**Day 1 target:** WP01 sub-tasks 1 to 4. **Day 2:** sub-tasks 5 to 8.

---

## PART D: Daily routine

**Start of day (5 minutes)**
```bash
cd ~/projects/email-platform
git status --short           # should be clean
git log --oneline | head -5  # remember where you stopped
tail -40 docs/PROGRESS.md    # the last handoff note
```
Then `opencode` and `/new`.

**During the day:** repeat the loop. Take a break while the AI is working. Do not give it two jobs at once.

**End of day:** make sure everything is committed. Push the branch: `git push`.

**Speed check (do it at the end of Day 2 and again at the end of week 1)**
Count the sub-tasks you finished. The full first release (M0) has **77 sub-tasks** across the 11 work packages. The 4-week plan needs roughly **4 per working day**.
- 4 or more per day: you are on schedule.
- 2 to 3 per day: the plan will take 6 to 8 weeks. Tell me, and we will cut scope (the cut order is in `docs/technical-design/07-plan.md`) or use a stronger model for part of the work.
- Fewer than 2 per day: the local models are too slow or too weak for this. Tell me what is going wrong.

---

## PART E: How to review when you are not reading every line

### The 5-minute check after every sub-task
1. `make ci` is green.
2. `bash tools/redflags.sh` says no red flags.
3. `git diff --stat` size is reasonable. **If one sub-task changed more than about 600 lines, undo it** and ask for a smaller piece.
4. Check that tests were added and **none were deleted or weakened**:
```bash
git diff --stat -- '*_test.go' '*.test.ts' '*.test.tsx'
git status --short | grep -i test
```
A suspicious sign: the AI "fixed" a failing test by changing the test instead of the code.
5. `/review-diff` said PASS.

### Files you must read personally (or have a stronger model review)
These protect other organizations' data and passwords. A mistake here is serious:
- `backend/internal/platform/crypto/` (encryption)
- `backend/internal/platform/db/` (tenant isolation, RLS)
- `backend/internal/identity/` (login, sessions, 2FA, permissions)
- `backend/internal/providers/dialer/` (blocks internal network access)
- `backend/internal/sending/` (never send twice)
- `backend/internal/suppression/` (unsubscribe)
- every file in `backend/db/migrations/`

For these, do not just trust the tests the AI wrote. Also use the **gate checklists** in `docs/review/`.

### Optional: a stronger reviewer for the critical files
Local 24B to 30B models are weaker at security code. If you have access to a stronger hosted model:
1. In OpenCode type `/connect` and add the provider (it asks for an API key).
2. Type `/models` and pick the stronger model.
3. Switch to **Plan** mode and paste:
```
Review the files in backend/internal/platform/crypto for security problems only. Check: nonce reuse, wrong or missing associated data, key handling, secrets in logs or errors, timing issues, and anything that would let ciphertext be moved between rows or tenants. List problems by severity with file and line. Do not edit files.
```
4. Replace the path for each critical module. Then pick your local model again with `/models`.

---

## PART F: When things go wrong

| What you see | What to do |
|---|---|
| The AI repeats the same action again and again | Press Escape. Type `/stuck`. Then `/undo`, `/new`, retry with a smaller task |
| It says the context is full or answers get strange and slow | Type `/compact` once. If it continues, `/new` and restart that sub-task |
| It starts reading `docs/SRS_Email_Automation_Platform.md` or many big files | Press Escape. Tell it: `Read only the files listed in docs/tasks/<WPxx>.md.` |
| It writes code for several sub-tasks at once | Press Escape, `/undo`, `/new`, and repeat with the exact sub-task number |
| It changes or deletes tests to make them pass | Reject the change. `/undo`. Tell it: `Do not change tests. Fix the code.` |
| It adds a library you did not expect | Ask: `Why this dependency, what is its license and is it maintained?` If the answer is weak, `/undo` |
| It invents a business rule | Check `docs/srs/`. If the SRS is silent, add the question to `docs/OPEN_QUESTIONS.md` and ask me |
| `make ci` fails on something unrelated to the sub-task | Do not fix by hand. Ask Build: `Fix only this failure: <paste the error>` |
| A tool is missing (`sqlc`, `goose`, `golangci-lint`) | Ask the AI how WP01 installs tools (`make tools`). Do not install random versions |
| The model is very slow or times out | Check `ollama ps`. Another program may be using memory. Close it. Consider a smaller context or a different model |
| Memory runs out (WSL freezes) | Stop Docker containers you do not need (`docker ps`, then `docker stop <name>`), close other programs, and keep the `.wslconfig` limits from step A1. If it keeps happening, this machine is too small for the dev stack plus the models, so tell me |
| You are lost | `git status`, `git log --oneline | head`, `tail -40 docs/PROGRESS.md`, and ask me with the output |
| You want to throw away everything since the last commit | `git restore . && git clean -fd` (this deletes uncommitted work) |

---

## PART G: Which model to use when

| Model (your names) | Mode | Use it for | Do not use it for |
|---|---|---|---|
| **Qwen3.8 27B (131k)** | Plan | Planning a sub-task, `/review-diff`, `/explain`, reading several design files | Writing lots of code |
| **Qwen3-Coder 30B-A3B (64k)** | Build | Default for writing code and tests | Planning across many files |
| **Devstral Small 2 24B (64k)** | Build (backup) | A second attempt when Qwen3-Coder is stuck. **Start a `/new` session when you switch** | Mixing two models in one session |
| **Gemma4 26B (64k)** | explore | Questions like "where is X defined?" or "summarize this package" | Editing files |
| **Qwen3-VL 2B (32k)** | none | Only if you must read a screenshot | Any coding |

**Rule of thumb:** one model per session. Switching models in the middle of a sub-task confuses both.

---

## PART H: The master checklist (do in this order)

Tick each box when the sub-task is committed. The sub-tasks are numbered in each brief. The `Gate` rows are review stops: **do not continue until the gate passes.**

| Order | Work package | Brief | Sub-tasks | Notes |
|---|---|---|---|---|
| 1 | **WP01** Foundations | `docs/tasks/WP01.md` | 8 | Tooling, CI, fake SMTP server |
| 2 | **WP02** Tenancy and identity | `docs/tasks/WP02.md` | 9 | Login, 2FA, roles, isolation |
| 3 | **WP03** Crypto and providers | `docs/tasks/WP03.md` | 7 | Encrypted credentials, SSRF protection |
| **Gate 1** | | `docs/review/GATE1-security-core.md` | | Read the critical files. Do not skip |
| 4 | **WP04** Templates | `docs/tasks/WP04.md` | 6 | |
| 5 | **WP05** Imports and chunked uploads | `docs/tasks/WP05.md` | 7 | |
| 6 | **WP06** Attachments | `docs/tasks/WP06.md` | 7 | |
| 7 | **WP07** Campaigns | `docs/tasks/WP07.md` | 6 | |
| 8 | **WP08** Sending pipeline | `docs/tasks/WP08.md` | 10 | **Hardest package.** Consider a stronger reviewer. Can start after WP02 if you want to overlap |
| **Gate 2** | | `docs/review/GATE2-primary-workflow.md` | | Full click-through on staging |
| 9 | **WP09** Suppression and unsubscribe | `docs/tasks/WP09.md` | 6 | |
| 10 | **WP10** Reports, notifications, audit viewer | `docs/tasks/WP10.md` | 5 | |
| 11 | **WP11** Operations and deployment | `docs/tasks/WP11.md` | 6 | |
| **Gate 3** | | `docs/review/GATE3-pilot.md` | | Go or no-go for a first real campaign |

Total: **77 sub-tasks.** Do not overlap work packages unless a brief says it can start early. One person reviewing is the bottleneck, so go one at a time.

---

## PART I: Quick reference card

**Inside OpenCode**

| You want to | Type or press |
|---|---|
| New clean session | `/new` |
| Switch Plan and Build | **Tab** |
| Choose a model | `/models` |
| Plan one sub-task | `/plan-subtask WP01 1` (Plan mode) |
| Build one sub-task | `/build-subtask WP01 1` (Build mode) |
| Review the changes | `/review-diff WP01 1` (Plan mode) |
| Explain the changes | `/explain` |
| Write the handoff note | `/handoff WP01 1` |
| Agent is looping | Escape, then `/stuck` |
| Undo the last AI step | `/undo` (and `/redo`) |
| Context is getting full | `/compact` |
| List old sessions | `/sessions` |
| Mention a file to the AI | type `@` and the file name |
| Interrupt | Escape |
| Help | `/help` |
| **Never run** | `/init` |

**In the terminal**
```bash
make ci                       # all tests
bash tools/redflags.sh        # quick danger scan
git status --short            # what changed
git diff --stat               # how much changed
git add -A && git commit -m "WP01.1: what was done"
git switch -c wpNN-name       # new branch for a work package
tail -40 docs/PROGRESS.md     # last handoff notes
```

**If the custom commands do not work** (for example a typo or a different version), paste this by hand in the right mode:

*Plan mode:*
```
Plan ONE sub-task: WP01 sub-task 1. Read AGENTS.md, the last 40 lines of docs/PROGRESS.md and docs/tasks/WP01.md, and only the files that brief lists. Answer in at most 60 lines: files to change, contracts first, tests with requirement IDs, commands to run, risks. Do not write code or edit files.
```
*Build mode:*
```
Implement ONE sub-task: WP01 sub-task 1. Follow the plan above. Contract first, then tests, then code. Respect AGENTS.md. Never weaken a test. Run make ci. After 2 failed fixes stop and explain. Do not start the next sub-task and do not commit. Finish with files changed and test results.
```

---

## Rules to remember (print this)
1. **One sub-task per session.** `/new` every time.
2. **Plan first, then Build.** Tab switches.
3. **You run `make ci`.** Do not trust a summary.
4. **Two strikes, then undo** and make it smaller.
5. **Never skip a gate.** Safety is never cut.
6. **Critical files are read by a human** (you), or by a stronger reviewer.
7. **Commit after every green sub-task.**
8. **When in doubt, ask me with the exact error text.**
