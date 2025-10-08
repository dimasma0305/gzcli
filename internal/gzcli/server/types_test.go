package server

import (
	"testing"
	"time"
)

func TestChallengeInfo_ConnectedUsers(t *testing.T) {
	challenge := &ChallengeInfo{
		Slug:         "test_web_challenge",
		Name:         "Test Challenge",
		Status:       StatusStopped,
		ConnectedIPs: make(map[string]bool),
	}

	// Test initial state
	if challenge.GetConnectedUsers() != 0 {
		t.Errorf("Expected 0 connected users, got %d", challenge.GetConnectedUsers())
	}

	// Add users
	challenge.AddConnectedIP("192.168.1.1")
	challenge.AddConnectedIP("192.168.1.2")
	challenge.AddConnectedIP("192.168.1.3")

	if challenge.GetConnectedUsers() != 3 {
		t.Errorf("Expected 3 connected users, got %d", challenge.GetConnectedUsers())
	}

	// Add duplicate (should not increase count)
	challenge.AddConnectedIP("192.168.1.1")

	if challenge.GetConnectedUsers() != 3 {
		t.Errorf("Expected 3 connected users after duplicate, got %d", challenge.GetConnectedUsers())
	}

	// Remove user
	challenge.RemoveConnectedIP("192.168.1.2")

	if challenge.GetConnectedUsers() != 2 {
		t.Errorf("Expected 2 connected users after removal, got %d", challenge.GetConnectedUsers())
	}
}

func TestChallengeInfo_Status(t *testing.T) {
	challenge := &ChallengeInfo{
		Slug:   "test_web_challenge",
		Name:   "Test Challenge",
		Status: StatusStopped,
	}

	// Test get status
	if challenge.GetStatus() != StatusStopped {
		t.Errorf("Expected status stopped, got %s", challenge.GetStatus())
	}

	// Test set status
	challenge.SetStatus(StatusStarting)
	if challenge.GetStatus() != StatusStarting {
		t.Errorf("Expected status starting, got %s", challenge.GetStatus())
	}

	challenge.SetStatus(StatusRunning)
	if challenge.GetStatus() != StatusRunning {
		t.Errorf("Expected status running, got %s", challenge.GetStatus())
	}
}

func TestChallengeInfo_CooldownCheck(t *testing.T) {
	challenge := &ChallengeInfo{
		Slug:        "test_web_challenge",
		Name:        "Test Challenge",
		LastRestart: time.Now().Add(-6 * time.Minute), // 6 minutes ago
	}

	// Should not be in cooldown (6 minutes > 5 minutes)
	inCooldown, remaining := challenge.IsInCooldown()
	if inCooldown {
		t.Error("Expected not in cooldown after 6 minutes")
	}
	if remaining != 0 {
		t.Errorf("Expected 0 remaining time, got %v", remaining)
	}

	// Set recent restart
	challenge.SetLastRestart(time.Now().Add(-2 * time.Minute)) // 2 minutes ago

	// Should be in cooldown (2 minutes < 5 minutes)
	inCooldown, remaining = challenge.IsInCooldown()
	if !inCooldown {
		t.Error("Expected in cooldown after 2 minutes")
	}
	if remaining <= 0 || remaining > 5*time.Minute {
		t.Errorf("Expected remaining time between 0 and 5 minutes, got %v", remaining)
	}
}

func TestChallengeInfo_GracePeriod(t *testing.T) {
	challenge := &ChallengeInfo{
		Slug: "test_web_challenge",
		Name: "Test Challenge",
	}

	// Grace period is now fixed at 2 minutes
	gracePeriod := challenge.CalculateGracePeriod()
	expected := 2 * time.Minute

	if gracePeriod != expected {
		t.Errorf("Expected grace period %v, got %v", expected, gracePeriod)
	}
}
