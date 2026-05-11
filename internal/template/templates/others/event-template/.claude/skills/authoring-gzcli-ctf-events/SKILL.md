---
name: authoring-gzcli-ctf-events
description: Use this skill when creating, reviewing, or updating a gzcli CTF event repository or challenge directory. It covers `.gzevent`, category layout, `.example/` starter templates, `.structure/` scaffolding, required challenge roots like `challenge.yml`, `dist/`, `src/`, and `solver/`, plus sync/watch behavior, upload validation constraints, and writeup placement.
---

# GZCLI CTF Event Authoring

## Quick Use
- Confirm whether the task is about the event repo itself or a specific challenge inside a category.
- For structure questions, read `../../../.agents/skills/authoring-gzcli-ctf-events/references/event-repo-structure.md`.
- For challenge folder contents and packaging, read `../../../.agents/skills/authoring-gzcli-ctf-events/references/challenge-directory-contract.md`.
- For sync, watch, and upload behavior, read `../../../.agents/skills/authoring-gzcli-ctf-events/references/sync-watch-and-upload-rules.md`.

## Core Workflow
1. Read `.gzevent` when event metadata matters.
2. Keep challenges under the fixed category directories only.
3. Start new challenges from the closest `.example/` template instead of inventing a layout from scratch.
4. Keep deployable challenge roots compatible with the gzcli contract:
   `challenge.yml`, `dist/`, `src/`, `solver/`, with optional `Dockerfile`, `docker-compose.yml`, `.dockerignore`, and `.gitignore`.
5. Validate the intended solve path end-to-end before treating the challenge as ready.

## Operational Rules
- `challenge.yml` must remain the canonical challenge definition filename.
- Do not leave `challenge.yml` identical to a stock template.
- If `provide: ./dist` is set, `dist/` must contain real files.
- `solver/` must contain meaningful intended-solution material.
- Treat `solver/` and writeup content as documentation, not deployment inputs.

## Related Files
- `../../../.agents/skills/authoring-gzcli-ctf-events/references/event-repo-structure.md`
- `../../../.agents/skills/authoring-gzcli-ctf-events/references/challenge-directory-contract.md`
- `../../../.agents/skills/authoring-gzcli-ctf-events/references/sync-watch-and-upload-rules.md`
