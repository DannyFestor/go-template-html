package postgres

import (
	"fmt"
	"uuid"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

func parseUserID(id domain.UserID) (uuid.UUID, error) {
	return parseID("user", id)
}

func parseJobID(id domain.JobID) (uuid.UUID, error) {
	return parseID("job", id)
}

func parseID[ID ~string](entity string, id ID) (uuid.UUID, error) {
	parsed, err := uuid.Parse(string(id))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse the %s ID %q: %w", entity, id, err)
	}

	return parsed, nil
}
