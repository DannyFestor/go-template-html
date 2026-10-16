package postgres

import (
	"fmt"
	"math"
	"uuid"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type jobRef struct {
	id      uuid.UUID
	attempt int32
}

func newJobRef(job domain.ClaimedJob) (jobRef, error) {
	id, err := parseJobID(job.ID)
	if err != nil {
		return jobRef{}, err
	}

	return jobRef{id: id, attempt: int32(min(job.Attempt, math.MaxInt32))}, nil //nolint:gosec // min keeps it in range.
}

func parseJobID(id domain.JobID) (uuid.UUID, error) {
	parsed, err := uuid.Parse(string(id))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse the job ID %q: %w", id, err)
	}

	return parsed, nil
}
