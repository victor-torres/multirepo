package commands_test

import (
	"reflect"
	"strings"
	"testing"

	"multirepo/commands"
	"multirepo/internal/testutil"
	"multirepo/repositories"
)

func TestPrintRepositoryCounterSingular(t *testing.T) {
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"only": {},
	}}

	output := testutil.CaptureStdout(t, func() {
		commands.PrintRepositoryCounter(config)
	})

	if !strings.Contains(output, "1 repository detected") {
		t.Errorf("got %q, want singular '1 repository detected'", output)
	}
}

func TestPrintRepositoryCounterPlural(t *testing.T) {
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"first":  {},
		"second": {},
	}}

	output := testutil.CaptureStdout(t, func() {
		commands.PrintRepositoryCounter(config)
	})

	if !strings.Contains(output, "2 repositories detected") {
		t.Errorf("got %q, want plural '2 repositories detected'", output)
	}
}

func TestPrintRepositoryCounterZero(t *testing.T) {
	config := repositories.Config{}

	output := testutil.CaptureStdout(t, func() {
		commands.PrintRepositoryCounter(config)
	})

	if !strings.Contains(output, "0 repositories detected") {
		t.Errorf("got %q, want '0 repositories detected'", output)
	}
}

func TestGetOrderedRepoNames(t *testing.T) {
	config := repositories.Config{Repos: map[string]repositories.Repository{
		"pytest":   {},
		"fastapi":  {},
		"pydantic": {},
	}}

	names := commands.GetOrderedRepoNames(config)
	want := []string{"fastapi", "pydantic", "pytest"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("got %v, want %v", names, want)
	}
}

func TestGetOrderedRepoNamesEmpty(t *testing.T) {
	names := commands.GetOrderedRepoNames(repositories.Config{})
	if len(names) != 0 {
		t.Errorf("got %v, want empty slice", names)
	}
}
