package server

import (
	"sync"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

const (
	healthCheckInterval = 30 * time.Second
)

// HealthMonitor monitors challenge health
type HealthMonitor struct {
	challenges *ChallengeManager
	executor   *Executor
	wsManager  *WSManager
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(challenges *ChallengeManager, executor *Executor, wsManager *WSManager) *HealthMonitor {
	return &HealthMonitor{
		challenges: challenges,
		executor:   executor,
		wsManager:  wsManager,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the health monitoring loop
func (hm *HealthMonitor) Start() {
	hm.wg.Add(1)
	go hm.monitorLoop()
	log.Info("Health monitor started")
}

// Stop stops the health monitoring loop
func (hm *HealthMonitor) Stop() {
	close(hm.stopChan)
	hm.wg.Wait()
	log.Info("Health monitor stopped")
}

// monitorLoop is the main monitoring loop
func (hm *HealthMonitor) monitorLoop() {
	defer hm.wg.Done()

	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.performHealthChecks()
		}
	}
}

// performHealthChecks checks the health of all challenges
func (hm *HealthMonitor) performHealthChecks() {
	challenges := hm.challenges.ListChallenges()

	for _, challenge := range challenges {
		// Only check challenges that should be running
		status := challenge.GetStatus()
		if status != StatusRunning {
			continue
		}

		// Perform health check
		isHealthy, err := hm.executor.CheckHealth(challenge)
		if err != nil {
			log.Error("Health check error for %s: %v", challenge.Name, err)
			continue
		}

		// If not healthy, update status
		if !isHealthy {
			log.Error("Challenge %s is unhealthy (expected running, but not found)", challenge.Name)
			challenge.SetStatus(StatusUnhealthy)

			// Broadcast status update
			if hm.wsManager != nil {
				hm.wsManager.broadcastStatus(challenge.Slug)
				hm.wsManager.broadcastError(challenge.Slug, "Challenge is unhealthy. Please restart.")
			}
		}
	}
}
