package checkout

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/environ"
)

// gitOutput runs git in dir. GIT_* variables are dropped, because a GIT_DIR exported by a surrounding git hook
// would override the checkout dir is in.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", args...) //nolint:gosec // The hooks build every argument.
	cmd.Dir = dir
	cmd.Env = environ.WithoutPrefix(os.Environ(), "GIT_")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), dir, err)
	}

	return strings.TrimSpace(string(output)), nil
}
