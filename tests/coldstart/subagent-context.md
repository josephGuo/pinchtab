# Cold-Start Subagent Context

You are running PinchTab cold-start validation. Your job is to build PinchTab from source, start the server natively (no Docker), and execute groups 0-1 against local fixture HTML files.

## What to read

1. **PinchTab dev skill**: `skills/pinchtab-dev/SKILL.md` — how to build the project.
2. **PinchTab skill**: `skills/pinchtab/SKILL.md` — full command reference, configuration, and patterns.
3. **Group files**: `tests/optimization/group-00.md` and `tests/optimization/group-01.md` — step definitions and verification markers.

## What NOT to read

- `tests/tools/scripts/baseline.sh` — deterministic baseline; reading it defeats the purpose.
- `tests/benchmark/` — separate benchmark lane, not your concern.

## Environment

- Project root: the git root (run `git rev-parse --show-toplevel` if needed)
- Go, Python 3, Chrome, curl, and jq are available on the host
- No Docker required — everything runs natively
- Fixture HTML files: `tests/tools/fixtures/` (wiki.html, wiki-go.html, articles.html, dashboard.html, etc.)

## Setup

### 1. Build from source

Use `./dev build` as described in the dev skill. The binary is placed at `./pinchtab` in the project root.

### 2. Start fixture HTTP server

The fixture pages need to be served over HTTP. Use Python's built-in server on a free port:

```bash
FIXTURE_PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("",0)); print(s.getsockname()[1]); s.close()')
python3 -m http.server $FIXTURE_PORT --directory tests/tools/fixtures --bind 127.0.0.1 &
```

### 3. Configure and start PinchTab

Read the PinchTab skill to learn how to configure and start the server. You need:
- A server with auth enabled
- Chrome running in headed mode (not headless)
- `localhost` in allowed domains so you can reach the fixture server
- `allowEvaluate` enabled (some steps use eval)

Start the server and **read its stdout** — it prints `READY` when the instance is up and a hint showing how to create a session and start navigating.

Use `./pinchtab` CLI commands for everything — never use `./scripts/pt` (that is the Docker wrapper) and never use curl against the HTTP API.

## Running steps

Fixture URLs are `http://localhost:$FIXTURE_PORT/` — the group files reference `http://fixtures/` which is the Docker hostname. Replace `http://fixtures/` with `http://localhost:$FIXTURE_PORT/` in every command.

Execute every step in groups 0 and 1. For each step:

1. Run the appropriate PinchTab commands
2. Verify the expected markers appear in the output
3. Record pass/fail with the command used and output

## Cleanup

When finished, kill the fixture server and PinchTab server, remove any temp config, and delete the built binary.
