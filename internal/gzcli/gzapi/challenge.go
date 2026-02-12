//nolint:revive // Challenge struct field names match API responses
package gzapi

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type gameChallengeCache struct {
	mu     sync.RWMutex
	byGame map[string]*gameChallengeCacheEntry
}

type gameChallengeCacheEntry struct {
	byTitle map[string]Challenge
	byID    map[int]string
}

var challengeCache = &gameChallengeCache{
	byGame: make(map[string]*gameChallengeCacheEntry),
}

func challengeCacheScope(gameID int, api *GZAPI) string {
	if api != nil && api.Creds != nil {
		return fmt.Sprintf("%s|%s|%d", strings.TrimSpace(api.Url), api.Creds.Username, gameID)
	}
	if api != nil {
		return fmt.Sprintf("%s|%d", strings.TrimSpace(api.Url), gameID)
	}
	return fmt.Sprintf("game|%d", gameID)
}

func (c *gameChallengeCache) setGameChallenges(gameID int, api *GZAPI, challenges []Challenge) {
	c.mu.Lock()
	defer c.mu.Unlock()
	scope := challengeCacheScope(gameID, api)

	entry := &gameChallengeCacheEntry{
		byTitle: make(map[string]Challenge, len(challenges)),
		byID:    make(map[int]string, len(challenges)),
	}
	for i := range challenges {
		ch := challenges[i]
		entry.byTitle[ch.Title] = ch
		entry.byID[ch.Id] = ch.Title
	}
	c.byGame[scope] = entry
}

func (c *gameChallengeCache) upsertChallenge(gameID int, api *GZAPI, challenge Challenge) {
	c.mu.Lock()
	defer c.mu.Unlock()
	scope := challengeCacheScope(gameID, api)

	entry, ok := c.byGame[scope]
	if !ok {
		entry = &gameChallengeCacheEntry{
			byTitle: make(map[string]Challenge),
			byID:    make(map[int]string),
		}
		c.byGame[scope] = entry
	}

	if oldTitle, ok := entry.byID[challenge.Id]; ok && oldTitle != challenge.Title {
		delete(entry.byTitle, oldTitle)
	}

	entry.byTitle[challenge.Title] = challenge
	entry.byID[challenge.Id] = challenge.Title
}

func (c *gameChallengeCache) getByTitle(gameID int, api *GZAPI, title string) (Challenge, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	scope := challengeCacheScope(gameID, api)

	entry, ok := c.byGame[scope]
	if !ok {
		return Challenge{}, false
	}
	ch, ok := entry.byTitle[title]
	return ch, ok
}

func (c *gameChallengeCache) deleteByID(gameID int, api *GZAPI, challengeID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	scope := challengeCacheScope(gameID, api)

	entry, ok := c.byGame[scope]
	if !ok {
		return
	}

	if title, ok := entry.byID[challengeID]; ok {
		delete(entry.byID, challengeID)
		delete(entry.byTitle, title)
	}
}

type Challenge struct {
	Id                   int         `json:"id" yaml:"id"`
	Title                string      `json:"title" yaml:"title"`
	Content              string      `json:"content" yaml:"content"`
	Category             string      `json:"category" yaml:"category"`
	Type                 string      `json:"type" yaml:"type"`
	Hints                []string    `json:"hints" yaml:"hints"`
	FlagTemplate         string      `json:"flagTemplate" yaml:"flagTemplate"`
	IsEnabled            *bool       `json:"isEnabled,omitempty" yaml:"isEnabled,omitempty"`
	AcceptedCount        int         `json:"acceptedCount" yaml:"acceptedCount"`
	FileName             string      `json:"fileName" yaml:"fileName"`
	Attachment           *Attachment `json:"attachment" yaml:"attachment"`
	TestContainer        interface{} `json:"testContainer" yaml:"testContainer"`
	Flags                []Flag      `json:"flags" yaml:"flags"`
	ContainerImage       string      `json:"containerImage" yaml:"containerImage"`
	MemoryLimit          int         `json:"memoryLimit" yaml:"memoryLimit"`
	CpuCount             int         `json:"cpuCount" yaml:"cpuCount"`
	StorageLimit         int         `json:"storageLimit" yaml:"storageLimit"`
	ContainerExposePort  int         `json:"exposePort" yaml:"exposePort"`
	NetworkMode          string      `json:"networkMode" yaml:"networkMode"`
	EnableTrafficCapture bool        `json:"enableTrafficCapture" yaml:"enableTrafficCapture"`
	DisableBloodBonus    bool        `json:"disableBloodBonus" yaml:"disableBloodBonus"`
	DeadlineUtc          int64       `json:"deadlineUtc" yaml:"deadlineUtc"`
	SubmissionLimit      int         `json:"submissionLimit" yaml:"submissionLimit"`
	OriginalScore        int         `json:"originalScore" yaml:"originalScore"`
	MinScoreRate         float64     `json:"minScoreRate" yaml:"minScoreRate"`
	Difficulty           float64     `json:"difficulty" yaml:"difficulty"`
	GameId               int         `json:"-" yaml:"gameId"`
	CS                   *GZAPI      `json:"-" yaml:"-"`
}

