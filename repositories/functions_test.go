package repositories_test

import (
	"testing"

	"multirepo/internal/testutil"
	"multirepo/repositories"
)

func TestParseConfigValid(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "repositories.yaml", `
repositories:
  fastapi:
    path: /tmp/multirepo/fastapi
    url: https://github.com/tiangolo/fastapi.git
    tag: 0.111.0
  pytest:
    path: /tmp/multirepo/pytest
    url: https://github.com/pytest-dev/pytest.git
    branch: main
  pydantic:
    path: /tmp/multirepo/pydantic
    url: https://github.com/pydantic/pydantic.git
    commit: 7061f36
`)
	testutil.Chdir(t, dir)

	config, err := repositories.ParseConfig()
	if err != nil {
		t.Fatalf("ParseConfig returned error: %v", err)
	}

	if len(config.Repos) != 3 {
		t.Fatalf("expected 3 repositories, got %d", len(config.Repos))
	}

	fastapi := config.Repos["fastapi"]
	if fastapi.Path != "/tmp/multirepo/fastapi" {
		t.Errorf("fastapi.Path = %q, want /tmp/multirepo/fastapi", fastapi.Path)
	}
	if fastapi.URL != "https://github.com/tiangolo/fastapi.git" {
		t.Errorf("fastapi.URL = %q", fastapi.URL)
	}
	if fastapi.Tag != "0.111.0" {
		t.Errorf("fastapi.Tag = %q, want 0.111.0", fastapi.Tag)
	}

	if config.Repos["pytest"].Branch != "main" {
		t.Errorf("pytest.Branch = %q, want main", config.Repos["pytest"].Branch)
	}
	if config.Repos["pydantic"].Commit != "7061f36" {
		t.Errorf("pydantic.Commit = %q, want 7061f36", config.Repos["pydantic"].Commit)
	}
}

func TestParseConfigMissingFile(t *testing.T) {
	testutil.Chdir(t, t.TempDir())

	_, err := repositories.ParseConfig()
	if err == nil {
		t.Error("expected an error when repositories.yaml does not exist, got nil")
	}
}

func TestParseConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "repositories.yaml", "repositories: [not: valid: yaml\n")
	testutil.Chdir(t, dir)

	_, err := repositories.ParseConfig()
	if err == nil {
		t.Error("expected an error for invalid YAML, got nil")
	}
}

func TestParseConfigEmptyFile(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "repositories.yaml", "")
	testutil.Chdir(t, dir)

	config, err := repositories.ParseConfig()
	if err != nil {
		t.Fatalf("ParseConfig returned error for empty file: %v", err)
	}
	if len(config.Repos) != 0 {
		t.Errorf("expected 0 repositories for empty file, got %d", len(config.Repos))
	}
}

func TestParseTargetCommit(t *testing.T) {
	target, err := repositories.ParseTarget(repositories.Repository{Commit: "abc1234"})
	if err != nil {
		t.Fatalf("ParseTarget returned error: %v", err)
	}
	if target.Type != "commit" || target.Name != "abc1234" || target.DisplayName != "abc1234" {
		t.Errorf("got %+v, want commit/abc1234/abc1234", target)
	}
}

func TestParseTargetTag(t *testing.T) {
	target, err := repositories.ParseTarget(repositories.Repository{Tag: "v1.0.0"})
	if err != nil {
		t.Fatalf("ParseTarget returned error: %v", err)
	}
	if target.Type != "tag" || target.Name != "v1.0.0" {
		t.Errorf("got %+v, want tag/v1.0.0", target)
	}
}

func TestParseTargetBranch(t *testing.T) {
	target, err := repositories.ParseTarget(repositories.Repository{Branch: "main"})
	if err != nil {
		t.Fatalf("ParseTarget returned error: %v", err)
	}
	if target.Type != "branch" || target.Name != "main" {
		t.Errorf("got %+v, want branch/main", target)
	}
}

// Commit takes precedence over tag, and tag over branch, when several
// references are defined for the same repository.
func TestParseTargetPrecedence(t *testing.T) {
	target, err := repositories.ParseTarget(repositories.Repository{
		Commit: "abc1234",
		Tag:    "v1.0.0",
		Branch: "main",
	})
	if err != nil {
		t.Fatalf("ParseTarget returned error: %v", err)
	}
	if target.Type != "commit" {
		t.Errorf("commit should take precedence, got type %q", target.Type)
	}

	target, err = repositories.ParseTarget(repositories.Repository{
		Tag:    "v1.0.0",
		Branch: "main",
	})
	if err != nil {
		t.Fatalf("ParseTarget returned error: %v", err)
	}
	if target.Type != "tag" {
		t.Errorf("tag should take precedence over branch, got type %q", target.Type)
	}
}

func TestParseTargetMissingReference(t *testing.T) {
	_, err := repositories.ParseTarget(repositories.Repository{
		Path: "/tmp/foo",
		URL:  "https://example.com/foo.git",
	})
	if err == nil {
		t.Error("expected an error when no commit, tag, or branch is set, got nil")
	}
}

func TestApplyLockOverridesTargets(t *testing.T) {
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"alpha": {Path: "/tmp/alpha", URL: "u", Branch: "main"},
		"beta":  {Path: "/tmp/beta", URL: "u", Tag: "v1.0.0"},
	}}
	lock := repositories.LockConfig{Repos: map[string]repositories.LockedRepository{
		"alpha": {Commit: "1111111111111111111111111111111111111111"},
		"beta":  {Commit: "2222222222222222222222222222222222222222"},
	}}

	locked, err := repositories.ApplyLock(config, lock)
	if err != nil {
		t.Fatalf("ApplyLock returned error: %v", err)
	}

	alpha := locked.Repos["alpha"]
	if alpha.Commit != "1111111111111111111111111111111111111111" || alpha.Branch != "" || alpha.Tag != "" {
		t.Errorf("alpha after ApplyLock = %+v, want commit-only target", alpha)
	}
	beta := locked.Repos["beta"]
	if beta.Commit != "2222222222222222222222222222222222222222" || beta.Tag != "" {
		t.Errorf("beta after ApplyLock = %+v, want commit-only target", beta)
	}
	if beta.Path != "/tmp/beta" || beta.URL != "u" {
		t.Errorf("beta after ApplyLock = %+v, want path and url preserved", beta)
	}
}

func TestApplyLockMissingEntry(t *testing.T) {
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"alpha": {Path: "/tmp/alpha", URL: "u", Branch: "main"},
	}}
	lock := repositories.LockConfig{Repos: map[string]repositories.LockedRepository{}}

	_, err := repositories.ApplyLock(config, lock)
	if err == nil {
		t.Fatal("expected an error for a repository missing from the lock file, got nil")
	}
}

func TestParseLockMissingFile(t *testing.T) {
	testutil.Chdir(t, t.TempDir())

	_, err := repositories.ParseLock()
	if err == nil {
		t.Fatal("expected an error when repositories.lock does not exist, got nil")
	}
}
