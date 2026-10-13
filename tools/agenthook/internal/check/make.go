package check

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/environ"
)

// Run is one finished Make target. Stdout is kept apart from Stderr for targets whose output is read.
type Run struct {
	Target Target
	Stdout string
	Stderr string
	Failed bool
}

func (r Run) Output() string {
	return strings.TrimSpace(r.Stdout + "\n" + r.Stderr)
}

// Make runs targets with the Makefile at a checkout root, which ADR-0010 makes the only command surface.
type Make struct{}

func (Make) Run(root string, target Target) Run {
	args := append([]string{"-s", "--no-print-directory", "-C", root}, target.makeArgs()...)
	cmd := exec.CommandContext(context.Background(), "make", args...) //nolint:gosec // The hooks build every argument.
	// The hook itself runs under make, which exports its command-line variables, such as GO_MODULE, both on their own
	// and in MAKEFLAGS, so every target run here would pick them up.
	cmd.Env = environ.Without(os.Environ(), "MAKEFLAGS", "MFLAGS", "MAKELEVEL", goModuleVariable, filesVariable)

	var stdout, stderr bytes.Buffer

	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := cmd.Run()

	return Run{Target: target, Stdout: stdout.String(), Stderr: stderr.String(), Failed: err != nil}
}
