package git_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"multirepo/git"
	"multirepo/internal/testutil"
	"multirepo/repositories"
)

func repoAt(path string) repositories.Repository {
	return repositories.Repository{Path: path}
}

func TestExistsTrueForGitRepository(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)

	if !git.Exists(repoAt(origin)) {
		t.Errorf("Exists = false for a valid git repository at %s", origin)
	}
}

func TestExistsFalseForPlainDirectory(t *testing.T) {
	dir := t.TempDir()

	if git.Exists(repoAt(dir)) {
		t.Errorf("Exists = true for a directory that is not a git repository: %s", dir)
	}
}

func TestExistsFalseForMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist")

	if git.Exists(repoAt(path)) {
		t.Errorf("Exists = true for a path that does not exist: %s", path)
	}
}

// Exists ran `git status` in the configured path, which succeeds for any
// directory located *inside* some other git repository (git walks up to
// find the enclosing work tree). Sync would then skip the clone and
// happily checkout refs in the enclosing repository.
func TestExistsFalseForDirectoryInsideAnotherRepository(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	nested := filepath.Join(origin, "vendor", "dependency")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if git.Exists(repoAt(nested)) {
		t.Errorf("Exists = true for a plain directory nested inside another git repository: %s", nested)
	}
}

func TestIsDirtyCleanRepository(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)

	dirty, err := git.IsDirty(repoAt(origin))
	if err != nil {
		t.Fatalf("IsDirty returned error: %v", err)
	}
	if dirty {
		t.Error("IsDirty = true for a freshly committed repository")
	}
}

func TestIsDirtyModifiedFile(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.WriteFile(t, clone, "file.txt", "modified content\n")

	dirty, err := git.IsDirty(repoAt(clone))
	if err != nil {
		t.Fatalf("IsDirty returned error: %v", err)
	}
	if !dirty {
		t.Error("IsDirty = false for a repository with a modified tracked file")
	}
}

func TestIsDirtyUntrackedFile(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.WriteFile(t, clone, "new-file.txt", "untracked\n")

	dirty, err := git.IsDirty(repoAt(clone))
	if err != nil {
		t.Fatalf("IsDirty returned error: %v", err)
	}
	if !dirty {
		t.Error("IsDirty = false for a repository with an untracked file")
	}
}

// IsDirty greps the human-readable `git status --long` output for the
// English phrase "working tree clean". That breaks whenever the phrase
// appears for another reason (here: an untracked file whose name contains
// it) and on any non-English locale, where the phrase never appears and
// every clean repository is reported dirty. Parsing `--porcelain` output
// is locale-independent and unambiguous.
func TestIsDirtyDoesNotParseHumanReadableOutput(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.WriteFile(t, clone, "nothing to commit, working tree clean", "untracked\n")

	dirty, err := git.IsDirty(repoAt(clone))
	if err != nil {
		t.Fatalf("IsDirty returned error: %v", err)
	}
	if !dirty {
		t.Error("IsDirty = false for a dirty repository whose untracked file name contains the phrase 'working tree clean'")
	}
}

func TestIsDirtyMissingRepository(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := git.IsDirty(repoAt(path))
	if err == nil {
		t.Error("expected an error for a missing repository, got nil")
	}
}

func TestGetCurrentCommit(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	want := testutil.RunGit(t, origin, "rev-parse", "HEAD")

	commit, err := git.GetCurrentCommit(repoAt(origin))
	if err != nil {
		t.Fatalf("GetCurrentCommit returned error: %v", err)
	}
	if commit != want {
		t.Errorf("GetCurrentCommit = %q, want %q", commit, want)
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(commit) {
		t.Errorf("GetCurrentCommit = %q, want a full 40-character hash", commit)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)

	branch, err := git.GetCurrentBranch(repoAt(origin))
	if err != nil {
		t.Fatalf("GetCurrentBranch returned error: %v", err)
	}
	if branch != "main" {
		t.Errorf("GetCurrentBranch = %q, want main", branch)
	}
}

func TestGetCurrentBranchDetachedHead(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.RunGit(t, clone, "checkout", "v1.0.0")

	branch, err := git.GetCurrentBranch(repoAt(clone))
	if err != nil {
		t.Fatalf("GetCurrentBranch returned error: %v", err)
	}
	if branch != "" {
		t.Errorf("GetCurrentBranch = %q on a detached HEAD, want empty string", branch)
	}
}

func TestGetCurrentTagsOnTaggedCommit(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.RunGit(t, clone, "checkout", "v1.0.0")

	tags, err := git.GetCurrentTags(repoAt(clone))
	if err != nil {
		t.Fatalf("GetCurrentTags returned error: %v", err)
	}
	if tags != "v1.0.0" {
		t.Errorf("GetCurrentTags = %q, want v1.0.0", tags)
	}
}

func TestGetCurrentTagsOnUntaggedCommit(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)

	tags, err := git.GetCurrentTags(repoAt(origin))
	if err != nil {
		t.Fatalf("GetCurrentTags returned error: %v", err)
	}
	if tags != "" {
		t.Errorf("GetCurrentTags = %q on an untagged commit, want empty string", tags)
	}
}

// Documents the FIXME in commands/functions.go: when several tags point at
// HEAD, GetCurrentTags returns all of them joined by newlines, which the
// status command then compares against a single tag name.
func TestGetCurrentTagsMultipleTags(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.RunGit(t, clone, "checkout", "v1.0.0")
	testutil.RunGit(t, clone, "tag", "extra-tag")

	tags, err := git.GetCurrentTags(repoAt(clone))
	if err != nil {
		t.Fatalf("GetCurrentTags returned error: %v", err)
	}
	if !strings.Contains(tags, "v1.0.0") || !strings.Contains(tags, "extra-tag") {
		t.Errorf("GetCurrentTags = %q, want both v1.0.0 and extra-tag listed", tags)
	}
}

