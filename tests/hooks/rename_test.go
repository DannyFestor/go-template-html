package hooks_test

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	renameScript = "rename.sh"

	templateModule      = "github.com/DannyFestor/go-template-html"
	templateSlug        = "go-template-html"
	templateDBName      = "go_template_html"
	templateDisplayName = "Go Template"

	cloneModule = "example.com/acme/my-shop"

	fixtureFile     = "config.txt"
	templateFixture = "module: " + templateModule + "/internal/web\n" +
		"slug: " + templateSlug + "\n" +
		"database: " + templateDBName + "\n" +
		"name: " + templateDisplayName + "\n"
	renamedFixture = "module: " + cloneModule + "/internal/web\n" +
		"slug: my-shop\n" +
		"database: my_shop\n" +
		"name: My Shop\n"
)

func newTemplateRepo(t *testing.T) string {
	t.Helper()

	repo := newRepo(t)
	commitContent(t, repo, "go.mod", "module "+templateModule+"\n")
	commitContent(t, repo, fixtureFile, templateFixture)

	return repo
}

func renameSweep(t *testing.T, repo string, args ...string) scriptRun {
	t.Helper()

	return runScript(t, repo, "", renameScript, append([]string{"sweep"}, args...)...)
}

func renameVerify(t *testing.T, repo string) scriptRun {
	t.Helper()

	return runScript(t, repo, "", renameScript, "verify")
}

func TestRenameSweepReplacesEveryTemplateTokenInTrackedFiles(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)

	run := renameSweep(t, repo, cloneModule)

	requirePassed(t, run)
	requireContent(t, repo, fixtureFile, renamedFixture)
	requireContent(t, repo, "go.mod", "module "+cloneModule+"\n")
}

func TestRenameSweepNamesTheAppAfterTheGivenDisplayNameLiterally(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	commitContent(t, repo, "app.txt", templateDisplayName+"\n")

	run := renameSweep(t, repo, cloneModule, `Acme & Co | A\B`)

	requirePassed(t, run)
	requireContent(t, repo, "app.txt", "Acme & Co | A\\B\n")
}

func TestRenameSweepSummarisesTheNames(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)

	run := renameSweep(t, repo, cloneModule)

	requireOutput(t, run, "module "+cloneModule)
	requireOutput(t, run, "slug my-shop")
	requireOutput(t, run, "database my_shop")
	requireOutput(t, run, "display name My Shop")
}

func TestRenameSweepLeavesUntrackedFilesAlone(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	writeFile(t, repo, ".env", templateFixture)

	run := renameSweep(t, repo, cloneModule)

	requirePassed(t, run)
	requireContent(t, repo, ".env", templateFixture)
}

func TestRenameSweepLeavesTheFilesHoldingTheTemplateTokensAlone(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	commitContent(t, repo, "scripts/rename.sh", templateFixture)
	commitContent(t, repo, "tests/hooks/rename_test.go", templateFixture)

	run := renameSweep(t, repo, cloneModule)

	requirePassed(t, run)
	requireContent(t, repo, "scripts/rename.sh", templateFixture)
	requireContent(t, repo, "tests/hooks/rename_test.go", templateFixture)
}

func TestRenameSweepKeepsAnExecutableFileExecutable(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	writeFile(t, repo, "run.sh", templateFixture)

	err := os.Chmod(filepath.Join(repo, "run.sh"), 0o700) //nolint:gosec // The fixture is a script.
	if err != nil {
		t.Fatalf("make run.sh executable: %v", err)
	}

	git(t, repo, "add", "run.sh")
	git(t, repo, "commit", "--quiet", "--message", "add run.sh")

	run := renameSweep(t, repo, cloneModule)

	requirePassed(t, run)
	requireContent(t, repo, "run.sh", renamedFixture)
	requireExecutable(t, filepath.Join(repo, "run.sh"))
}

func TestRenameSweepRefusesARenamedClone(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	requirePassed(t, renameSweep(t, repo, cloneModule))

	run := renameSweep(t, repo, "example.com/acme/other")

	requireRefused(t, run, "already renamed")
	requireContent(t, repo, fixtureFile, renamedFixture)
}

func TestRenameSweepRefusesARenamedCloneWhateverTheModule(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	requirePassed(t, renameSweep(t, repo, cloneModule))

	run := renameSweep(t, repo, "")

	requireRefused(t, run, "already renamed")
}

func TestRenameSweepRefusesAModuleWithoutAValidSlug(t *testing.T) {
	t.Parallel()

	for _, module := range []string{"", "example.com/acme/Shop", "example.com/acme/1shop", "example.com/acme/my_shop"} {
		repo := newTemplateRepo(t)

		run := renameSweep(t, repo, module)

		requireRefused(t, run, "MODULE")
		requireContent(t, repo, fixtureFile, templateFixture)
	}
}

func TestRenameVerifyFailsOnALeftoverTokenAndNamesItsFile(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	requirePassed(t, renameSweep(t, repo, cloneModule))
	commitContent(t, repo, "internal/view/page.txt", templateDisplayName+"\n")

	run := renameVerify(t, repo)

	requireRefused(t, run, "internal/view/page.txt")
}

func TestRenameVerifyPrintsTheNextStepsWhenNoTokenIsLeft(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	requirePassed(t, renameSweep(t, repo, cloneModule))

	run := renameVerify(t, repo)

	requirePassed(t, run)
	requireOutput(t, run, "make setup")
	requireOutput(t, run, "make github-setup")
	requireOutput(t, run, "commit")
}

func TestRenameVerifyIgnoresTheFilesHoldingTheTemplateTokens(t *testing.T) {
	t.Parallel()

	repo := newTemplateRepo(t)
	commitContent(t, repo, "scripts/rename.sh", templateFixture)
	commitContent(t, repo, "tests/hooks/rename_test.go", templateFixture)
	requirePassed(t, renameSweep(t, repo, cloneModule))

	run := renameVerify(t, repo)

	requirePassed(t, run)
}

func requireExecutable(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}

	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("%s has mode %v, want it executable", path, info.Mode().Perm())
	}
}
