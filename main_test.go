package main_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"multirepo/internal/testutil"
)

// binPath points at the multirepo binary compiled once in TestMain.
var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "multirepo-bin")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binName := "multirepo"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath = filepath.Join(dir, binName)
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build multirepo binary: %v\n%s", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runBinary executes the multirepo binary in dir and returns its combined
// output and exit code.
func runBinary(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return string(out), exitErr.ExitCode()
		}
		t.Fatalf("failed to run %s: %v\n%s", binPath, err, out)
	}
	return string(out), 0
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	testutil.WriteFile(t, dir, "repositories.yaml", content)
}

func TestUsageWithoutArguments(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	dir := t.TempDir()
	writeConfig(t, dir, fmt.Sprintf(`
repositories:
  myrepo:
    path: %s
    url: %s
    branch: main
`, origin, origin))

	output, exitCode := runBinary(t, dir)

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(output, "usage: multirepo <command>") {
		t.Errorf("expected usage text, got:\n%s", output)
	}
}

// The binary parses repositories.yaml before looking at the arguments, so
// asking for usage (or any command) outside a configured directory dies
// with a file-not-found error instead of printing usage.
func TestUsageWithoutConfigFile(t *testing.T) {
	output, _ := runBinary(t, t.TempDir())

	if !strings.Contains(output, "usage: multirepo") {
		t.Errorf("expected usage text even without repositories.yaml, got:\n%s", output)
	}
}

func TestUnknownCommandPrintsUsage(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	dir := t.TempDir()
	writeConfig(t, dir, fmt.Sprintf(`
repositories:
  myrepo:
    path: %s
    url: %s
    branch: main
`, origin, origin))

	output, exitCode := runBinary(t, dir, "bogus")

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(output, "usage: multirepo <command>") {
		t.Errorf("expected usage text for unknown command, got:\n%s", output)
	}
}

func TestRunWithoutArgumentsPrintsUsage(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	dir := t.TempDir()
	writeConfig(t, dir, fmt.Sprintf(`
repositories:
  myrepo:
    path: %s
    url: %s
    branch: main
`, origin, origin))

	output, exitCode := runBinary(t, dir, "run")

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(output, "usage: multirepo run") {
		t.Errorf("expected run usage text, got:\n%s", output)
	}
}

