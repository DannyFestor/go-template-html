package shell_test

import (
	"slices"
	"testing"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/shell"
)

const startDir = "/repo"

func testStart() shell.Start {
	return shell.Start{
		Dir:        startDir,
		Environ:    []string{"HOME=/home/dev", "TMPDIR=/tmp/session"},
		Substitute: substitutePwd,
	}
}

func substitutePwd(command shell.Command) (string, bool) {
	return command.Dir, command.Name() == "pwd"
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		want   []shell.Command
	}{
		{
			name:   "quoted arguments",
			script: `git commit -m"Add login"`,
			want:   []shell.Command{ran(startDir, "git", "commit", "-mAdd login")},
		},
		{
			name:   "expanded variables and tilde",
			script: `ls "$TMPDIR" ~/notes`,
			want:   []shell.Command{ran(startDir, "ls", "/tmp/session", "/home/dev/notes")},
		},
		{
			name:   "command substitution stays unresolved",
			script: `cp a "$(mktemp)"`,
			want: []shell.Command{
				ran(startDir, "cp", "a", ""),
				ran(startDir, "mktemp"),
			},
		},
		{
			name:   "cd moves the following commands",
			script: `cd internal && go test ./... ; cd /other; ls`,
			want: []shell.Command{
				ran(startDir, "cd", "internal"),
				ran("/repo/internal", "go", "test", "./..."),
				ran("/repo/internal", "cd", "/other"),
				ran("/other", "ls"),
			},
		},
		{
			name:   "cd inside a subshell does not leak",
			script: `(cd sub && make) && ls`,
			want: []shell.Command{
				ran(startDir, "cd", "sub"),
				ran("/repo/sub", "make"),
				ran(startDir, "ls"),
			},
		},
		{
			name:   "cd without an argument goes home",
			script: `cd; ls`,
			want: []shell.Command{
				ran(startDir, "cd"),
				ran("/home/dev", "ls"),
			},
		},
		{
			name:   "cd to an unresolved directory leaves it unknown",
			script: `cd "$(git rev-parse --show-toplevel)"; ls`,
			want: []shell.Command{
				ran(startDir, "cd", ""),
				ran(startDir, "git", "rev-parse", "--show-toplevel"),
				ran("", "ls"),
			},
		},
		{
			name:   "wrappers are dropped",
			script: `env -i FOO=1 command nohup git push`,
			want:   []shell.Command{ran(startDir, "git", "push")},
		},
		{
			name:   "file redirects are writes",
			script: `echo x > internal/foo.go 2>>"$TMPDIR/log" &>/abs/all`,
			want: []shell.Command{
				writing(ran(startDir, "echo", "x"), "/repo/internal/foo.go", "/tmp/session/log", "/abs/all"),
			},
		},
		{
			name:   "descriptor duplications and input redirects are not writes",
			script: `make 2>&1 < input >&-`,
			want:   []shell.Command{ran(startDir, "make")},
		},
		{
			name:   "redirect of a compound command",
			script: `{ echo a; } > out`,
			want: []shell.Command{
				writing(ran(startDir), "/repo/out"),
				ran(startDir, "echo", "a"),
			},
		},
		{
			name:   "heredoc is the input",
			script: "python3 - <<'EOF'\nopen('a', 'w')\nEOF\n",
			want: []shell.Command{
				reading(ran(startDir, "python3", "-"), "open('a', 'w')\n"),
			},
		},
		{
			name:   "here-string is the input",
			script: `node <<< "require('fs')"`,
			want:   []shell.Command{reading(ran(startDir, "node"), "require('fs')")},
		},
		{
			name:   "shell -c scripts are parsed too",
			script: `cd sub && bash -c "git reset --hard"`,
			want: []shell.Command{
				ran(startDir, "cd", "sub"),
				ran("/repo/sub", "bash", "-c", "git reset --hard"),
				ran("/repo/sub", "git", "reset", "--hard"),
			},
		},
		{
			name:   "a heredoc fed to a shell is parsed too",
			script: "sh <<EOF\necho x > f\nEOF\n",
			want: []shell.Command{
				reading(ran(startDir, "sh"), "echo x > f\n"),
				writing(ran(startDir, "echo", "x"), "/repo/f"),
			},
		},
		{
			name:   "variables the script assigns expand later",
			script: `f=internal/foo.go; export g="$f.bak"; echo x > "$f" > "$g"`,
			want: []shell.Command{
				ran(startDir),
				writing(ran(startDir, "echo", "x"), "/repo/internal/foo.go", "/repo/internal/foo.go.bak"),
			},
		},
		{
			name:   "a variable assigned an unresolved value expands to nothing",
			script: `TMPDIR=$(mktemp -d); echo x > "$TMPDIR"`,
			want: []shell.Command{
				ran(startDir),
				ran(startDir, "mktemp", "-d"),
				ran(startDir, "echo", "x"),
			},
		},
		{
			name:   "an assignment inside a subshell does not leak",
			script: `(f=a); echo x > "/repo/$f"`,
			want: []shell.Command{
				ran(startDir),
				writing(ran(startDir, "echo", "x"), "/repo"),
			},
		},
		{
			name:   "substitutions the caller resolves are expanded",
			script: `cd "$(pwd)/sub" && echo x > "$(pwd)/out"`,
			want: []shell.Command{
				ran(startDir, "cd", "/repo/sub"),
				ran(startDir, "pwd"),
				writing(ran("/repo/sub", "echo", "x"), "/repo/sub/out"),
				ran("/repo/sub", "pwd"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := shell.Parse(test.script, testStart())
			if err != nil {
				t.Fatalf("Parse(%q): %v", test.script, err)
			}

			requireCommands(t, got, test.want)
		})
	}
}

func TestParseRejectsInvalidSyntax(t *testing.T) {
	t.Parallel()

	_, err := shell.Parse(`echo "unterminated`, testStart())
	if err == nil {
		t.Fatal("Parse succeeded, want a syntax error")
	}
}

func ran(dir string, args ...string) shell.Command {
	return shell.Command{Args: args, Dir: dir, Writes: nil, Input: ""}
}

func writing(command shell.Command, paths ...string) shell.Command {
	command.Writes = paths

	return command
}

func reading(command shell.Command, input string) shell.Command {
	command.Input = input

	return command
}

func requireCommands(t *testing.T, got, want []shell.Command) {
	t.Helper()

	if !slices.EqualFunc(got, want, equalCommands) {
		t.Errorf("commands =\n%+v\nwant\n%+v", got, want)
	}
}

func equalCommands(a, b shell.Command) bool {
	return slices.Equal(a.Args, b.Args) && a.Dir == b.Dir && slices.Equal(a.Writes, b.Writes) && a.Input == b.Input
}
