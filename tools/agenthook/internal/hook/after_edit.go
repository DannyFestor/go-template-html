package hook

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/check"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/checkout"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/session"
)

func (s Session) AfterEdit(path string) (string, error) {
	found, rel, inCheckout := checkout.Locate(path)
	if !inCheckout {
		return "", nil
	}

	err := session.Open(found.GitDir, s.id).RecordEdit(rel)
	if err != nil {
		return "", fmt.Errorf("record the edit: %w", err)
	}

	return check.AfterEdit(found.Root, rel).FirstFailure(check.Make{}, found.Root), nil
}
