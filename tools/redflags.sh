#!/usr/bin/env bash
# Quick red-flag scan for agent-written code. Run from the repository root:  bash tools/redflags.sh
# It does not prove the code is safe. It only catches common mistakes quickly.
found=0

report() { # title, matches
  if [ -n "$2" ]; then
    echo "--- RED FLAG: $1"; echo "$2"; echo; found=1
  fi
}

go_files() { find backend -name '*.go' -not -name '*_test.go' -not -path '*/gen/*' -not -path '*/sqlc/*' -not -path '*/db/queries/*' 2>/dev/null; }
frontend_files() { find frontend/app frontend/components frontend/lib -type f \( -name '*.ts' -o -name '*.tsx' \) -not -name 'openapi.ts' 2>/dev/null; }

report "InsecureSkipVerify (must never exist)" \
  "$(go_files | xargs grep -n 'InsecureSkipVerify' 2>/dev/null | head -20)"

report "text/template or html/template used (user templates must use the purpose-built renderer)" \
  "$(go_files | xargs grep -nE '"(text|html)/template"' 2>/dev/null | head -20)"

report "SQL written inside Go code (all SQL must go through sqlc)" \
  "$(go_files | xargs grep -nE '"[[:space:]]*(SELECT|INSERT INTO|UPDATE|DELETE FROM)[[:space:]]' 2>/dev/null | head -20)"

report "Printing or default logger used (use the redacting zap logger)" \
  "$(go_files | xargs grep -nE '(fmt\.Print|log\.Print|log\.Fatal)' 2>/dev/null | grep -v 'cmd/' | head -20)"

report "Hard-coded secret-looking value" \
  "$(go_files | xargs grep -niE '(password|secret|apikey|api_key|token)[[:space:]]*(:=|=)[[:space:]]*"[^"]{8,}"' 2>/dev/null | head -20)"

report "Tests skipped" \
  "$(find backend -name '*_test.go' 2>/dev/null | xargs grep -nE 't\.Skip(Now)?\(' 2>/dev/null | head -20)"

report "Forbidden words in user-facing code (delivered / opened / bounced)" \
  "$( (frontend_files | xargs grep -niE '\b(delivered|bounced|opened)\b' 2>/dev/null; go_files | xargs grep -niE '"[^"]*\b(delivered|bounced)\b[^"]*"' 2>/dev/null) | head -20)"

report "Deleted test files in uncommitted changes" \
  "$(git status --short 2>/dev/null | grep -E '^( D|D ).*_test\.' | head -20)"

big=$( (find backend -type f -name '*.go' -not -name 'openapi.gen.go' -not -path '*/sqlc/*' 2>/dev/null; frontend_files) | xargs wc -l 2>/dev/null | awk '$1>400 && $2!="total"{print $1" lines: "$2}' | head -20)
report "Files over 400 lines (agents should keep files small)" "$big"

if [ "$found" -eq 0 ]; then echo "No red flags found."; else echo "Red flags found. Ask the agent to fix them, or look yourself."; exit 1; fi
