package commands_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"multirepo/commands"
	"multirepo/git"
	"multirepo/internal/testutil"
	"multirepo/repositories"
)

func singleRepoConfig(name string, repo repositories.Repository) repositories.Config {
	return repositories.Config{Repos: map[string]repositories.Repository{name: repo}}
}

// --- Status ---

func TestStatusCleanTag(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	testutil.RunGit(t, clone, "checkout", "v1.0.0")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Tag: "v1.0.0",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if !strings.Contains(output, "✔") {
		t.Errorf("expected ✔ for a clean repository on the right tag, got:\n%s", output)
	}
	if !strings.Contains(output, "(tag: v1.0.0)") {
		t.Errorf("expected '(tag: v1.0.0)' in output, got:\n%s", output)
	}
	if strings.Contains(output, "✗") {
		t.Errorf("unexpected ✗ in output:\n%s", output)
	}
}

func TestStatusCleanBranch(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if !strings.Contains(output, "✔") || !strings.Contains(output, "(branch: main)") {
		t.Errorf("expected ✔ and '(branch: main)', got:\n%s", output)
	}
}

func TestStatusCleanCommit(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	head := testutil.RunGit(t, clone, "rev-parse", "HEAD")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Commit: head,
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if !strings.Contains(output, "✔") || !strings.Contains(output, head) {
		t.Errorf("expected ✔ and commit hash %s, got:\n%s", head, output)
	}
}

func TestStatusWrongBranch(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // checked out on main
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "feature",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if !strings.Contains(output, "✗") {
		t.Errorf("expected ✗ when on the wrong branch, got:\n%s", output)
	}
	if !strings.Contains(output, "branch: feature ➜") {
		t.Errorf("expected 'branch: feature ➜ ...' mismatch marker, got:\n%s", output)
	}
}

func TestStatusUncommittedChanges(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	testutil.WriteFile(t, clone, "file.txt", "dirty content\n")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if !strings.Contains(output, "(uncommitted changes)") {
		t.Errorf("expected '(uncommitted changes)' for a dirty repository, got:\n%s", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("expected ✗ for a dirty repository, got:\n%s", output)
	}
}

func TestStatusRepositoryNotFound(t *testing.T) {
	config := singleRepoConfig("ghost", repositories.Repository{
		Path: filepath.Join(t.TempDir(), "does-not-exist"),
		URL:  "https://example.com/ghost.git",
		Tag:  "v1.0.0",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if !strings.Contains(output, "repository not found") {
		t.Errorf("expected 'repository not found' for a missing repository, got:\n%s", output)
	}
}

// Git accepts abbreviated commit hashes (4+ characters), and the README
// itself uses a 7-character commit in its example. Status slices the
// configured commit with [:7], which panics on anything shorter.
func TestStatusShortCommitHashDoesNotPanic(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	head := testutil.RunGit(t, clone, "rev-parse", "HEAD")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Commit: head[:5],
	})

	var panicked interface{}
	testutil.CaptureStdout(t, func() {
		defer func() { panicked = recover() }()
		_ = commands.Status(config, "")
	})

	if panicked != nil {
		t.Errorf("Status panicked on a 5-character commit hash (valid git abbreviation): %v", panicked)
	}
}

// Exposes the FIXME in Status: git.GetCurrentTags returns every tag
// pointing at HEAD joined by newlines, so a repository on the correct tag
// is reported as a mismatch as soon as a second tag points at the same
// commit.
func TestStatusMultipleTagsOnTargetCommit(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	testutil.RunGit(t, clone, "checkout", "v1.0.0")
	testutil.RunGit(t, clone, "tag", "extra-tag")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Tag: "v1.0.0",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if strings.Contains(output, "✗") {
		t.Errorf("repository is on the target tag v1.0.0 but Status reports a mismatch because another tag points at the same commit:\n%s", output)
	}
}

// Status checks GetCurrentCommit's error but silently discards the
// errors from IsDirty, GetCurrentTags, and GetCurrentBranch, printing a
// row computed from garbage instead of failing. A bare repository is a
// deterministic reproduction: `git log` succeeds there but `git status`
// fails.
func TestStatusReturnsErrorWhenGitStatusFails(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	bare := filepath.Join(t.TempDir(), "bare.git")
	cmd := exec.Command("git", "clone", "--bare", origin, bare)
	if out, cloneErr := cmd.CombinedOutput(); cloneErr != nil {
		t.Fatalf("git clone --bare failed: %v\n%s", cloneErr, out)
	}
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: bare, URL: origin, Branch: "main",
	})

	var err error
	testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "")
	})

	if err == nil {
		t.Error("Status = nil for a repository where `git status` fails (bare repository); the IsDirty error was silently discarded")
	}
}

