// Package testutil provides helpers for tests that need real git
// repositories and stdout capturing. It is only imported from _test files.
package testutil

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// RunGit runs a git command in dir and fails the test on error.
// It returns the trimmed combined output.
func RunGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// configureIdentity sets a local committer identity so commits and
// stashes work on CI runners without global git config.
func configureIdentity(t *testing.T, dir string) {
	t.Helper()
	RunGit(t, dir, "config", "user.email", "test@example.com")
	RunGit(t, dir, "config", "user.name", "Test User")
	RunGit(t, dir, "config", "commit.gpgsign", "false")
	RunGit(t, dir, "config", "tag.gpgsign", "false")
	// Git for Windows defaults to core.autocrlf=true, which rewrites line
	// endings on checkout and can make freshly cloned fixtures look dirty.
	RunGit(t, dir, "config", "core.autocrlf", "false")
}

// CreateOriginRepo creates a git repository in a temp directory with:
//   - a "main" branch with two commits touching file.txt
//   - tag "v1.0.0" on the first commit
//   - branch "feature" on the first commit
//
// It returns the repository path.
func CreateOriginRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	RunGit(t, dir, "init", "-b", "main")
	configureIdentity(t, dir)

	WriteFile(t, dir, "file.txt", "first version\n")
	RunGit(t, dir, "add", ".")
	RunGit(t, dir, "commit", "-m", "first commit")
	RunGit(t, dir, "tag", "v1.0.0")
	RunGit(t, dir, "branch", "feature")

	WriteFile(t, dir, "file.txt", "second version\n")
	RunGit(t, dir, "add", ".")
	RunGit(t, dir, "commit", "-m", "second commit")

	return dir
}

// FileURL converts a local directory path to a file:// URL. Git silently
// ignores --depth for plain local-path clones; the file:// transport
// honors it.
func FileURL(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + p
}

// CloneRepo clones src into a new temp directory and returns the clone path.
func CloneRepo(t *testing.T, src string) string {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "clone")
	// autocrlf must be off *before* the initial checkout: Git for
	// Windows defaults to true, which writes CRLF to the working tree,
	// and flipping the setting afterwards makes every file look
	// modified against the LF blobs in the index.
	cmd := exec.Command("git", "clone", "-c", "core.autocrlf=false", src, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git clone %s failed: %v\n%s", src, err, out)
	}
	configureIdentity(t, dest)
	return dest
}

// WriteFile writes content to dir/name, failing the test on error.
func WriteFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// CaptureStdout runs fn with os.Stdout redirected to a pipe and returns
// everything written to stdout. fn must not panic; recover inside fn if
// the code under test may panic.
func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-done
}

// Chdir changes the working directory for the duration of the test.
func Chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}