func TestClone(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	repo := repositories.Repository{Path: dest, URL: origin}

	output, err := git.Clone(repo, false)
	if err != nil {
		t.Errorf("Clone returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "file.txt")); err != nil {
		t.Errorf("cloned repository is missing file.txt: %v", err)
	}
	if !git.Exists(repo) {
		t.Error("Exists = false after Clone")
	}
	if !strings.Contains(output, "git clone") {
		t.Errorf("Clone output does not echo the command, got: %q", output)
	}
}

func TestCloneRecurse(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	repo := repositories.Repository{Path: dest, URL: origin}

	if _, err := git.Clone(repo, true); err != nil {
		t.Errorf("Clone with recurse returned error: %v", err)
	}

	if !git.Exists(repo) {
		t.Error("Exists = false after Clone with recurse")
	}
}

func TestCloneWithDepth(t *testing.T) {
	origin := testutil.CreateOriginRepo(t) // main has two commits
	dest := filepath.Join(t.TempDir(), "clone")
	repo := repositories.Repository{Path: dest, URL: testutil.FileURL(origin), Depth: 1}

	testutil.CaptureStdout(t, func() {
		if _, err := git.Clone(repo, false); err != nil {
			t.Fatalf("Clone with depth returned error: %v", err)
		}
	})

	count := testutil.RunGit(t, dest, "rev-list", "--count", "HEAD")
	if count != "1" {
		t.Errorf("rev-list --count HEAD = %s after a depth-1 clone, want 1", count)
	}
}

func TestCloneWithoutDepthIsFull(t *testing.T) {
	origin := testutil.CreateOriginRepo(t) // main has two commits
	dest := filepath.Join(t.TempDir(), "clone")
	repo := repositories.Repository{Path: dest, URL: testutil.FileURL(origin)}

	testutil.CaptureStdout(t, func() {
		if _, err := git.Clone(repo, false); err != nil {
			t.Fatalf("Clone returned error: %v", err)
		}
	})

	count := testutil.RunGit(t, dest, "rev-list", "--count", "HEAD")
	if count != "2" {
		t.Errorf("rev-list --count HEAD = %s after a default clone, want the full history (2)", count)
	}
}

func TestCloneInvalidURL(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "clone")
	repo := repositories.Repository{
		Path: dest,
		URL:  filepath.Join(t.TempDir(), "no-such-origin"),
	}

	if _, err := git.Clone(repo, false); err == nil {
		t.Error("expected an error when cloning a nonexistent origin, got nil")
	}
}

func TestCheckoutTag(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	repo := repositories.Repository{Path: clone, URL: origin, Tag: "v1.0.0"}

	if _, err := git.Checkout(repo, false); err != nil {
		t.Errorf("Checkout returned error: %v", err)
	}

	want := testutil.RunGit(t, origin, "rev-parse", "v1.0.0^{commit}")
	got := testutil.RunGit(t, clone, "rev-parse", "HEAD")
	if got != want {
		t.Errorf("HEAD after checkout = %s, want tag commit %s", got, want)
	}
}

func TestCheckoutBranch(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	repo := repositories.Repository{Path: clone, URL: origin, Branch: "feature"}

	if _, err := git.Checkout(repo, false); err != nil {
		t.Errorf("Checkout returned error: %v", err)
	}

	got := testutil.RunGit(t, clone, "branch", "--show-current")
	if got != "feature" {
		t.Errorf("current branch after checkout = %q, want feature", got)
	}
}

func TestCheckoutCommit(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	want := testutil.RunGit(t, origin, "rev-parse", "v1.0.0^{commit}")
	repo := repositories.Repository{Path: clone, URL: origin, Commit: want}

	if _, err := git.Checkout(repo, false); err != nil {
		t.Errorf("Checkout returned error: %v", err)
	}

	got := testutil.RunGit(t, clone, "rev-parse", "HEAD")
	if got != want {
		t.Errorf("HEAD after checkout = %s, want %s", got, want)
	}
}

func TestCheckoutMissingReference(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	repo := repositories.Repository{Path: clone, URL: origin, Branch: "does-not-exist"}

	if _, err := git.Checkout(repo, false); err == nil {
		t.Error("expected an error when checking out a missing reference, got nil")
	}
}

func TestStashAndStashDrop(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.WriteFile(t, clone, "file.txt", "dirty content\n")
	repo := repoAt(clone)

	if _, err := git.Stash(repo); err != nil {
		t.Fatalf("Stash returned error: %v", err)
	}

	dirty, err := git.IsDirty(repo)
	if err != nil {
		t.Fatalf("IsDirty returned error: %v", err)
	}
	if dirty {
		t.Error("repository is still dirty after Stash")
	}

	if _, err := git.StashDrop(repo); err != nil {
		t.Errorf("StashDrop returned error: %v", err)
	}

	stashList := testutil.RunGit(t, clone, "stash", "list")
	if stashList != "" {
		t.Errorf("stash list is not empty after StashDrop: %q", stashList)
	}
}

func TestStashUntrackedFile(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))
	testutil.WriteFile(t, clone, "new-file.txt", "untracked\n")
	repo := repoAt(clone)

	if _, err := git.Stash(repo); err != nil {
		t.Fatalf("Stash returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(clone, "new-file.txt")); !os.IsNotExist(err) {
		t.Error("untracked file still present after Stash -u")
	}
}

func TestStashDropWithoutStash(t *testing.T) {
	clone := testutil.CloneRepo(t, testutil.CreateOriginRepo(t))

	if _, err := git.StashDrop(repoAt(clone)); err == nil {
		t.Error("expected an error when dropping a nonexistent stash, got nil")
	}
}