// --- Sync ---

func TestSyncClonesMissingRepository(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	dest := filepath.Join(t.TempDir(), "myrepo")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: dest, URL: origin, Tag: "v1.0.0",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, false, false)
	})
	if err != nil {
		t.Fatalf("Sync returned error: %v\n%s", err, output)
	}

	if _, statErr := os.Stat(filepath.Join(dest, "file.txt")); statErr != nil {
		t.Fatalf("repository was not cloned: %v", statErr)
	}

	wantCommit := testutil.RunGit(t, origin, "rev-parse", "v1.0.0^{commit}")
	gotCommit := testutil.RunGit(t, dest, "rev-parse", "HEAD")
	if gotCommit != wantCommit {
		t.Errorf("HEAD after sync = %s, want tag commit %s", gotCommit, wantCommit)
	}
	if !strings.Contains(output, "git clone") {
		t.Errorf("expected clone command echoed in output, got:\n%s", output)
	}
}

func TestSyncChecksOutTargetReference(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // on main
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "feature",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, false, false)
	})
	if err != nil {
		t.Fatalf("Sync returned error: %v\n%s", err, output)
	}

	got := testutil.RunGit(t, clone, "branch", "--show-current")
	if got != "feature" {
		t.Errorf("current branch after sync = %q, want feature", got)
	}
}

func TestSyncIsIdempotent(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	dest := filepath.Join(t.TempDir(), "myrepo")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: dest, URL: origin, Tag: "v1.0.0",
	})

	for i := 0; i < 2; i++ {
		var err error
		output := testutil.CaptureStdout(t, func() {
			err = commands.Sync(config, false, false)
		})
		if err != nil {
			t.Fatalf("Sync run %d returned error: %v\n%s", i+1, err, output)
		}
	}
}

// The README states: "If the repository is dirty, we'll abort the
// operation." Sync never checks IsDirty, so a dirty repository is synced
// (and reported clean by checkout) without aborting.
func TestSyncAbortsOnDirtyRepository(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // on main
	testutil.WriteFile(t, clone, "file.txt", "uncommitted local work\n")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, false, false)
	})

	if err == nil {
		t.Error("README says sync aborts when the repository is dirty, but Sync returned nil for a repository with uncommitted changes")
	}
}

// Sync never fetched, so a tag (or commit) created in origin after the
// initial clone could not be checked out: sync failed and the README
// told users to fetch by hand.
func TestSyncFetchesNewTagFromOrigin(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // on main
	testutil.WriteFile(t, origin, "file.txt", "third version\n")
	testutil.RunGit(t, origin, "add", ".")
	testutil.RunGit(t, origin, "commit", "-m", "third commit")
	testutil.RunGit(t, origin, "tag", "v2.0.0")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Tag: "v2.0.0",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, false, false)
	})
	if err != nil {
		t.Fatalf("Sync returned error for a tag that exists in origin but not locally: %v\n%s", err, output)
	}

	want := testutil.RunGit(t, origin, "rev-parse", "v2.0.0^{commit}")
	got := testutil.RunGit(t, clone, "rev-parse", "HEAD")
	if got != want {
		t.Errorf("HEAD after sync = %s, want tag commit %s", got, want)
	}
}

// A repository pinned to a branch never advanced: sync checked out the
// branch but never fetched, so the local branch stayed on whatever
// commit it had at clone time.
func TestSyncFastForwardsBranchToOrigin(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // on main
	testutil.WriteFile(t, origin, "file.txt", "third version\n")
	testutil.RunGit(t, origin, "add", ".")
	testutil.RunGit(t, origin, "commit", "-m", "third commit")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, false, false)
	})
	if err != nil {
		t.Fatalf("Sync returned error: %v\n%s", err, output)
	}

	want := testutil.RunGit(t, origin, "rev-parse", "main")
	got := testutil.RunGit(t, clone, "rev-parse", "HEAD")
	if got != want {
		t.Errorf("HEAD after sync = %s, want origin's main %s (branch was not fast-forwarded)", got, want)
	}
}

