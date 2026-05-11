# GZCLI Event Repo Instructions

## Scope
Applies to the entire generated event repository under `events/<event-name>/`.

## Use Local Skills
- Use `.claude/skills/authoring-gzcli-ctf-events/SKILL.md` whenever the task involves this event's structure, `.gzevent`, category folders, `.example/`, `.structure/`, challenge packaging, sync/watch behavior, upload validation, solver expectations, or writeup placement.
- The skill body links into `.agents/skills/authoring-gzcli-ctf-events/references/` for the detailed contract; treat those references as the source of truth.
- Invoke as `/authoring-gzcli-ctf-events` if you want to load it explicitly; Claude will also trigger it automatically when the request matches its description.

## Event Model
- `.gzevent` is the event metadata source of truth.
- Category directories are fixed and case-sensitive:
  `Misc`, `Crypto`, `Pwn`, `Web`, `Reverse`, `Blockchain`, `Forensics`, `Hardware`, `Mobile`, `PPC`, `OSINT`, `Game Hacking`, `AI`, `Pentest`.
- `.example/` contains starter challenge templates.
- `.structure/` is the scaffold copied by `gzcli structure` into challenge folders.
- `.claude/skills/` contains project-local Claude Code skills for working in this event repo.
- `.agents/skills/` contains the shared skill references plus Codex/OpenAI-flavored manifests.

## CTF Workflow
1. Collect challenge context first: category, prompt, files or endpoint, flag format, and rules.
2. Build one intended solve path and keep one fallback.
3. Validate from a clean start to flag retrieval.
4. Keep solver material and writeups reproducible and technically falsifiable.

## Safety
Only provide exploitation guidance for authorized CTF, lab, or challenge environments.
