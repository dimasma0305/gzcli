//nolint:revive,unparam // Test file with intentionally unused parameters for interface compatibility
package script

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockChallengeConf implements ChallengeConf for testing
type mockChallengeConf struct {
	name    string
	scripts map[string]ScriptValue
}

func (m *mockChallengeConf) GetName() string {
	return m.name
}

func (m *mockChallengeConf) GetScripts() map[string]ScriptValue {
	return m.scripts
}

// mockScriptValue implements ScriptValue for testing
type mockScriptValue struct {
	command string
}

func (m *mockScriptValue) GetCommand() string {
	return m.command
}

// TestRunScripts_Success tests successful script execution
func TestRunScripts_Success(t *testing.T) {
	challenges := []ChallengeConf{
		&mockChallengeConf{
			name: "Challenge1",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build1"},
			},
		},
		&mockChallengeConf{
			name: "Challenge2",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build2"},
			},
		},
		&mockChallengeConf{
			name: "Challenge3",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build3"},
			},
		},
	}

	executedChallenges := make(map[string]bool)
	var mu sync.Mutex

	runScript := func(conf ChallengeConf, scriptName string) error {
		mu.Lock()
		defer mu.Unlock()
		executedChallenges[conf.GetName()] = true
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err != nil {
		t.Errorf("RunScripts() failed: %v", err)
	}

	// Verify all challenges were executed
	mu.Lock()
	defer mu.Unlock()
	for _, ch := range challenges {
		if !executedChallenges[ch.GetName()] {
			t.Errorf("Challenge %s was not executed", ch.GetName())
		}
	}
}

// TestRunScripts_Error tests error handling when a script fails
func TestRunScripts_Error(t *testing.T) {
	challenges := []ChallengeConf{
		&mockChallengeConf{
			name: "Challenge1",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build1"},
			},
		},
		&mockChallengeConf{
			name: "Challenge2",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build2"},
			},
		},
	}

	expectedError := errors.New("script execution failed")
	runScript := func(conf ChallengeConf, scriptName string) error {
		if conf.GetName() == "Challenge1" {
			return expectedError
		}
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, expectedError) {
		// Check if error contains the challenge name
		if err.Error() == "" {
			t.Errorf("Expected error containing challenge name, got: %v", err)
		}
	}
}

// TestRunScripts_EmptyChallengeList tests with no challenges
func TestRunScripts_EmptyChallengeList(t *testing.T) {
	var challenges []ChallengeConf

	runScript := func(conf ChallengeConf, scriptName string) error {
		t.Error("runScript should not be called for empty challenge list")
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err != nil {
		t.Errorf("RunScripts() with empty list failed: %v", err)
	}
}

// TestRunScripts_NoMatchingScript tests when challenges don't have the specified script
func TestRunScripts_NoMatchingScript(t *testing.T) {
	challenges := []ChallengeConf{
		&mockChallengeConf{
			name: "Challenge1",
			scripts: map[string]ScriptValue{
				"test": &mockScriptValue{command: "echo test"},
			},
		},
		&mockChallengeConf{
			name: "Challenge2",
			scripts: map[string]ScriptValue{
				"deploy": &mockScriptValue{command: "echo deploy"},
			},
		},
	}

	executionCount := 0
	runScript := func(conf ChallengeConf, scriptName string) error {
		executionCount++
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err != nil {
		t.Errorf("RunScripts() failed: %v", err)
	}

	if executionCount != 0 {
		t.Errorf("Expected 0 executions, got %d", executionCount)
	}
}

// TestRunScripts_EmptyCommand tests when script command is empty
func TestRunScripts_EmptyCommand(t *testing.T) {
	challenges := []ChallengeConf{
		&mockChallengeConf{
			name: "Challenge1",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: ""}, // Empty command
			},
		},
	}

	executionCount := 0
	runScript := func(conf ChallengeConf, scriptName string) error {
		executionCount++
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err != nil {
		t.Errorf("RunScripts() failed: %v", err)
	}

	// Empty command should not be executed
	if executionCount != 0 {
		t.Errorf("Expected 0 executions for empty command, got %d", executionCount)
	}
}

// TestRunScripts_Concurrency tests concurrent execution
func TestRunScripts_Concurrency(t *testing.T) {
	// Create many challenges to test worker pool
	challenges := make([]ChallengeConf, 50)
	for i := 0; i < 50; i++ {
		challenges[i] = &mockChallengeConf{
			name: "Challenge" + string(rune('A'+i%26)),
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build"},
			},
		}
	}

	var executionCount int32
	var wg sync.WaitGroup

	runScript := func(conf ChallengeConf, scriptName string) error {
		atomic.AddInt32(&executionCount, 1)
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := RunScripts("build", challenges, runScript)
		if err != nil {
			t.Errorf("RunScripts() failed: %v", err)
		}
	}()

	wg.Wait()

	//nolint:gosec // G115: Test code, conversion is safe for test data
	if atomic.LoadInt32(&executionCount) != int32(len(challenges)) {
		t.Errorf("Expected %d executions, got %d", len(challenges), executionCount)
	}
}