// Guard for the design choice: tag and commit targets that are already
// available locally must not require network access.
func TestSyncTagTargetWorksWithoutRemote(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // on main, v1.0.0 fetched
	if err := os.RemoveAll(origin); err != nil {
		t.Fatal(err)
	}
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Tag: "v1.0.0",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, false, false)
	})
	if err != nil {
		t.Fatalf("Sync should not need the remote for a locally available tag: %v\n%s", err, output)
	}
}

func TestSyncForceDiscardsUncommittedChanges(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // on main, file.txt = "second version"
	testutil.WriteFile(t, clone, "file.txt", "uncommitted local work\n")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, true, false)
	})
	if err != nil {
		t.Fatalf("Sync --force returned error: %v\n%s", err, output)
	}

	content, readErr := os.ReadFile(filepath.Join(clone, "file.txt"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "second version\n" {
		t.Errorf("file.txt = %q after sync --force, want committed content %q", content, "second version\n")
	}

	dirty, dirtyErr := git.IsDirty(repositories.Repository{Path: clone})
	if dirtyErr != nil {
		t.Fatal(dirtyErr)
	}
	if dirty {
		t.Error("repository is still dirty after sync --force")
	}
}

// Sync --force runs `git stash -u` followed by an unconditional `git
// stash drop`. On a clean working tree the stash creates nothing, so the
// drop deletes the user's most recent pre-existing stash entry instead.
func TestSyncForcePreservesExistingStashOnCleanTree(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin) // on main
	testutil.WriteFile(t, clone, "file.txt", "precious stashed work\n")
	testutil.RunGit(t, clone, "stash") // working tree is clean again
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Sync(config, true, false)
	})
	if err != nil {
		t.Fatalf("Sync --force returned error: %v\n%s", err, output)
	}

	stashList := testutil.RunGit(t, clone, "stash", "list")
	if stashList == "" {
		t.Error("sync --force on a clean working tree dropped a pre-existing stash entry")
	}
}

// --- Run ---

func TestRunSingleRepository(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Run(config, "myrepo", "git", []string{"status"})
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !strings.Contains(output, "➜ "+clone+"$ git status") {
		t.Errorf("expected command echo for %s, got:\n%s", clone, output)
	}
	if !strings.Contains(output, "working tree clean") {
		t.Errorf("expected git status output, got:\n%s", output)
	}
}

func TestRunAllRepositories(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	cloneA := testutil.CloneRepo(t, origin)
	cloneB := testutil.CloneRepo(t, origin)
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"alpha": {Path: cloneA, URL: origin, Branch: "main"},
		"beta":  {Path: cloneB, URL: origin, Branch: "main"},
	}}

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Run(config, "--all", "git", []string{"status"})
	})
	if err != nil {
		t.Fatalf("Run --all returned error: %v", err)
	}

	if count := strings.Count(output, "➜ "); count != 2 {
		t.Errorf("expected the command to run in 2 repositories, found %d command echoes:\n%s", count, output)
	}
}

// Run with --all returned on the first command failure, silently
// skipping the remaining repositories. It should run the command
// everywhere and report the failures at the end.
func TestRunAllContinuesAfterFailure(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	cloneA := testutil.CloneRepo(t, origin)
	cloneB := testutil.CloneRepo(t, origin)
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"alpha": {Path: cloneA, URL: origin, Branch: "main"},
		"beta":  {Path: cloneB, URL: origin, Branch: "main"},
	}}

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Run(config, "--all", "git", []string{"not-a-real-subcommand"})
	})

	if count := strings.Count(output, "➜ "); count != 2 {
		t.Errorf("expected the failing command to be attempted in 2 repositories, found %d command echoes:\n%s", count, output)
	}
	if err == nil {
		t.Fatal("expected an error when the command fails, got nil")
	}
	if !strings.Contains(err.Error(), "alpha") || !strings.Contains(err.Error(), "beta") {
		t.Errorf("error = %q, want it to name the failing repositories alpha and beta", err)
	}
}

