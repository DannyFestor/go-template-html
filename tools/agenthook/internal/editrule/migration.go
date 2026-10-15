package editrule

import "github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"

const (
	migrationPattern = "migrations/*.sql"
	createMigration  = "Run `make migrate-create NAME=<snake_case_name>`, which writes the next numbered stub and " +
		"refreshes migrations.sum, then edit that stub."
)

func (j Judge) migrationEdit(edit Edit) decision.Decision {
	switch {
	case !matches(migrationPattern, edit.Path):
		return decision.Passed()
	case edit.Created:
		return decision.Denied(edit.Path + " would be a new migration made by hand. " + createMigration)
	case j.mainline.OnMainline(edit.Path):
		return decision.Denied(edit.Path + " is already on origin/main, so databases have applied it and it must " +
			"not change. Fix it forward in a new migration: " + createMigration)
	default:
		return decision.Passed()
	}
}
