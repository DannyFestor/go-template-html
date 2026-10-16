package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type job interface {
	Kind() string
}

type handler[T job] interface {
	Handle(ctx context.Context, job T) error
}

type registration struct {
	settings Settings
	handle   func(ctx context.Context, args []byte) error
}

type Registry struct {
	registrations map[string]registration
	periodic      []periodicJob
}

func NewRegistry() *Registry {
	return &Registry{registrations: map[string]registration{}, periodic: nil}
}

func RegisterPeriodic[T job](registry *Registry, handler handler[T], settings Settings, interval time.Duration) {
	Register(registry, handler, settings)

	var zero T

	registry.periodic = append(registry.periodic, periodicJob{job: zero, interval: interval})
}

func Register[T job](registry *Registry, handler handler[T], settings Settings) {
	var zero T

	registry.registrations[zero.Kind()] = registration{
		settings: settings,
		handle: func(ctx context.Context, args []byte) error {
			var decoded T

			err := json.Unmarshal(args, &decoded)
			if err != nil {
				return fmt.Errorf("decode the %s job: %w", zero.Kind(), err)
			}

			return handler.Handle(ctx, decoded)
		},
	}
}

func (r *Registry) leases() []domain.JobLease {
	leases := make([]domain.JobLease, 0, len(r.registrations))
	for kind, registered := range r.registrations {
		leases = append(leases, domain.JobLease{Kind: kind, Duration: registered.settings.Lease})
	}

	return leases
}

func (r *Registry) lookup(kind string) registration {
	return r.registrations[kind]
}
