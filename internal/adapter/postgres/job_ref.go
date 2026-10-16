package postgres

import (
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
