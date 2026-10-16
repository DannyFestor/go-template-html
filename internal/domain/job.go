package domain

import "time"

type JobID string

type ClaimedJob struct {
	ID      JobID
	Kind    string
	Args    []byte
	Attempt int
}

type JobLease struct {
	Kind     string
	Duration time.Duration
}
