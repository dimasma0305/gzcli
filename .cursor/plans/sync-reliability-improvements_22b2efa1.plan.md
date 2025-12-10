---
name: sync-reliability-improvements
overview: Quick, high-level fixes to improve challenge sync reliability (attachments, flags, dedupe, cache).
todos:
  - id: fresh-context
    content: Use fresh remote challenge data before attachments/flags
    status: pending
  - id: attachment-retry
    content: Improve attachment 404 handling/log context
    status: pending
  - id: tests-stale-cache
    content: Add tests for stale cache vs live data path
    status: pending
  - id: refetch-after-dedupe
    content: Refetch challenges after duplicate deletion
    status: pending
  - id: flags-refresh
    content: Refresh challenge after flag mutations
    status: pending
---

# Sync Reliability Improvements

- Ensure remote challenge context is always fresh before attachments/flags, preferring live data over stale cache; verify CS/GameId propagation. Key file: `internal/gzcli/challenge/sync.go`.
- Harden attachment creation with better 404 handling: retry fetch/update challenge metadata when POST /attachment fails, and log request context (game/challenge IDs, type). Files: `internal/gzcli/challenge/attachment.go`, `internal/gzcli/challenge/sync.go`.
- Add unit coverage for the attachment/flag path using fresh remote data and simulating stale cache to prevent regressions. File: `internal/gzcli/challenge/sync_test.go`.
- Validate remote list after duplicate cleanup to ensure we don’t reuse deleted IDs; re-fetch challenges when deletions occur. Files: `internal/gzcli/challenge/sync.go`, `internal/gzcli/gzcli.go`.
- Add guardrails for flags to refresh challenge state after flag mutations, similar to attachments, to avoid stale IDs. File: `internal/gzcli/challenge/flag.go`.