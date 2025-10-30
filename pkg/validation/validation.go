package validation

import internalvalidation "github.com/mergewayhq/mergeway-cli/internal/validation"

// Phase identifies a specific validation phase.
type Phase = internalvalidation.Phase

const (
	PhaseFormat     = internalvalidation.PhaseFormat
	PhaseSchema     = internalvalidation.PhaseSchema
	PhaseReferences = internalvalidation.PhaseReferences
)

// Options configures validation execution.
type Options = internalvalidation.Options

// Result aggregates validation errors.
type Result = internalvalidation.Result

// Error captures a single validation failure.
type Error = internalvalidation.Error

// Validate runs validation for the provided root and configuration.
var Validate = internalvalidation.Validate
