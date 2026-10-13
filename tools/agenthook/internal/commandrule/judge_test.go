package commandrule_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/commandrule"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

const (
	featureCheckout = "/repo"
	mainCheckout    = "/main"
)

func testEnviron() []string {
	return []string{"HOME=/home/dev", "TMPDIR=/tmp/session"}
}

// fakeCheckouts maps each checkout root to the branch it has checked out.
type fakeCheckouts map[string]string

func (f fakeCheckouts) Root(path string) (string, bool) {
	for root := range f {
		if path == root || strings.HasPrefix(path, root+string(filepath.Separator)) {
			return root, true
		}
	}

	return "", false
}

func (f fakeCheckouts) Branch(dir string) string {
	root, _ := f.Root(dir)

	return f[root]
}

type judgeTest struct {
	script string
	want   decision.Verdict
}

func TestShellFileWritesIntoTheRepoAreDenied(t *testing.T) {
	t.Parallel()

	requireVerdicts(t, []judgeTest{
		{`echo x > internal/foo.go`, decision.Deny},
		{`echo x >> /repo/README.md`, decision.Deny},
		{`cd /elsewhere && echo x > out`, decision.Pass},
		{`echo x > "$TMPDIR/out"`, decision.Pass},
		{`make lint > /dev/null 2>&1`, decision.Pass},
		{`git diff | tee internal/x.diff`, decision.Deny},
		{`make test | tee -a "$TMPDIR/log"`, decision.Pass},
		{`sed -i 's/a/b/' internal/foo.go`, decision.Deny},
		{`sed -i '' -e 's/a/b/' internal/foo.go`, decision.Deny},
		{`sed -Ei.bak 's/a/b/' internal/foo.go`, decision.Deny},
		{`sed --in-place=.bak 's/a/b/' internal/foo.go`, decision.Deny},
		{`sed -n 's/a/b/p' internal/foo.go`, decision.Pass},
		{`sed -i 's/a/b/' "$TMPDIR/out"`, decision.Pass},
		{`perl -pi -e 's/a/b/' internal/foo.go`, decision.Deny},
		{`perl -i.bak -pe 's/a/b/' internal/foo.go`, decision.Deny},
		{`perl -MFile::Find -ne 'print' internal/foo.go`, decision.Pass},
		{`cp /tmp/a internal/b`, decision.Deny},
		{`cp internal/a /tmp/b`, decision.Pass},
		{`mv -t internal /tmp/a`, decision.Deny},
		{`ln -s /tmp/a link`, decision.Deny},
		{`patch -p1 < "$TMPDIR/fix.diff"`, decision.Deny},
		{`cd /elsewhere && patch -p1 < fix.diff`, decision.Pass},
		{`git apply "$TMPDIR/fix.diff"`, decision.Deny},
		{`git apply --check "$TMPDIR/fix.diff"`, decision.Pass},
		{"python3 - <<'EOF'\nopen('internal/x.go', 'w').write('')\nEOF\n", decision.Deny},
		{"python3 - <<'EOF'\nprint(open('go.mod').read())\nEOF\n", decision.Pass},
		{`python3 -c "import pathlib; pathlib.Path('x').write_text('')"`, decision.Deny},
		{`node -e "require('fs').writeFileSync('x', '')"`, decision.Deny},
		{`ruby -e 'File.write("x", "")'`, decision.Deny},
		{`perl -e 'open(my $f, ">", "x")'`, decision.Deny},
		{`bash -c 'echo x > internal/foo.go'`, decision.Deny},
		{`cd internal && echo x > "$(git rev-parse --show-toplevel)/foo.go"`, decision.Deny},
		{`cd /elsewhere && cd "$(git -C /repo rev-parse --show-toplevel)" && echo x > foo.go`, decision.Deny},
		{`cd /elsewhere && echo x > "$(pwd)/out"`, decision.Pass},
		{`f=internal/foo.go; echo x > "$f"`, decision.Deny},
		{`tmp=$(mktemp); echo x > "$tmp"`, decision.Pass},
		{`git commit -m"Add login"`, decision.Pass},
	})
}

