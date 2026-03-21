# Sync, Watch, and Upload Rules

## Watcher Behavior

File changes map to update behavior:
- `challenge.yml`: metadata update
- `dist/`: attachment update
- `src/`: full redeploy
- `Dockerfile`, `docker-compose.yml`, `Makefile`: full redeploy
- `solver/` and writeup content: no deploy action

Anything outside the allowed challenge update paths is ignored by the watcher.

## Upload Validation Expectations

The upload path expects:
- the challenge definition file to be named exactly `challenge.yml`
- `dist/`, `src/`, and `solver/` to exist
- `solver/` to contain meaningful content
- `challenge.yml` to differ from the stock embedded templates
- `dist/` to contain real files when `provide: ./dist` is used

Additional rules:
- if `dashboard.config` is present, it must point to `./src/docker-compose.yml`
- for `StaticContainer` using a local image tag, include `Dockerfile` or `src/Dockerfile`

## Recommended Development Loop
1. Create or update the challenge from the closest `.example/` template.
2. Validate the full solve path from a clean start.
3. Use `gzcli sync` for explicit synchronization.
4. Use `gzcli watch start` for iterative development after the base layout is correct.
