package testdb

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/adapter/goose"
	"github.com/DannyFestor/go-template-html/internal/migrationsum"
	"github.com/DannyFestor/go-template-html/migrations"
)

const (
	templateLockID   = 7_301_984_512
	templateHashSize = 8
)

type template struct {
	settings settings
	name     string
}

func prepareTemplate(ctx context.Context) (template, error) {
	s, err := settingsFromEnv()
	if err != nil {
		return template{}, err
	}

	name, err := templateName(s)
	if err != nil {
		return template{}, err
	}

	err = withAdmin(ctx, s, func(a admin) error { return ensureTemplate(ctx, a, s, name) })
	if err != nil {
		return template{}, err
	}

	return template{settings: s, name: name}, nil
}

func templatePrefix(s settings) string {
	return s.name + "_tpl_"
}

func templateName(s settings) (string, error) {
	sum, err := migrationsum.Compute(migrations.FS)
	if err != nil {
		return "", fmt.Errorf("hash the migrations: %w", err)
	}

	total := sum.Total()

	return templatePrefix(s) + hex.EncodeToString(total[:templateHashSize]), nil
}

func ensureTemplate(ctx context.Context, a admin, s settings, name string) error {
	err := a.exec(ctx, "SELECT pg_advisory_lock($1)", templateLockID)
	if err != nil {
		return err
	}

	exists, err := a.databaseExists(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return buildTemplate(ctx, a, s, name)
}

// The template is migrated under another name and renamed when done, so a run that dies halfway leaves no template
// that looks complete.
func buildTemplate(ctx context.Context, a admin, s settings, name string) error {
	building := name + "_building"

	err := a.drop(ctx, building)
	if err != nil {
		return err
	}

	err = a.create(ctx, building, "template0")
	if err != nil {
		return err
	}

	err = migrate(ctx, s.database(building))
	if err != nil {
		return err
	}

	err = a.rename(ctx, building, name)
	if err != nil {
		return err
	}

	return pruneStaleTemplates(ctx, a, s, name)
}

func migrate(ctx context.Context, database Database) error {
	err := goose.WithMigrator(ctx, database.PostgresConfig(), migrations.FS, func(migrator *goose.Migrator) error {
		_, err := migrator.Up(ctx)
		if err != nil {
			return fmt.Errorf("apply the migrations: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("migrate the template: %w", err)
	}

	return nil
}

func pruneStaleTemplates(ctx context.Context, a admin, s settings, current string) error {
	names, err := a.databasesStartingWith(ctx, templatePrefix(s))
	if err != nil {
		return err
	}

	for _, name := range names {
		if name == current {
			continue
		}

		err = a.drop(ctx, name)
		if err != nil {
			return err
		}
	}

	return nil
}
