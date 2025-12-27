package server

import (
	"testing"
	"time"
)

func TestVotingManager_StartVote(t *testing.T) {
	vm := NewVotingManager()
	slug := "test_challenge"

	// Test starting a vote
	err := vm.StartVote(slug, nil)
	if err != nil {
		t.Errorf("Failed to start vote: %v", err)
	}

	// Test starting duplicate vote
	err = vm.StartVote(slug, nil)
	if err == nil {
		t.Error("Expected error when starting duplicate vote, got nil")
	}

	// Verify vote exists
	if !vm.HasActiveVote(slug) {
		t.Error("Vote should exist after starting")
	}
}

func TestVotingManager_CastVote(t *testing.T) {
	vm := NewVotingManager()
	slug := "test_challenge"

	// Start a vote
	_ = vm.StartVote(slug, nil)

	// Cast yes vote
	err := vm.CastVote(slug, "192.168.1.1", true)
	if err != nil {
		t.Errorf("Failed to cast vote: %v", err)
	}

	// Cast no vote
	err = vm.CastVote(slug, "192.168.1.2", false)
	if err != nil {
		t.Errorf("Failed to cast vote: %v", err)
	}

	// Update existing vote
	err = vm.CastVote(slug, "192.168.1.1", false)
	if err != nil {
		t.Errorf("Failed to update vote: %v", err)
	}

	// Try to vote on non-existent challenge
	err = vm.CastVote("nonexistent", "192.168.1.1", true)
	if err == nil {
		t.Error("Expected error when voting on non-existent challenge")
	}
}

func TestVotingManager_GetVoteStatus(t *testing.T) {
	vm := NewVotingManager()
	slug := "test_challenge"

	// Start a vote
	_ = vm.StartVote(slug, nil)

	// Cast votes
	_ = vm.CastVote(slug, "192.168.1.1", true)
	_ = vm.CastVote(slug, "192.168.1.2", true)
	_ = vm.CastVote(slug, "192.168.1.3", false)
	_ = vm.CastVote(slug, "192.168.1.4", false)

	// All 4 IPs are connected
	connectedIPs := map[string]bool{
		"192.168.1.1": true,
		"192.168.1.2": true,
		"192.168.1.3": true,
		"192.168.1.4": true,
	}

	yesPercent, noPercent, totalVoters, exists := vm.GetVoteStatus(slug, connectedIPs)

	if !exists {
		t.Error("Vote should exist")
	}

	if totalVoters != 4 {
		t.Errorf("Expected 4 total voters, got %d", totalVoters)
	}

	if yesPercent != 50.0 {
		t.Errorf("Expected 50%% yes votes, got %.2f%%", yesPercent)
	}

	if noPercent != 50.0 {
		t.Errorf("Expected 50%% no votes, got %.2f%%", noPercent)
	}
}

func TestVotingManager_CheckThreshold(t *testing.T) {
	vm := NewVotingManager()
	slug := "test_challenge"

	// Start a vote
	_ = vm.StartVote(slug, nil)

	connectedIPs := map[string]bool{
		"192.168.1.1": true,
		"192.168.1.2": true,
		"192.168.1.3": true,
		"192.168.1.4": true,
	}

	// Test: Not enough votes (in progress)
	approved, rejected, inProgress := vm.CheckThreshold(slug, connectedIPs)
	if approved || rejected || !inProgress {
		t.Error("Expected in progress with no votes")
	}

	// Cast 2 yes votes (50% threshold)
	_ = vm.CastVote(slug, "192.168.1.1", true)
	_ = vm.CastVote(slug, "192.168.1.2", true)

	// Test: Threshold reached (approved)
	approved, rejected, inProgress = vm.CheckThreshold(slug, connectedIPs)
	if !approved || rejected || inProgress {
		t.Error("Expected approved with 50% yes votes")
	}

	// Start new vote
	vm.EndVote(slug, "test")
	_ = vm.StartVote(slug, nil)

	// Cast 2 no votes (50% threshold)
	_ = vm.CastVote(slug, "192.168.1.1", false)
	_ = vm.CastVote(slug, "192.168.1.2", false)

	// Test: Threshold reached (rejected)
	approved, rejected, inProgress = vm.CheckThreshold(slug, connectedIPs)
	if approved || !rejected || inProgress {
		t.Error("Expected rejected with 50% no votes")
	}
}

func TestVotingManager_EndVote(t *testing.T) {
	vm := NewVotingManager()
	slug := "test_challenge"

	// Start a vote
	_ = vm.StartVote(slug, nil)

	if !vm.HasActiveVote(slug) {
		t.Error("Vote should exist")
	}

	// End the vote
	vm.EndVote(slug, "test")

	if vm.HasActiveVote(slug) {
		t.Error("Vote should not exist after ending")
	}
}

func TestVotingManager_GetVoteAge(t *testing.T) {
	vm := NewVotingManager()
	slug := "test_challenge"

	// Test non-existent vote
	_, exists := vm.GetVoteAge(slug)
	if exists {
		t.Error("Expected no vote age for non-existent vote")
	}

	// Start a vote
	_ = vm.StartVote(slug, nil)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	age, exists := vm.GetVoteAge(slug)
	if !exists {
		t.Error("Expected vote age to exist")
	}

	if age < 100*time.Millisecond {
		t.Errorf("Expected vote age >= 100ms, got %v", age)
	}
}

func TestVotingManager_OnlyCountConnectedUsers(t *testing.T) {
	vm := NewVotingManager()
	slug := "test_challenge"

	// Start a vote
	_ = vm.StartVote(slug, nil)

	// Cast votes from 4 IPs
	_ = vm.CastVote(slug, "192.168.1.1", true)
	_ = vm.CastVote(slug, "192.168.1.2", true)
	_ = vm.CastVote(slug, "192.168.1.3", false)
	_ = vm.CastVote(slug, "192.168.1.4", false)

	// Only 2 IPs are connected
	connectedIPs := map[string]bool{
		"192.168.1.1": true,
		"192.168.1.2": true,
	}

	yesPercent, noPercent, totalVoters, exists := vm.GetVoteStatus(slug, connectedIPs)

	if !exists {
		t.Error("Vote should exist")
	}

	if totalVoters != 2 {
		t.Errorf("Expected 2 total voters (connected), got %d", totalVoters)
	}

	// Both connected users voted yes, so 100% yes
	if yesPercent != 100.0 {
		t.Errorf("Expected 100%% yes votes (from connected users), got %.2f%%", yesPercent)
	}

	if noPercent != 0.0 {
		t.Errorf("Expected 0%% no votes (from connected users), got %.2f%%", noPercent)
	}
}
