package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/logreport"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/buildinfo"
	"github.com/DannyFestor/go-template-html/internal/worker"
)

type WorkerApp struct {
	Jobs   *worker.Registry
	worker *worker.Worker
	pool   *postgres.Pool
}

func RunWorker() int {
	return run(ServeWorker)
}

func ServeWorker(ctx context.Context, config Config, logger *slog.Logger) error {
	app, err := NewWorkerApp(ctx, config, logger, newSMTPSender(config))
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "worker started",
		slog.String("version", buildinfo.Version),
		slog.Int("concurrency", config.Queue.Concurrency),
	)

	defer app.Close()

	err = app.worker.Run(ctx)
	if err != nil {
		return fmt.Errorf("run the worker: %w", err)
	}

	return nil
}

func NewWorkerApp(ctx context.Context, config Config, logger *slog.Logger, sender mailSender) (*WorkerApp, error) {
	err := checkWorkerPoolSize(config)
	if err != nil {
		return nil, err
	}

	pool, err := postgres.Open(ctx, workerPostgresConfig(config.DB))
	if err != nil {
		return nil, fmt.Errorf("connect to the database: %w", err)
	}

	jobQueue := newJobQueue(pool)
	jobs := worker.NewRegistry()
	registerQueueJobs(jobs, jobQueue, config.Queue)
	registerMailJobs(jobs, sender)
	registerAuthJobs(jobs, pool)

	return &WorkerApp{
		Jobs:   jobs,
		worker: worker.New(workerDependencies(jobQueue, jobs, logger), workerConfig(config)),
		pool:   pool,
	}, nil
}

func (a *WorkerApp) RunDue(ctx context.Context) error {
	err := a.worker.RunDue(ctx)
	if err != nil {
		return fmt.Errorf("run the due jobs: %w", err)
	}

	return nil
}

func (a *WorkerApp) Close() {
	a.pool.Close()
}

func workerDependencies(jobQueue *postgres.JobQueue, jobs *worker.Registry, logger *slog.Logger) worker.Dependencies {
	return worker.Dependencies{
		Store:    jobQueue,
		Registry: jobs,
		Clock:    clock.System{},
		Reporter: logreport.New(logger),
		Logger:   logger,
	}
}

func workerConfig(config Config) worker.Config {
	return worker.Config{
		Concurrency:     config.Queue.Concurrency,
		PollInterval:    config.Queue.PollInterval,
		ShutdownTimeout: config.App.ShutdownTimeout,
	}
}
