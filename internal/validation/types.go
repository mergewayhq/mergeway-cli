package validation

import "github.com/mergewayhq/mergeway-cli/internal/config"

// Phase identifies a specific validation phase.
type Phase string

const (
	PhaseFormat     Phase = "format"
	PhaseSchema     Phase = "schema"
	PhaseReferences Phase = "references"
)

// Options configures validation execution.
type Options struct {
	Phases   []Phase
	FailFast bool
}

// Result aggregates validation errors.
type Result struct {
	Errors []Error
}

// Error captures a single validation failure.
type Error struct {
	Phase   Phase
	Type    string
	ID      string
	File    string
	Message string
}

type rawObject struct {
	typeDef *config.TypeDefinition
	file    string
	index   int
	data    map[string]any
	id      string
}

type typeObjects struct {
	objects []*rawObject
}

type schemaIndex struct {
	byType map[string]map[string]*rawObject
}