func TestPushesTargetingMainAreDenied(t *testing.T) {
	t.Parallel()

	requireVerdicts(t, []judgeTest{
		{`git push origin main`, decision.Deny},
		{`git push origin HEAD:main`, decision.Deny},
		{`git push origin feature:refs/heads/main`, decision.Deny},
		{`git push origin +main`, decision.Deny},
		{`git push --all origin`, decision.Deny},
		{`git -C /main push`, decision.Deny},
		{`cd /main && git push origin`, decision.Deny},
		{`cd /main && git push -u origin HEAD`, decision.Deny},
		{`git push`, decision.Pass},
		{`git push -u origin HEAD`, decision.Pass},
		{`git push --force-with-lease origin feature`, decision.Pass},
		{`git -c push.default=current push origin feature`, decision.Pass},
		{`cd /main && git push origin feature`, decision.Pass},
	})
}

func TestDestructiveGitCommandsAsk(t *testing.T) {
	t.Parallel()

	requireVerdicts(t, []judgeTest{
		{`git reset --hard`, decision.Ask},
		{`git reset --hard origin/main`, decision.Ask},
		{`git clean -fd`, decision.Ask},
		{`git clean -xdf`, decision.Ask},
		{`git checkout -- internal/foo.go`, decision.Ask},
		{`git checkout .`, decision.Ask},
		{`git restore internal/foo.go`, decision.Ask},
		{`git stash drop`, decision.Ask},
		{`git stash clear`, decision.Ask},
		{`git branch -D feature`, decision.Ask},
		{`git branch --delete --force feature`, decision.Ask},
		{`git worktree remove --force ../tree`, decision.Ask},
		{`git reset --soft HEAD~1`, decision.Pass},
		{`git clean -n`, decision.Pass},
		{`git checkout feature`, decision.Pass},
		{`git restore --staged internal/foo.go`, decision.Pass},
		{`git stash`, decision.Pass},
		{`git branch -d feature`, decision.Pass},
		{`git worktree remove ../tree`, decision.Pass},
		{`cd /elsewhere && git reset --hard`, decision.Pass},
	})
}

func TestDependencyAddsAsk(t *testing.T) {
	t.Parallel()

	requireVerdicts(t, []judgeTest{
		{`go get example.com/x`, decision.Ask},
		{`go mod edit -require=example.com/x@v1.0.0`, decision.Ask},
		{`go mod edit -tool example.com/x`, decision.Ask},
		{`pnpm add -D left-pad`, decision.Ask},
		{`VP_HOME=.vite-plus bin/vp add -D left-pad`, decision.Ask},
		{`npm install left-pad`, decision.Ask},
		{`npm i -D left-pad`, decision.Ask},
		{`go mod tidy`, decision.Pass},
		{`go mod edit -droprequire=example.com/x`, decision.Pass},
		{`VP_HOME=.vite-plus bin/vp install --frozen-lockfile`, decision.Pass},
		{`npm install`, decision.Pass},
		{`npm ci`, decision.Pass},
		{`cd /elsewhere && go get example.com/x`, decision.Pass},
	})
}

func TestInvalidShellPasses(t *testing.T) {
	t.Parallel()

	requireVerdicts(t, []judgeTest{{`echo "unterminated > internal/foo.go`, decision.Pass}})
}

func TestReasonsNameTheWayForward(t *testing.T) {
	t.Parallel()

	tests := []struct {
		script     string
		wantReason string
	}{
		{`echo x > internal/foo.go`, "edit tool"},
		{`git push origin main`, "pull request"},
		{`git reset --hard`, "discards"},
		{`go get example.com/x`, "dependency"},
	}

	for _, test := range tests {
		got := newJudge().Script(test.script, featureCheckout)
		if !strings.Contains(got.Reason, test.wantReason) {
			t.Errorf("Script(%q).Reason = %q, want it to contain %q", test.script, got.Reason, test.wantReason)
		}
	}
}

func newJudge() commandrule.Judge {
	return commandrule.NewJudge(fakeCheckouts{featureCheckout: "feature", mainCheckout: "main"}, testEnviron())
}

func requireVerdicts(t *testing.T, tests []judgeTest) {
	t.Helper()

	judge := newJudge()

	for _, test := range tests {
		got := judge.Script(test.script, featureCheckout)
		if got.Verdict != test.want {
			t.Errorf("Script(%q) = %v %q, want %v", test.script, got.Verdict, got.Reason, test.want)
		}
	}
}
