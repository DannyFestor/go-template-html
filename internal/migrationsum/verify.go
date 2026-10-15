package migrationsum

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

var ErrOutOfDate = errors.New(FileName + " is out of date, run make generate")

func Verify(fsys fs.FS) error {
	recorded, err := fs.ReadFile(fsys, FileName)
	if err != nil {
		return fmt.Errorf("read %s: %w", FileName, err)
	}

	sum, err := Compute(fsys)
	if err != nil {
		return err
	}

	if string(recorded) != sum.String() {
		return fmt.Errorf("%w: %s", ErrOutOfDate, strings.Join(changedFiles(string(recorded), sum.String()), ", "))
	}

	return nil
}

func changedFiles(recorded, computed string) []string {
	var names []string

	for _, line := range symmetricDifference(fileLines(recorded), fileLines(computed)) {
		name, _, _ := strings.Cut(line, " ")
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}

	return names
}

func fileLines(sum string) []string {
	return strings.Split(strings.TrimSuffix(sum, "\n"), "\n")[1:]
}

func symmetricDifference(left, right []string) []string {
	var difference []string

	for _, line := range left {
		if !slices.Contains(right, line) {
			difference = append(difference, line)
		}
	}

	for _, line := range right {
		if !slices.Contains(left, line) {
			difference = append(difference, line)
		}
	}

	return difference
}
