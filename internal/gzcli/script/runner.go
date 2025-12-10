// Package script provides utilities for running challenge scripts concurrently.
//
// This package handles the execution of build, test, and deployment scripts
// across multiple challenges using a worker pool for optimal performance.
//
// Example usage:
//
//	challenges := []ChallengeConf{challenge1, challenge2, challenge3}
//
//	runScript := func(conf ChallengeConf, scriptName string) error {
//	    cmd := exec.Command("sh", "-c", conf.GetScripts()[scriptName].GetCommand())
//	    return cmd.Run()
//	}
//
//	if err := script.RunScripts("build", challenges, runScript); err != nil {
//	    log.Fatalf("Script execution failed: %v", err)
//	}
package script

import (
	"fmt"
	"sync"
)

const maxParallelScripts = 10

// Failure captures a failed script execution for a specific challenge.
type Failure struct {
	Challenge string
	Err       error
}

// ChallengeConf interface for accessing challenge configuration
type ChallengeConf interface {
	GetName() string
	GetScripts() map[string]ScriptValue
}

// ScriptValue interface for accessing script values
//
//nolint:revive // Name ScriptValue is kept for backward compatibility
type ScriptValue interface {
	GetCommand() string
}

// RunScripts executes scripts with a worker pool and returns all failures (it does not stop at the first error).
func RunScripts(script string, challengesConf []ChallengeConf, runScript func(ChallengeConf, string) error) ([]Failure, error) {
	workChan := make(chan ChallengeConf, len(challengesConf))
	failChan := make(chan Failure, len(challengesConf))
	var wg sync.WaitGroup

	for i := 0; i < maxParallelScripts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for challengeConf := range workChan {
				if err := runScript(challengeConf, script); err != nil {
					failChan <- Failure{
						Challenge: challengeConf.GetName(),
						Err:       fmt.Errorf("script error in %s: %w", challengeConf.GetName(), err),
					}
				}
			}
		}()
	}

	for _, conf := range challengesConf {
		scripts := conf.GetScripts()
		if scriptValue, ok := scripts[script]; ok && scriptValue.GetCommand() != "" {
			workChan <- conf
		}
	}
	close(workChan)
	wg.Wait()
	close(failChan)

	failures := make([]Failure, 0, len(failChan))
	for f := range failChan {
		failures = append(failures, f)
	}

	if len(failures) > 0 {
		return failures, fmt.Errorf("%d script failures", len(failures))
	}
	return nil, nil
}
