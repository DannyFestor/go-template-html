package commandrule

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	interpreterName = regexp.MustCompile(`^(python[0-9.]*|node|ruby|perl|php|deno|bun)$`)
	fileWriteCall   = regexp.MustCompile(strings.Join([]string{
		`\bf?open\s*\([^)]*['"][rbt]*[wax+][rbt+]*['"]`,
		`\bopen\s*\(?[^)]*['"]\+?>`,
		`\.write_(text|bytes)\s*\(`,
		`\b(write|append|copy)(Text)?File(Sync)?\s*\(`,
		`\bcreateWriteStream\s*\(`,
		`\b(rename|unlink|rmdir|mkdir)(Sync)?\s*\(`,
		`\bos\.(replace|remove|makedirs)\s*\(`,
		`\bshutil\.\w+\s*\(`,
		`\b(File|IO)\.(write|rename|delete)\b`,
		`\bFileUtils\.`,
		`\bfile_put_contents\s*\(`,
	}, "|"))
)

func interpreterWritesFiles(command invocation) bool {
	if !interpreterName.MatchString(filepath.Base(command.Name())) {
		return false
	}

	return fileWriteCall.MatchString(inlineScript(command))
}

func inlineScript(command invocation) string {
	spec := argSpec{valueShorts: "ceE", attachedShorts: "", valueLongs: []string{"--eval", "--print"}}
	inlineCode := []string{"-c", "-e", "-E", "--eval", "--print"}

	args := spec.scan(command.Args[1:])
	if args.has(inlineCode...) {
		return args.value(inlineCode...)
	}

	if len(args.operands) == 0 || args.operands[0] == "-" {
		return command.Input
	}

	return ""
}
