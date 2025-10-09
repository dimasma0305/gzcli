<!-- 70fc0bfa-e8d1-4673-97f5-adc9bac95f27 f11e89f3-1288-4106-b750-b79104a8caff -->
# Optimize Sync Process - Reduce API Calls

## Problem Analysis

From the terminal output, the sync process makes redundant API calls:
- `GET /api/edit/games` called **twice** (lines 10, 17)
- Challenge fetching makes 2 API calls per challenge via `GetChallenge()`
- Attachment processing creates files before checking if hash matches

## Optimization Strategy

### 1. Eliminate Duplicate Game List Fetching

**Current flow:**
```
InitWithEvent() → GetConfigWithEvent() → validateCachedGameWithKey() → api.GetGames()  [CALL 1]
Sync() → api.GetGames()  [CALL 2 - Same data!]
```

**Solution:** Pass validated game list from config loading to sync
- Modify `GetConfigWithEvent()` to return games list alongside config
- Pass games list through `gz.Sync()` to avoid re-fetching
- Update struct/function signatures to support this

**Files to modify:**
- `internal/gzcli/config/config.go` - Return games from validation
- `internal/gzcli/gzcli.go` - Accept and use games list in Sync()

### 2. Optimize Challenge Fetching in handleExistingChallenge

**Current behavior** (line 160-165 in sync.go):
```go
// Cache miss - calls GetChallenge() which makes 2 API calls:
//   1. GET /api/edit/games/{id}/challenges (list all)
//   2. GET /api/edit/games/{id}/challenges/{id} (get specific)
challengeData, err = config.Event.GetChallenge(challengeConf.Name)
```

**Solution:** Use already-fetched challenges list
- Add challenges list parameter to `handleExistingChallenge()`
- Find challenge in memory first
- Only call API as fallback if not found
- Reduces API calls from 2 to 0 for cache misses

**Files to modify:**
- `internal/gzcli/challenge/sync.go` - Pass challenges list, search in-memory first

### 3. Early Hash Check for Attachments

**Current flow** (attachment.go):
- Create zip file
- Create unique file
- Calculate hash
- **Then** check if hash matches (line 108)

**Optimization:** Check existing hash before file operations
- If challenge has attachment, get its hash from URL
- Fetch file list from API to get hash-to-file mapping
- Compare hashes early
- Skip file creation if hash will match

**Files to modify:**
- `internal/gzcli/challenge/attachment.go` - Add early hash comparison

### 4. Condense Logging

**Current:** 15+ log lines for unchanged attachment (lines 37-52)

**Condensed approach:**
- Single log for "Processing attachment" with outcome
- Detailed steps only on actual changes
- Group related operations: "Created zip, uploaded, hash: xxx"

**Files to modify:**
- `internal/gzcli/challenge/attachment.go` - Reduce log verbosity
- `internal/gzcli/challenge/sync.go` - Consolidate operation logs

## Implementation Order

1. **Pass games list through sync** - Eliminates 1 duplicate API call
2. **Optimize challenge fetching** - Eliminates 2N API calls (N = cached challenges)
3. **Early attachment hash check** - Reduces file I/O overhead
4. **Condense logging** - Improves output readability

## Expected Impact

- **API calls reduced by:** ~40-60% (1 game list + 2 per existing challenge)
- **File operations reduced:** Skip zip creation when attachment unchanged
- **Log output:** ~70% reduction in lines for unchanged resources
- **Sync speed:** 20-30% faster for incremental syncs

### To-dos

- [ ] Modify GetConfigWithEvent to return games list and pass it through Sync() to eliminate duplicate API call
- [ ] Update handleExistingChallenge to search in-memory challenges list before calling API
- [ ] Add early hash comparison in attachment handling to skip file operations when unchanged
- [ ] Reduce verbose logging in attachment and sync operations while keeping essential information