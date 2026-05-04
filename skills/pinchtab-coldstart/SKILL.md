---
name: pinchtab-coldstart
description: "Run the PinchTab cold-start test. Spawns a subagent that builds PinchTab from source, starts the server natively (no Docker), and executes groups 0-1 (14 steps) using only the skill docs. Use when asked to 'run cold start', 'cold-start test', or 'test the agent onboarding flow'."
---

# PinchTab Cold-Start Test

Validate that an AI agent can go from zero to working with PinchTab using only the skill docs — no hand-holding.

## Prerequisites

Before spawning the subagent, clean the environment:

```bash
PROJECT_ROOT=$(git rev-parse --show-toplevel)
pkill -f 'pinchtab' 2>/dev/null
pkill -f 'Google Chrome.*pinchtab' 2>/dev/null
sleep 2
rm -f ~/.local/state/pinchtab/current-tab 2>/dev/null
rm -f "$PROJECT_ROOT/pinchtab" 2>/dev/null
lsof -ti:9867 2>/dev/null | xargs kill 2>/dev/null
```

Wait 2 seconds after cleanup before spawning the agent.

## Execution

Spawn a single subagent with the prompt below. Replace `{PROJECT_ROOT}` and `{TIMESTAMP}` with actual values.

```
You are running a PinchTab cold-start validation. Your working directory is {PROJECT_ROOT}.

Start by reading the context file, then follow its instructions:

1. Read `tests/coldstart/subagent-context.md` — your full instructions.
2. Read the skill files it references.
3. Read the group files it references.
4. Execute all steps in groups 0 and 1.

Report pass/fail for every step. Write your full results to `/tmp/pinchtab-coldstart-{TIMESTAMP}.md`.
```

## Interpreting results

- **14/14 PASS**: The skill docs are sufficient for a cold start.
- **Any failure**: Check what the agent got stuck on — that's a gap in the skill docs or the CLI ergonomics.

Key things to look for in the report:
- Did the agent use the default port (9867) or pick a custom one?
- Did the agent read the server's READY output or poll health in a loop?
- Did the agent use `./pinchtab` CLI or fall back to curl/HTTP API?
- Did the agent modify `~/.pinchtab/config.json` or use a temp config with `PINCHTAB_CONFIG`?
- Did step 1.2 (click follows link) pass without an eval workaround?

## Comparing runs

Track token usage and tool call counts across runs to measure improvements:

| Metric | Good | Needs work |
|--------|------|------------|
| Total tokens | < 40k | > 50k |
| Tool calls | < 40 | > 50 |
| Port | 9867 (default) | Custom port |
| Server wait | Read READY | Polled health |
| API usage | CLI only | curl/HTTP fallback |
