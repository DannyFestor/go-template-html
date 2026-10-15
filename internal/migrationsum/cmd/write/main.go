package main

import (
	"fmt"
	"os"

	"github.com/DannyFestor/go-template-html/internal/migrationsum"
)

const sumFileMode = 0o644

func main() {
	err := write()
	if err != nil {
		fmt.Fprintln(os.Stderr, "migrationsum:", err)
		os.Exit(1)
	}
}

func write() error {
	sum, err := migrationsum.Compute(os.DirFS("."))
	if err != nil {
		return fmt.Errorf("compute %s: %w", migrationsum.FileName, err)
	}

	err = os.WriteFile(migrationsum.FileName, []byte(sum.String()), sumFileMode)
	if err != nil {
		return fmt.Errorf("write %s: %w", migrationsum.FileName, err)
	}

	return nil
}
