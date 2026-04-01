package jj

import (
	"strings"
	"time"
)

// RepoWatcher polls the repository for changes every second,
// signaling on C when the operation state changes.
type RepoWatcher struct {
	runner   *Runner
	C        chan struct{}
	done     chan struct{}
	cancel   chan struct{}
	prevOpID string
}

func NewRepoWatcher(runner *Runner) *RepoWatcher {
	rw := &RepoWatcher{
		runner: runner,
		C:      make(chan struct{}, 1),
		done:   make(chan struct{}),
		cancel: make(chan struct{}),
	}
	rw.prevOpID, _ = runner.Run("operation", "log", "--no-graph", "--limit=1", "-T", "self.id()")
	rw.prevOpID = strings.TrimSpace(rw.prevOpID)

	go rw.loop()
	return rw
}

func (rw *RepoWatcher) loop() {
	defer close(rw.done)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rw.cancel:
			return
		case <-ticker.C:
			if rw.changed() {
				select {
				case rw.C <- struct{}{}:
				default:
				}
			}
		}
	}
}

func (rw *RepoWatcher) changed() bool {
	opID, err := rw.runner.Run("operation", "log", "--no-graph", "--limit=1", "-T", "self.id()")
	if err != nil {
		return false
	}
	opID = strings.TrimSpace(opID)
	if opID != rw.prevOpID {
		rw.prevOpID = opID
		return true
	}
	return false
}

func (rw *RepoWatcher) Close() error {
	close(rw.cancel)
	<-rw.done
	return nil
}
