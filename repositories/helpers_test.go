package repositories_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"multirepo/repositories"
)

func TestResolvePathPlain(t *testing.T) {
	path, err := repositories.ResolvePath("/tmp/multirepo/fastapi")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}
	if path != "/tmp/multirepo/fastapi" {
		t.Errorf("got %q, want /tmp/multirepo/fastapi", path)
	}
}

func TestResolvePathEnvVar(t *testing.T) {
	t.Setenv("MY_BASE_DIRECTORY", "/srv/repos")

	path, err := repositories.ResolvePath("$MY_BASE_DIRECTORY/fastapi")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}
	if path != "/srv/repos/fastapi" {
		t.Errorf("got %q, want /srv/repos/fastapi", path)
	}
}

func TestResolvePathEnvVarBraces(t *testing.T) {
	t.Setenv("MY_BASE_DIRECTORY", "/srv/repos")

	path, err := repositories.ResolvePath("${MY_BASE_DIRECTORY}/fastapi")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}
	if path != "/srv/repos/fastapi" {
		t.Errorf("got %q, want /srv/repos/fastapi", path)
	}
}

func TestResolvePathTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	path, err := repositories.ResolvePath("~/repos/fastapi")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}
	want := filepath.Join(home, "repos/fastapi")
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
}

// A bare "~" is conventionally the user's home directory, but ResolvePath
// only handles the "~/" prefix, so "~" is returned untouched and later
// treated as a relative directory literally named "~".
func TestResolvePathBareTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	path, err := repositories.ResolvePath("~")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}
	if path != home {
		t.Errorf("ResolvePath(%q) = %q, want home directory %q", "~", path, home)
	}
}

// ResolvePath calls log.Fatal when a path references an undefined
// environment variable, which terminates the whole process instead of
// returning an error to the caller. This test re-executes the test binary
// so the fatal exit doesn't kill the test run, and fails if the
// subprocess exits non-zero.
func TestResolvePathUndefinedEnvVarReturnsError(t *testing.T) {
	if os.Getenv("MULTIREPO_TEST_RESOLVE_SUBPROCESS") == "1" {
		_, _ = repositories.ResolvePath("$MULTIREPO_UNDEFINED_VARIABLE_12345/repo")
		// Only reached if ResolvePath returns instead of exiting.
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestResolvePathUndefinedEnvVarReturnsError")
	cmd.Env = append(os.Environ(), "MULTIREPO_TEST_RESOLVE_SUBPROCESS=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("ResolvePath terminated the process on an undefined environment variable instead of returning an error: %v\n%s", err, out)
	}
}
