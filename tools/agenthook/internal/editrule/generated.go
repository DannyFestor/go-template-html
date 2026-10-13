package editrule

import (
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/comment"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

const regenerateWithMake = "Edit its source and run `make generate`."

type generatedFile struct {
	pattern string
	remedy  string
}

func generatedFiles() []generatedFile {
	return []generatedFile{
		{pattern: "internal/asset/dist/**", remedy: "Edit internal/asset/src/ and run `make vite-build`."},
		{pattern: "**/migrations.sum", remedy: regenerateWithMake},
		{pattern: "**/go.sum", remedy: "Run `go get` or `go mod tidy`, which rewrite it."},
		{
			pattern: "**/pnpm-lock.yaml",
			remedy:  "Run `VP_HOME=\"$PWD/.vite-plus\" bin/vp add` or `remove` from the checkout root, which rewrite it.",
		},
	}
}

func generatedOutput(edit Edit) decision.Decision {
	if comment.HasGeneratedHeader(edit.Original) || comment.HasGeneratedHeader(edit.Inserted) {
		return deniedAsGenerated(edit, regenerateWithMake)
	}

	for _, generated := range generatedFiles() {
		if matches(generated.pattern, edit.Path) {
			return deniedAsGenerated(edit, generated.remedy)
		}
	}

	return decision.Passed()
}

func deniedAsGenerated(edit Edit, remedy string) decision.Decision {
	return decision.Denied(edit.Path + " is generated, so edits to it are refused. " + remedy)
}
