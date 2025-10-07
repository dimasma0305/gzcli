package filesystem

import (
	"context"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
	"github.com/dimasma0305/gzcli/internal/log"
)

// WorkerPool manages concurrent processing of file system events
type WorkerPool struct {
	workers   int
	eventChan chan fsnotify.Event
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewWorkerPool creates a new worker pool for event processing
func NewWorkerPool(workers int) *WorkerPool {
	if workers <= 0 {
		workers = 4 // Default to 4 workers
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:   workers,
		eventChan: make(chan fsnotify.Event, workers*10), // Buffer for bursty events
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(handler EventHandler, config types.WatcherConfig) {
	log.InfoH3("Starting worker pool with %d workers", wp.workers)

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i, handler, config)
	}
}

// worker processes events from the event channel
func (wp *WorkerPool) worker(id int, handler EventHandler, config types.WatcherConfig) {
	defer wp.wg.Done()

	log.DebugH3("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			log.DebugH3("Worker %d shutting down", id)
			return

		case event, ok := <-wp.eventChan:
			if !ok {
				log.DebugH3("Worker %d: event channel closed", id)
				return
			}

			// Process the event
			if ShouldProcessEvent(event, config) {
				log.DebugH3("Worker %d processing: %s (%s)", id, event.Name, event.Op.String())
				ProcessEvent(event, handler)
			}
		}
	}
}

// Submit submits an event for processing
func (wp *WorkerPool) Submit(event fsnotify.Event) {
	select {
	case wp.eventChan <- event:
		// Event submitted successfully
	case <-wp.ctx.Done():
		// Worker pool is shutting down
		return
	default:
		// Channel is full, log warning but don't block
		log.Error("Worker pool event channel full, event may be delayed: %s", event.Name)
		// Try to submit with blocking
		select {
		case wp.eventChan <- event:
		case <-wp.ctx.Done():
			return
		}
	}
}

// Stop stops the worker pool gracefully
func (wp *WorkerPool) Stop() {
	log.InfoH3("Stopping worker pool...")
	wp.cancel()
	close(wp.eventChan)
	wp.wg.Wait()
	log.InfoH3("Worker pool stopped")
}

// WatchLoopWithWorkerPool is an optimized event loop using a worker pool
func WatchLoopWithWorkerPool(watcher *fsnotify.Watcher, config types.WatcherConfig, handler EventHandler, workers int, ctx <-chan struct{}) {
	pool := NewWorkerPool(workers)
	pool.Start(handler, config)
	defer pool.Stop()

	for {
		select {
		case <-ctx:
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Submit to worker pool for processing
			pool.Submit(event)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("Watcher error: %v", err)
		}
	}
}
