# Challenge Directory Contract

There are two useful views of a challenge directory in a gzcli event repo.

## 1. Working Tree View

Inside the event repository, a challenge folder may contain:
- `challenge.yml`
- `dist/`
- `src/`
- `solver/`
- optional infrastructure files like `Dockerfile` and `docker-compose.yml`
- generated maintainer or contributor docs copied from `.structure/`

This is the view maintainers work with during normal authoring.

## 2. Portable Upload View

If a challenge is being prepared for the upload server or archive validation path, keep the root compatible with the embedded starter templates:
- required:
  - `challenge.yml`
  - `dist/`
  - `src/`
  - `solver/`
- optional:
  - `Dockerfile`
  - `docker-compose.yml`
  - `.dockerignore`
  - `.gitignore`

Additional root entries may fail upload validation.

## Authoring Guidance
- Start from the closest `.example/` template.
- Rename the folder to the final challenge slug early.
- Replace placeholder metadata in `challenge.yml` immediately.
- Keep downloadable player artifacts in `dist/` — place raw files directly, **do not zip them**; gzcli automatically zips `dist/` contents when uploading/syncing.
- Keep service code and build inputs in `src/`.
- Keep the intended solution or solver in `solver/`.
