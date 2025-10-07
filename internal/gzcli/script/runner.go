package script

import (
	"context"
	"fmt"
	"sync"
)

const maxParallelScripts = 10

// ChallengeConf interface for accessing challenge configuration
type ChallengeConf interface {
	GetName() string
	GetScripts() map[string]ScriptValue
}

// ScriptValue interface for accessing script values
type ScriptValue interface {
	GetCommand() string
}

// RunScripts executes scripts with a worker pool
func RunScripts(script string, challengesConf []ChallengeConf, runScript func(ChallengeConf, string) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workChan := make(chan ChallengeConf, len(challengesConf))
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Create worker pool
	for i := 0; i < maxParallelScripts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for challengeConf := range workChan {
				select {
				case <-ctx.Done():
					return
				default:
					if err := runScript(challengeConf, script); err != nil {
						select {
						case errChan <- fmt.Errorf("script error in %s: %w", challengeConf.GetName(), err):
							cancel()
						default:
						}
					}
				}
			}
		}()
	}

	// Distribute work
	for _, conf := range challengesConf {
		scripts := conf.GetScripts()
		if scriptValue, ok := scripts[script]; ok && scriptValue.GetCommand() != "" {
			workChan <- conf
		}
	}
	close(workChan)
	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