func TestRunUnknownRepository(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	testutil.CaptureStdout(t, func() {
		err = commands.Run(config, "nope", "git", []string{"status"})
	})

	if err == nil {
		t.Fatal("expected an error for an unknown repository name, got nil")
	}
	if !strings.Contains(err.Error(), "Unknown repository 'nope'") {
		t.Errorf("error = %q, want it to mention \"Unknown repository 'nope'\"", err)
	}
}

// Run calls log.Fatal when the executed command exits non-zero, which
// kills the whole multirepo process (skipping remaining repositories with
// --all) instead of returning an error. This test re-executes the test
// binary so the fatal exit doesn't kill the test run.
func TestRunFailingCommandReturnsError(t *testing.T) {
	if os.Getenv("MULTIREPO_TEST_RUN_SUBPROCESS") == "1" {
		dir, err := os.MkdirTemp("", "multirepo-run")
		if err != nil {
			os.Exit(2)
		}
		defer os.RemoveAll(dir)

		config := singleRepoConfig("myrepo", repositories.Repository{
			Path: dir, URL: "unused", Branch: "main",
		})
		_ = commands.Run(config, "myrepo", "git", []string{"not-a-real-subcommand"})
		// Only reached if Run returns instead of exiting.
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunFailingCommandReturnsError")
	cmd.Env = append(os.Environ(), "MULTIREPO_TEST_RUN_SUBPROCESS=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Run terminated the process when the executed command failed instead of returning an error: %v\n%s", err, out)
	}
}

// --- Status output formats ---

func TestStatusJSONOutput(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	testutil.RunGit(t, clone, "checkout", "v1.0.0")
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"myrepo": {Path: clone, URL: origin, Tag: "v1.0.0"},
		"ghost":  {Path: filepath.Join(t.TempDir(), "missing"), URL: origin, Tag: "v1.0.0"},
	}}

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "json")
	})
	if err != nil {
		t.Fatalf("Status json returned error: %v", err)
	}

	var rows []commands.StatusRow
	if jsonErr := json.Unmarshal([]byte(output), &rows); jsonErr != nil {
		t.Fatalf("status --json did not produce valid JSON: %v\noutput:\n%s", jsonErr, output)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}

	// Rows are ordered by repository name: ghost, myrepo.
	ghost, myrepo := rows[0], rows[1]
	if ghost.Name != "ghost" || ghost.Found {
		t.Errorf("ghost row = %+v, want found=false", ghost)
	}
	want := testutil.RunGit(t, clone, "rev-parse", "HEAD")
	if myrepo.Name != "myrepo" || !myrepo.Found || myrepo.Commit != want {
		t.Errorf("myrepo row = %+v, want found row at commit %s", myrepo, want)
	}
	if !myrepo.InSync || myrepo.Dirty {
		t.Errorf("myrepo row = %+v, want in_sync=true dirty=false", myrepo)
	}
	if myrepo.TargetType != "tag" || myrepo.TargetName != "v1.0.0" {
		t.Errorf("myrepo target = %s %s, want tag v1.0.0", myrepo.TargetType, myrepo.TargetName)
	}
}

func TestStatusMarkdownOutput(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	testutil.WriteFile(t, clone, "file.txt", "dirty content\n")
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	output := testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "md")
	})
	if err != nil {
		t.Fatalf("Status md returned error: %v", err)
	}

	if !strings.Contains(output, "| repository |") {
		t.Errorf("expected a markdown table header, got:\n%s", output)
	}
	if !strings.Contains(output, "| myrepo ") {
		t.Errorf("expected a row for myrepo, got:\n%s", output)
	}
	if strings.Contains(output, "\x1b[") {
		t.Errorf("markdown output must not contain ANSI color codes, got:\n%s", output)
	}
}

func TestStatusUnknownFormat(t *testing.T) {
	origin := testutil.CreateOriginRepo(t)
	clone := testutil.CloneRepo(t, origin)
	config := singleRepoConfig("myrepo", repositories.Repository{
		Path: clone, URL: origin, Branch: "main",
	})

	var err error
	testutil.CaptureStdout(t, func() {
		err = commands.Status(config, "xml")
	})
	if err == nil {
		t.Error("expected an error for an unknown status format, got nil")
	}
}