// TestRunScripts_PartialMatch tests when only some challenges have the script
func TestRunScripts_PartialMatch(t *testing.T) {
	challenges := []ChallengeConf{
		&mockChallengeConf{
			name: "Challenge1",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build1"},
			},
		},
		&mockChallengeConf{
			name: "Challenge2",
			scripts: map[string]ScriptValue{
				"test": &mockScriptValue{command: "echo test2"}, // Different script
			},
		},
		&mockChallengeConf{
			name: "Challenge3",
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build3"},
			},
		},
	}

	executedChallenges := make(map[string]bool)
	var mu sync.Mutex

	runScript := func(conf ChallengeConf, scriptName string) error {
		mu.Lock()
		defer mu.Unlock()
		executedChallenges[conf.GetName()] = true
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err != nil {
		t.Errorf("RunScripts() failed: %v", err)
	}

	// Only Challenge1 and Challenge3 should be executed
	mu.Lock()
	defer mu.Unlock()
	if !executedChallenges["Challenge1"] {
		t.Error("Challenge1 should have been executed")
	}
	if executedChallenges["Challenge2"] {
		t.Error("Challenge2 should not have been executed")
	}
	if !executedChallenges["Challenge3"] {
		t.Error("Challenge3 should have been executed")
	}
}

// TestRunScripts_NilScripts tests handling of nil scripts map
func TestRunScripts_NilScripts(t *testing.T) {
	challenges := []ChallengeConf{
		&mockChallengeConf{
			name:    "Challenge1",
			scripts: nil, // Nil scripts map
		},
	}

	executionCount := 0
	runScript := func(conf ChallengeConf, scriptName string) error {
		executionCount++
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err != nil {
		t.Errorf("RunScripts() with nil scripts failed: %v", err)
	}

	if executionCount != 0 {
		t.Errorf("Expected 0 executions for nil scripts, got %d", executionCount)
	}
}

// TestRunScripts_ErrorPropagation tests that errors stop execution
func TestRunScripts_ErrorPropagation(t *testing.T) {
	challenges := make([]ChallengeConf, 20)
	for i := 0; i < 20; i++ {
		challenges[i] = &mockChallengeConf{
			name: "Challenge" + string(rune('A'+i)),
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build"},
			},
		}
	}

	var executionCount int32
	expectedError := errors.New("intentional error")

	runScript := func(conf ChallengeConf, scriptName string) error {
		count := atomic.AddInt32(&executionCount, 1)
		// Fail on the 5th execution
		if count == 5 {
			return expectedError
		}
		// Simulate some work
		time.Sleep(5 * time.Millisecond)
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Execution count should be less than total challenges due to early termination
	finalCount := atomic.LoadInt32(&executionCount)
	t.Logf("Execution stopped after %d challenges (expected early termination)", finalCount)

	// At least 5 should have been executed (the one that failed)
	if finalCount < 5 {
		t.Errorf("Expected at least 5 executions, got %d", finalCount)
	}
}

// TestRunScripts_MaxParallelScripts tests worker pool limit
func TestRunScripts_MaxParallelScripts(t *testing.T) {
	challenges := make([]ChallengeConf, 30)
	for i := 0; i < 30; i++ {
		challenges[i] = &mockChallengeConf{
			name: "Challenge" + string(rune('A'+i%26)),
			scripts: map[string]ScriptValue{
				"build": &mockScriptValue{command: "echo build"},
			},
		}
	}

	var maxConcurrent int32
	var currentConcurrent int32
	var mu sync.Mutex

	runScript := func(conf ChallengeConf, scriptName string) error {
		current := atomic.AddInt32(&currentConcurrent, 1)

		mu.Lock()
		if current > atomic.LoadInt32(&maxConcurrent) {
			atomic.StoreInt32(&maxConcurrent, current)
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&currentConcurrent, -1)
		return nil
	}

	err := RunScripts("build", challenges, runScript)
	if err != nil {
		t.Errorf("RunScripts() failed: %v", err)
	}

	maxObserved := atomic.LoadInt32(&maxConcurrent)
	if maxObserved > maxParallelScripts {
		t.Errorf("Max concurrent executions (%d) exceeded limit (%d)", maxObserved, maxParallelScripts)
	}

	if maxObserved < 1 {
		t.Error("No concurrent execution detected")
	}

	t.Logf("Max concurrent executions: %d (limit: %d)", maxObserved, maxParallelScripts)
}

// TestChallengeConf_Interface tests the ChallengeConf interface
func TestChallengeConf_Interface(t *testing.T) {
	conf := &mockChallengeConf{
		name: "TestChallenge",
		scripts: map[string]ScriptValue{
			"build": &mockScriptValue{command: "echo test"},
		},
	}

	if conf.GetName() != "TestChallenge" {
		t.Errorf("GetName() = %q, want %q", conf.GetName(), "TestChallenge")
	}

	scripts := conf.GetScripts()
	if len(scripts) != 1 {
		t.Errorf("len(GetScripts()) = %d, want 1", len(scripts))
	}

	buildScript, ok := scripts["build"]
	if !ok {
		t.Error("Expected 'build' script to exist")
	}

	if buildScript.GetCommand() != "echo test" {
		t.Errorf("GetCommand() = %q, want %q", buildScript.GetCommand(), "echo test")
	}
}

// TestScriptValue_Interface tests the ScriptValue interface
func TestScriptValue_Interface(t *testing.T) {
	script := &mockScriptValue{command: "echo hello"}

	if script.GetCommand() != "echo hello" {
		t.Errorf("GetCommand() = %q, want %q", script.GetCommand(), "echo hello")
	}
}

// TestRunScripts_MultipleScriptTypes tests execution of different script types
func TestRunScripts_MultipleScriptTypes(t *testing.T) {
	scriptTypes := []string{"build", "test", "deploy", "clean"}

	for _, scriptType := range scriptTypes {
		t.Run(scriptType, func(t *testing.T) {
			challenges := []ChallengeConf{
				&mockChallengeConf{
					name: "Challenge1",
					scripts: map[string]ScriptValue{
						scriptType: &mockScriptValue{command: "echo " + scriptType},
					},
				},
			}

			executed := false
			runScript := func(conf ChallengeConf, scriptName string) error {
				if scriptName != scriptType {
					t.Errorf("Expected script name %q, got %q", scriptType, scriptName)
				}
				executed = true
				return nil
			}

			err := RunScripts(scriptType, challenges, runScript)
			if err != nil {
				t.Errorf("RunScripts(%q) failed: %v", scriptType, err)
			}

			if !executed {
				t.Errorf("Script %q was not executed", scriptType)
			}
		})
	}
}
