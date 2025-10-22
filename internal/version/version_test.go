package version

import "testing"

func TestCurrent(t *testing.T) {
	origNumber, origCommit, origBuildDate := Number, Commit, BuildDate
	t.Cleanup(func() {
		Number, Commit, BuildDate = origNumber, origCommit, origBuildDate
	})

	Number = "1.2.3"
	Commit = "abc123"
	BuildDate = "2024-01-01T00:00:00Z"

	info := Current()

	if info.Version != Number {
		t.Fatalf("expected version %q, got %q", Number, info.Version)
	}
	if info.Commit != Commit {
		t.Fatalf("expected commit %q, got %q", Commit, info.Commit)
	}
	if info.BuildDate != BuildDate {
		t.Fatalf("expected build date %q, got %q", BuildDate, info.BuildDate)
	}
}

func TestCurrentOmitUnknown(t *testing.T) {
	origNumber, origCommit, origBuildDate := Number, Commit, BuildDate
	t.Cleanup(func() {
		Number, Commit, BuildDate = origNumber, origCommit, origBuildDate
	})

	Number = "0.0.1"
	Commit = "unknown"
	BuildDate = ""

	info := Current()

	if info.Commit != "" {
		t.Fatalf("expected commit to be omitted, got %q", info.Commit)
	}
	if info.BuildDate != "" {
		t.Fatalf("expected build date to be omitted, got %q", info.BuildDate)
	}
}

func TestSummary(t *testing.T) {
	origNumber, origCommit, origBuildDate := Number, Commit, BuildDate
	t.Cleanup(func() {
		Number, Commit, BuildDate = origNumber, origCommit, origBuildDate
	})

	Number = "2.0.0"
	Commit = "deadbeef"
	BuildDate = "2025-05-01T12:00:00Z"

	sum := Summary()
	if sum == "" {
		t.Fatal("expected non-empty summary")
	}

	if sum != "2.0.0 (deadbeef, 2025-05-01T12:00:00Z)" {
		t.Fatalf("unexpected summary: %s", sum)
	}

	Commit = "unknown"
	BuildDate = "unknown"

	sum = Summary()
	if sum != "2.0.0" {
		t.Fatalf("expected version-only summary, got %s", sum)
	}
}