func TestSyncAndStatusEndToEnd(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	workDir := t.TempDir()
	repoPath := filepath.Join(t.TempDir(), "myrepo")
	writeConfig(t, workDir, fmt.Sprintf(`
repositories:
  myrepo:
    path: %s
    url: %s
    tag: v1.0.0
`, repoPath, origin))

	output, exitCode := runBinary(t, workDir, "sync")
	if exitCode != 0 {
		t.Fatalf("sync exit code = %d, want 0\n%s", exitCode, output)
	}
	if !strings.Contains(output, "1 repository detected") {
		t.Errorf("expected repository counter in sync output, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(repoPath, "file.txt")); err != nil {
		t.Fatalf("repository was not cloned by sync: %v", err)
	}

	output, exitCode = runBinary(t, workDir, "status")
	if exitCode != 0 {
		t.Fatalf("status exit code = %d, want 0\n%s", exitCode, output)
	}
	if !strings.Contains(output, "✔") || !strings.Contains(output, "(tag: v1.0.0)") {
		t.Errorf("expected clean status on tag v1.0.0, got:\n%s", output)
	}
}

func TestRunCommandEndToEnd(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	workDir := t.TempDir()
	writeConfig(t, workDir, fmt.Sprintf(`
repositories:
  myrepo:
    path: %s
    url: %s
    branch: main
`, clone, origin))

	output, exitCode := runBinary(t, workDir, "run", "myrepo", "git", "status")
	if exitCode != 0 {
		t.Fatalf("run exit code = %d, want 0\n%s", exitCode, output)
	}
	if !strings.Contains(output, "working tree clean") {
		t.Errorf("expected git status output, got:\n%s", output)
	}
}

// Environment variables in repository paths come from a .env file next to
// repositories.yaml, as documented in the README.
func TestDotEnvFileEndToEnd(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	workDir := t.TempDir()
	baseDir := t.TempDir()
	repoPath := filepath.Join(baseDir, "myrepo")

	writeConfig(t, workDir, fmt.Sprintf(`
repositories:
  myrepo:
    path: $MULTIREPO_TEST_BASE_DIR/myrepo
    url: %s
    tag: v1.0.0
`, origin))
	testutil.WriteFile(t, workDir, ".env", "MULTIREPO_TEST_BASE_DIR="+baseDir+"\n")

	output, exitCode := runBinary(t, workDir, "sync")
	if exitCode != 0 {
		t.Fatalf("sync exit code = %d, want 0\n%s", exitCode, output)
	}
	if !strings.Contains(output, "Loading environment variables from .env file") {
		t.Errorf("expected .env loading notice, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(repoPath, "file.txt")); err != nil {
		t.Fatalf("repository was not cloned at the .env-resolved path: %v", err)
	}
}

// Locking pins branch-tracked repositories to exact commits: after the
// origin advances, sync --locked must stay on the locked commit while a
// plain sync follows the branch.
func TestLockAndSyncLockedEndToEnd(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	workDir := t.TempDir()
	writeConfig(t, workDir, fmt.Sprintf(`
repositories:
  myrepo:
    path: %s
    url: %s
    branch: main
`, clone, origin))

	output, exitCode := runBinary(t, workDir, "lock")
	if exitCode != 0 {
		t.Fatalf("lock exit code = %d, want 0\n%s", exitCode, output)
	}
	lockedCommit := testutil.RunGit(t, clone, "rev-parse", "HEAD")

	// Advance origin's main past the locked commit.
	testutil.WriteFile(t, origin, "file.txt", "third version\n")
	testutil.RunGit(t, origin, "add", ".")
	testutil.RunGit(t, origin, "commit", "-m", "third commit")

	output, exitCode = runBinary(t, workDir, "sync", "--locked")
	if exitCode != 0 {
		t.Fatalf("sync --locked exit code = %d, want 0\n%s", exitCode, output)
	}
	if got := testutil.RunGit(t, clone, "rev-parse", "HEAD"); got != lockedCommit {
		t.Errorf("HEAD after sync --locked = %s, want locked commit %s", got, lockedCommit)
	}

	output, exitCode = runBinary(t, workDir, "sync")
	if exitCode != 0 {
		t.Fatalf("plain sync exit code = %d, want 0\n%s", exitCode, output)
	}
	if got, want := testutil.RunGit(t, clone, "rev-parse", "HEAD"), testutil.RunGit(t, origin, "rev-parse", "main"); got != want {
		t.Errorf("HEAD after plain sync = %s, want origin main %s", got, want)
	}
}

func TestVersionFlag(t *testing.T) {
	for _, arg := range []string{"--version", "-v", "version"} {
		output, exitCode := runBinary(t, t.TempDir(), arg)
		if exitCode != 0 {
			t.Errorf("%q exit code = %d, want 0\n%s", arg, exitCode, output)
		}
		if !strings.Contains(output, "multirepo version") {
			t.Errorf("%q output = %q, want it to contain 'multirepo version'", arg, output)
		}
	}
}

func TestHelpFlag(t *testing.T) {
	for _, arg := range []string{"--help", "-h", "help"} {
		output, exitCode := runBinary(t, t.TempDir(), arg)
		if exitCode != 0 {
			t.Errorf("%q exit code = %d, want 0 (asking for help is not an error)\n%s", arg, exitCode, output)
		}
		if !strings.Contains(output, "usage: multirepo <command>") {
			t.Errorf("%q output should contain usage text, got:\n%s", arg, output)
		}
	}
}

func TestStatusJSONEndToEnd(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	workDir := t.TempDir()
	writeConfig(t, workDir, fmt.Sprintf(`
repositories:
  myrepo:
    path: %s
    url: %s
    branch: main
`, clone, origin))

	output, exitCode := runBinary(t, workDir, "status", "--json")
	if exitCode != 0 {
		t.Fatalf("status --json exit code = %d, want 0\n%s", exitCode, output)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("status --json output is not valid JSON: %v\noutput:\n%s", err, output)
	}
	if len(rows) != 1 || rows[0]["name"] != "myrepo" {
		t.Errorf("unexpected JSON rows: %v", rows)
	}
}

func TestSyncJobsFlagEndToEnd(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	workDir := t.TempDir()
	base := t.TempDir()
	writeConfig(t, workDir, fmt.Sprintf(`
repositories:
  alpha:
    path: %[1]s/alpha
    url: %[2]s
    tag: v1.0.0
  beta:
    path: %[1]s/beta
    url: %[2]s
    tag: v1.0.0
`, base, origin))

	output, exitCode := runBinary(t, workDir, "sync", "--jobs", "2")
	if exitCode != 0 {
		t.Fatalf("sync --jobs 2 exit code = %d, want 0\n%s", exitCode, output)
	}
	for _, name := range []string{"alpha", "beta"} {
		if _, err := os.Stat(filepath.Join(base, name, "file.txt")); err != nil {
			t.Errorf("repository %s was not synced: %v", name, err)
		}
	}
}
