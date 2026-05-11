# Event Repository Structure

This template is copied into `events/<event-name>/`.

Expected event-root files and directories:
- `.gzevent`: event metadata such as title, schedule, visibility, summary, and writeup policy
- `.example/`: starter challenge templates for common deployment models
- `.structure/`: reusable scaffold copied into challenge folders by `gzcli structure`
- `.agents/skills/`: shared skill references plus Codex/OpenAI-flavored manifests
- `.claude/skills/`: project-local Claude Code skills (slash-invocable as `/authoring-gzcli-ctf-events`)
- `AGENTS.md`: auto-loaded by Codex; points at `.agents/skills/`
- `CLAUDE.md`: auto-loaded by Claude Code; points at `.claude/skills/`
- `README.md`: participant and author guidance
- `Kriteria.md`: organizer challenge criteria
- Category directories:
  - `Misc`
  - `Crypto`
  - `Pwn`
  - `Web`
  - `Reverse`
  - `Blockchain`
  - `Forensics`
  - `Hardware`
  - `Mobile`
  - `PPC`
  - `OSINT`
  - `Game Hacking`
  - `AI`
  - `Pentest`

Use one folder per challenge inside the correct category:

```text
<Category>/<challenge-slug>/
```

When `gzcli structure` is used, `.structure/` content is copied into challenge directories. That scaffold may add repo-facing documentation like `README.md`, `writeups/`, or helper placeholders for maintainers and contributors.