func (c *Challenge) Delete() error {
	if c.CS == nil {
		return fmt.Errorf("GZAPI client is not initialized")
	}
	if err := c.CS.delete(fmt.Sprintf("/api/edit/games/%d/challenges/%d", c.GameId, c.Id), nil); err != nil {
		return err
	}
	challengeCache.deleteByID(c.GameId, c.CS, c.Id)
	return nil
}

func (c *Challenge) Update(challenge Challenge) (*Challenge, error) {
	if c.CS == nil {
		return nil, fmt.Errorf("GZAPI client is not initialized")
	}
	if err := c.CS.put(fmt.Sprintf("/api/edit/games/%d/challenges/%d", c.GameId, c.Id), &challenge, nil); err != nil {
		return nil, err
	}
	challenge.GameId = c.GameId
	challenge.CS = c.CS
	challengeCache.upsertChallenge(c.GameId, c.CS, challenge)
	return &challenge, nil
}

func (c *Challenge) Refresh() (*Challenge, error) {
	if c.CS == nil {
		return nil, fmt.Errorf("GZAPI client is not initialized")
	}
	var data Challenge
	if err := c.CS.get(fmt.Sprintf("/api/edit/games/%d/challenges/%d", c.GameId, c.Id), &data); err != nil {
		return nil, err
	}
	data.GameId = c.GameId
	data.CS = c.CS
	challengeCache.upsertChallenge(c.GameId, c.CS, data)
	return &data, nil
}

type CreateChallengeForm struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Tag      string `json:"tag"`
	Type     string `json:"type"`
}

func (g *Game) CreateChallenge(challenge CreateChallengeForm) (*Challenge, error) {
	if g.CS == nil {
		return nil, fmt.Errorf("GZAPI client is not initialized")
	}

	var data *Challenge
	if err := g.CS.post(fmt.Sprintf("/api/edit/games/%d/challenges", g.Id), challenge, &data); err != nil {
		return nil, err
	}
	data.GameId = g.Id
	data.CS = g.CS
	challengeCache.upsertChallenge(g.Id, g.CS, *data)
	return data, nil
}

func (g *Game) GetChallenges() ([]Challenge, error) {
	if g.CS == nil {
		return nil, fmt.Errorf("GZAPI client is not initialized")
	}

	var tmp []Challenge
	var data []Challenge
	if err := g.CS.get(fmt.Sprintf("/api/edit/games/%d/challenges", g.Id), &tmp); err != nil {
		return nil, err
	}
	if len(tmp) == 0 {
		return data, nil
	}

	workers := resolveChallengeFetchWorkers(len(tmp))
	var wg sync.WaitGroup
	jobs := make(chan int, len(tmp))
	errs := make([]error, len(tmp))
	details := make([]Challenge, len(tmp))
	ok := make([]bool, len(tmp))

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				var c Challenge
				if err := g.CS.get(fmt.Sprintf("/api/edit/games/%d/challenges/%d", g.Id, tmp[idx].Id), &c); err != nil {
					errs[idx] = fmt.Errorf("fetch challenge id %d: %w", tmp[idx].Id, err)
					continue
				}
				c.GameId = g.Id
				c.CS = g.CS
				details[idx] = c
				ok[idx] = true
			}
		}()
	}

	for i := range tmp {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	for i := range errs {
		if errs[i] != nil {
			return nil, errs[i]
		}
	}

	data = make([]Challenge, 0, len(tmp))
	for i := range details {
		if ok[i] {
			data = append(data, details[i])
		}
	}

	challengeCache.setGameChallenges(g.Id, g.CS, data)
	return data, nil
}

func resolveChallengeFetchWorkers(total int) int {
	if total <= 0 {
		return 1
	}

	workers := 6
	if workers > total {
		workers = total
	}

	if raw := strings.TrimSpace(os.Getenv("GZCLI_GET_CHALLENGES_WORKERS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			workers = parsed
		}
	}

	if workers > total {
		workers = total
	}
	if workers < 1 {
		return 1
	}
	return workers
}

func (g *Game) GetChallenge(name string) (*Challenge, error) {
	if cached, ok := challengeCache.getByTitle(g.Id, g.CS, name); ok {
		cached.GameId = g.Id
		cached.CS = g.CS
		return &cached, nil
	}

	var data []Challenge
	if err := g.CS.get(fmt.Sprintf("/api/edit/games/%d/challenges", g.Id), &data); err != nil {
		return nil, err
	}
	var challenge *Challenge
	for _, v := range data {
		if v.Title == name {
			challenge = &v
		}
	}
	if challenge == nil {
		return nil, fmt.Errorf("challenge not found")
	}
	if err := g.CS.get(fmt.Sprintf("/api/edit/games/%d/challenges/%d", g.Id, challenge.Id), &challenge); err != nil {
		return nil, err
	}
	challenge.GameId = g.Id
	challenge.CS = g.CS
	challengeCache.upsertChallenge(g.Id, g.CS, *challenge)
	return challenge, nil
}
