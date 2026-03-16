package jj

import "time"

// Commit represents a jj commit/change.
type Commit struct {
	ChangeID    string
	CommitID    string
	Description string
	Author      string
	Timestamp   time.Time
	Bookmarks   []string
	IsEmpty     bool
	IsConflict  bool
	Parents     []string
}

// FileChange represents a changed file in a revision.
type FileChange struct {
	Status string // A, M, D, R, or C
	Path   string
}

// GraphEntry pairs a Commit with its ANSI-colored graph lines from jj log.
type GraphEntry struct {
	Commit Commit
	Lines  []string
}
