package hook

import (
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/checkout"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/commandrule"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

func (s Session) BeforeCommand(script, dir string) decision.Decision {
	return commandrule.NewJudge(gitCheckouts{}, s.environ).Script(script, dir)
}

type gitCheckouts struct{}

func (gitCheckouts) Root(path string) (string, bool) {
	found, inCheckout := checkout.Find(path)

	return found.Root, inCheckout
}

func (gitCheckouts) Branch(dir string) string {
	found, inCheckout := checkout.Find(dir)
	if !inCheckout {
		return ""
	}

	return found.Branch()
}
