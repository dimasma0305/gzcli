package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

// Voting configuration constants
const (
	// VoteTimeout is the duration after which a vote expires
	VoteTimeout = 15 * time.Second
	// VoteThreshold is the minimum percentage of votes needed to approve an action
	VoteThreshold = 0.5 // 50%
)

// VotingManager manages restart votes for challenges
type VotingManager struct {
	votes map[string]*Vote // challenge slug -> Vote
	mu    sync.RWMutex
}

// NewVotingManager creates a new voting manager
func NewVotingManager() *VotingManager {
	return &VotingManager{
		votes: make(map[string]*Vote),
	}
}

// StartVote starts a new restart vote for a challenge
// onTimeout is called when the vote expires
func (vm *VotingManager) StartVote(slug string, onTimeout func()) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Check if vote already exists
	if _, exists := vm.votes[slug]; exists {
		return fmt.Errorf("vote already in progress")
	}

	// Create new vote
	vote := &Vote{
		InitiatedAt: time.Now(),
		Votes:       make(map[string]bool),
	}

	vm.votes[slug] = vote

	log.InfoH2("Restart vote started for challenge: %s", slug)

	// Start timeout timer
	go func() {
		time.Sleep(VoteTimeout)
		if onTimeout != nil {
			onTimeout()
		}
	}()

	return nil
}

// CastVote casts a vote (yes=true, no=false) from an IP
func (vm *VotingManager) CastVote(slug, ip string, voteYes bool) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vote, exists := vm.votes[slug]
	if !exists {
		return fmt.Errorf("no active vote for this challenge")
	}

	vote.mu.Lock()
	defer vote.mu.Unlock()

	// Check if IP already voted
	if _, hasVoted := vote.Votes[ip]; hasVoted {
		// Update their vote
		vote.Votes[ip] = voteYes
		log.InfoH3("Vote updated from IP: %s (vote: %v)", maskIP(ip), voteYes)
	} else {
		// New vote
		vote.Votes[ip] = voteYes
		log.InfoH3("Vote cast from IP: %s (vote: %v)", maskIP(ip), voteYes)
	}

	return nil
}

// GetVoteStatus returns the current vote status
func (vm *VotingManager) GetVoteStatus(slug string, connectedIPs map[string]bool) (yesPercent, noPercent float64, totalVoters int, exists bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	vote, exists := vm.votes[slug]
	if !exists {
		return 0, 0, 0, false
	}

	vote.mu.RLock()
	defer vote.mu.RUnlock()

	// Count total unique connected IPs (potential voters)
	totalVoters = len(connectedIPs)
	if totalVoters == 0 {
		return 0, 0, 0, true
	}

	// Count yes and no votes
	yesVotes := 0
	noVotes := 0

	for ip, voteValue := range vote.Votes {
		// Only count votes from currently connected users
		if _, isConnected := connectedIPs[ip]; isConnected {
			if voteValue {
				yesVotes++
			} else {
				noVotes++
			}
		}
	}

	// Calculate percentages
	yesPercent = float64(yesVotes) / float64(totalVoters) * 100
	noPercent = float64(noVotes) / float64(totalVoters) * 100

	return yesPercent, noPercent, totalVoters, true
}

// CheckThreshold checks if the vote has reached the threshold
// Returns: (approved, rejected, inProgress)
func (vm *VotingManager) CheckThreshold(slug string, connectedIPs map[string]bool) (bool, bool, bool) {
	yesPercent, noPercent, _, exists := vm.GetVoteStatus(slug, connectedIPs)

	if !exists {
		return false, false, false
	}

	// Check if yes threshold reached
	if yesPercent >= VoteThreshold*100 {
		return true, false, false
	}

	// Check if no threshold reached
	if noPercent >= VoteThreshold*100 {
		return false, true, false
	}

	// Vote still in progress
	return false, false, true
}

// EndVote ends a vote and removes it from the manager
func (vm *VotingManager) EndVote(slug string, reason string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if _, exists := vm.votes[slug]; exists {
		delete(vm.votes, slug)
		log.InfoH2("Restart vote ended for challenge: %s (reason: %s)", slug, reason)
	}
}

// HasActiveVote checks if a challenge has an active vote
func (vm *VotingManager) HasActiveVote(slug string) bool {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	_, exists := vm.votes[slug]
	return exists
}

// GetVoteAge returns how long a vote has been active
func (vm *VotingManager) GetVoteAge(slug string) (time.Duration, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	vote, exists := vm.votes[slug]
	if !exists {
		return 0, false
	}

	vote.mu.RLock()
	defer vote.mu.RUnlock()

	return time.Since(vote.InitiatedAt), true
}

// maskIP partially masks an IP address for privacy
func maskIP(ip string) string {
	// For IPv4: 192.168.x.x
	// For IPv6: 2001:db8:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx
	if len(ip) > 8 {
		return ip[:len(ip)/2] + "x.x"
	}
	return "x.x.x.x"
}
