package version

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed VERSION
var embeddedNumber string

var (
	// Number is the semantic version for this build, sourced from the embedded VERSION file.
	Number = strings.TrimSpace(embeddedNumber)
	// Commit captures the git commit for the build when provided via -ldflags.
	Commit = "unknown"
	// BuildDate records when the binary was built (ISO-8601, UTC).
	BuildDate = "unknown"
)

// Info describes the current build metadata.
type Info struct {
	Version   string `json:"version" yaml:"version"`
	Commit    string `json:"commit,omitempty" yaml:"commit,omitempty"`
	BuildDate string `json:"buildDate,omitempty" yaml:"buildDate,omitempty"`
}

// Current returns the active build metadata, omitting unset fields.
func Current() Info {
	info := Info{Version: Number}
	if Commit != "" && Commit != "unknown" {
		info.Commit = Commit
	}
	if BuildDate != "" && BuildDate != "unknown" {
		info.BuildDate = BuildDate
	}
	return info
}

// Summary renders a human-friendly description of the current build.
func Summary() string {
	info := Current()
	if info.Commit == "" && info.BuildDate == "" {
		return info.Version
	}

	details := make([]string, 0, 2)
	if info.Commit != "" {
		details = append(details, info.Commit)
	}
	if info.BuildDate != "" {
		details = append(details, info.BuildDate)
	}

	return fmt.Sprintf("%s (%s)", info.Version, strings.Join(details, ", "))
}
